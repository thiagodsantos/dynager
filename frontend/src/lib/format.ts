import type { AttributeValue } from '@/types/dynamodb'

export function getAttributeKind(value: AttributeValue): string {
  return Object.keys(value)[0]
}

export function getAttributePreview(value: AttributeValue): string {
  const [kind, payload] = Object.entries(value)[0]
  if (kind === 'M') return `Map · ${Object.keys(payload as object).length} fields`
  if (kind === 'L') return `List · ${(payload as unknown[]).length} values`
  if (Array.isArray(payload)) return `${kind} · ${payload.length} values`
  if (kind === 'NULL') return 'null'
  return String(payload)
}
