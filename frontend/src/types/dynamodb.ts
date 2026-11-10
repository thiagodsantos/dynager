export type AttributeValue =
  | { S: string }
  | { N: string }
  | { BOOL: boolean }
  | { NULL: true }
  | { M: Item }
  | { L: AttributeValue[] }
  | { SS: string[] }
  | { NS: string[] }
  | { BS: string[] }
  | { B: string }

export type Item = Record<string, AttributeValue>
export type KeyType = 'HASH' | 'RANGE'

export interface TableDescription {
  tableName: string
  partitionKey: { name: string; type: string }
  sortKey?: { name: string; type: string }
}

export interface TableSummary {
  name: string
}

export interface ItemPage {
  items: Item[]
  nextToken?: string
  count: number
  scannedCount: number
}

const attributeKeys = new Set(['S', 'N', 'BOOL', 'NULL', 'M', 'L', 'SS', 'NS', 'BS', 'B'])

export function isAttributeValue(value: unknown): value is AttributeValue {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false
  const entries = Object.entries(value)
  if (entries.length !== 1 || !attributeKeys.has(entries[0][0])) return false
  const [kind, payload] = entries[0]
  switch (kind) {
    case 'S':
    case 'N':
    case 'B': return typeof payload === 'string'
    case 'BOOL': return typeof payload === 'boolean'
    case 'NULL': return payload === true
    case 'M': return isItem(payload)
    case 'L': return Array.isArray(payload) && payload.every(isAttributeValue)
    case 'SS':
    case 'NS':
    case 'BS': return Array.isArray(payload) && payload.every((entry) => typeof entry === 'string')
    default: return false
  }
}

export function isItem(value: unknown): value is Item {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value) &&
    Object.values(value as Record<string, unknown>).every(isAttributeValue)
}

export function getItemKey(item: Item, metadata: TableDescription): Item | null {
  const names = [metadata.partitionKey.name, metadata.sortKey?.name].filter((name): name is string => Boolean(name))
  if (names.some((name) => !item[name])) return null
  return Object.fromEntries(names.map((name) => [name, item[name]]))
}

export function stringifyAttribute(value: AttributeValue): string {
  const [kind, payload] = Object.entries(value)[0]
  if (kind === 'S' || kind === 'N') return String(payload)
  if (kind === 'BOOL') return payload ? 'true' : 'false'
  if (kind === 'NULL') return 'null'
  return JSON.stringify(value)
}
