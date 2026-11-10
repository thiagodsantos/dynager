import { useState, type FormEvent } from 'react'
import { toast } from 'react-toastify'
import { Code2, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/atoms/Button'
import { JsonCodeEditor } from '@/components/atoms/JsonCodeEditor'
import { ErrorState } from '@/components/molecules/Feedback'
import { ItemDialogModal } from '@/components/organisms/ItemDialogModal'
import { getItemKey, isAttributeValue, isItem, type AttributeValue, type Item, type TableDescription } from '@/types/dynamodb'

type AttributeKind = 'S' | 'N' | 'BOOL' | 'NULL' | 'M' | 'L' | 'SS' | 'NS' | 'BS' | 'B'
type AttributeDraft = { name: string; kind: AttributeKind; value: string; required?: boolean }
type EntryMode = 'builder' | 'json'

const kinds: AttributeKind[] = ['S', 'N', 'BOOL', 'NULL', 'M', 'L', 'SS', 'NS', 'BS', 'B']

function initialAttributes(metadata: TableDescription): AttributeDraft[] {
  return [
    { name: metadata.partitionKey.name, kind: metadata.partitionKey.type as AttributeKind, value: '', required: true },
    ...(metadata.sortKey ? [{ name: metadata.sortKey.name, kind: metadata.sortKey.type as AttributeKind, value: '', required: true }] : []),
    { name: '', kind: 'S', value: '' },
  ]
}

function parseDraftValue(kind: AttributeKind, raw: string): AttributeValue {
  if (kind === 'S') return { S: raw }
  if (kind === 'N') return { N: raw }
  if (kind === 'B') return { B: raw }
  if (kind === 'BOOL') return { BOOL: raw === 'true' }
  if (kind === 'NULL') return { NULL: true }
  let value: unknown
  try { value = JSON.parse(raw) } catch { throw new Error(`${kind} values must be valid JSON.`) }
  if (kind === 'M') {
    if (!isItem(value)) throw new Error('Map values must be a DynamoDB JSON object.')
    return { M: value }
  }
  if (kind === 'L') {
    if (!Array.isArray(value) || !value.every(isAttributeValue)) throw new Error('List values must be an array of DynamoDB attribute values.')
    return { L: value }
  }
  if (!Array.isArray(value) || !value.every((entry) => typeof entry === 'string')) throw new Error(`${kind} values must be a JSON array of strings.`)
  if (kind === 'SS') return { SS: value }
  if (kind === 'NS') return { NS: value }
  return { BS: value }
}

function valuePlaceholder(kind: AttributeKind) {
  if (kind === 'BOOL') return 'true or false'
  if (kind === 'NULL') return 'No value required'
  if (kind === 'M') return '{"field":{"S":"value"}}'
  if (kind === 'L') return '[{"S":"value"}]'
  if (kind === 'SS' || kind === 'NS' || kind === 'BS') return '["value1","value2"]'
  if (kind === 'N') return 'e.g. 42'
  if (kind === 'B') return 'Base64-encoded value'
  return 'Enter a value'
}

export function AddItemDialog({ metadata, busy, onClose, onCreate }: {
  metadata: TableDescription
  busy: boolean
  onClose: () => void
  onCreate: (item: Item) => Promise<string | null>
}) {
  const [mode, setMode] = useState<EntryMode>('builder')
  const [attributes, setAttributes] = useState(() => initialAttributes(metadata))
  const [json, setJson] = useState('{}')
  const [error, setError] = useState('')

  function updateAttribute(index: number, changes: Partial<AttributeDraft>) {
    setAttributes((current) => current.map((attribute, currentIndex) => currentIndex === index ? { ...attribute, ...changes } : attribute))
  }

  function buildItem(): Item {
    const item: Item = {}
    for (const attribute of attributes) {
      if (!attribute.name.trim()) {
        if (attribute.value.trim()) throw new Error('Enter an attribute name for each value.')
        continue
      }
      const name = attribute.name.trim()
      if (Object.prototype.hasOwnProperty.call(item, name)) throw new Error(`Attribute name "${name}" is duplicated.`)
      if (!attribute.value.trim() && attribute.kind !== 'NULL') throw new Error(`Enter a value for "${name}".`)
      item[name] = parseDraftValue(attribute.kind, attribute.value)
    }
    return item
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    let item: unknown
    try {
      item = mode === 'builder' ? buildItem() : JSON.parse(json)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Enter a valid DynamoDB item.')
      return
    }
    if (!isItem(item) || Object.keys(item).length === 0) {
      setError('Enter a non-empty item using valid DynamoDB attribute values.')
      return
    }
    const key = getItemKey(item, metadata)
    if (!key) {
      setError(`Include the required primary key attribute${metadata.sortKey ? 's' : ''}: ${metadata.partitionKey.name}${metadata.sortKey ? ` and ${metadata.sortKey.name}` : ''}.`)
      return
    }
    const expectedKeys = [metadata.partitionKey, metadata.sortKey].filter((entry): entry is NonNullable<typeof entry> => Boolean(entry))
    for (const attribute of expectedKeys) {
      if (!isAttributeValue(item[attribute.name]) || Object.keys(item[attribute.name]).length !== 1 || !Object.prototype.hasOwnProperty.call(item[attribute.name], attribute.type)) {
        setError(`Primary key "${attribute.name}" must use DynamoDB type ${attribute.type}.`)
        return
      }
      const value = item[attribute.name]
      if ((attribute.type === 'S' && !('S' in value && value.S.length > 0)) || (attribute.type === 'N' && !('N' in value && value.N.length > 0)) || (attribute.type === 'B' && !('B' in value && value.B.length > 0))) {
        setError(`Primary key "${attribute.name}" cannot be empty.`)
        return
      }
    }
    const createError = await onCreate(item)
    if (createError) toast.error(createError)
  }

  return <ItemDialogModal icon={<Plus size={18} />} title="Add item" description={`Create a new record in ${metadata.tableName}.`} keyStatus="Primary key required" keyDescription={`${metadata.partitionKey.name}${metadata.sortKey ? ` + ${metadata.sortKey.name}` : ''}`} busy={busy} onClose={onClose} closeLabel="Close add item dialog" className="add-item-modal">
      <div className="item-entry-tabs" role="tablist" aria-label="Item entry method">
        <button type="button" role="tab" aria-selected={mode === 'builder'} className={mode === 'builder' ? 'item-entry-tab active' : 'item-entry-tab'} onClick={() => { setMode('builder'); setError('') }}>Select type and value</button>
        <button type="button" role="tab" aria-selected={mode === 'json'} className={mode === 'json' ? 'item-entry-tab active' : 'item-entry-tab'} onClick={() => { setMode('json'); setError('') }}><Code2 size={14} />Raw JSON</button>
      </div>
      <form onSubmit={(event) => void submit(event)}>
        {mode === 'builder' ? <div className="attribute-builder">
          {attributes.map((attribute, index) => <div className="attribute-builder-row" key={index}>
            <label className="field attribute-name-field"><span className="field-label">Attribute name{attribute.required ? ' · key' : ''}</span><input className="input" aria-label={`Attribute name ${index + 1}`} value={attribute.name} readOnly={attribute.required} placeholder="e.g. status" onChange={(event) => updateAttribute(index, { name: event.target.value })} /></label>
            <label className="field attribute-kind-field"><span className="field-label">Type</span><select className="input" aria-label={`Attribute type ${index + 1}`} value={attribute.kind} disabled={attribute.required} onChange={(event) => updateAttribute(index, { kind: event.target.value as AttributeKind, value: '' })}>{kinds.map((kind) => <option key={kind} value={kind}>{kind}</option>)}</select></label>
            <label className="field attribute-value-field"><span className="field-label">Value</span>{attribute.kind === 'BOOL' ? <select className="input" aria-label={`Attribute value ${index + 1}`} value={attribute.value} onChange={(event) => updateAttribute(index, { value: event.target.value })}><option value="">Choose…</option><option value="true">true</option><option value="false">false</option></select> : attribute.kind === 'NULL' ? <input className="input" aria-label={`Attribute value ${index + 1}`} value="null" disabled /> : <input className="input" aria-label={`Attribute value ${index + 1}`} value={attribute.value} placeholder={valuePlaceholder(attribute.kind)} onChange={(event) => updateAttribute(index, { value: event.target.value })} />}</label>
            <button type="button" className="icon-button attribute-remove" aria-label={`Remove attribute ${index + 1}`} disabled={attribute.required} onClick={() => setAttributes((current) => current.filter((_, currentIndex) => currentIndex !== index))}><Trash2 size={15} /></button>
          </div>)}
          <Button type="button" onClick={() => setAttributes((current) => [...current, { name: '', kind: 'S', value: '' }])}><Plus size={14} />Add attribute</Button>
          <p className="field-hint editor-hint">For maps, lists, and sets, enter a JSON value using DynamoDB attribute shapes.</p>
        </div> : <>
          <label className="field" htmlFor="new-item-json"><span className="field-label">DynamoDB JSON</span><JsonCodeEditor id="new-item-json" value={json} onChange={setJson} describedBy="new-item-json-help" /></label>
          <span id="new-item-json-help" className="field-hint editor-hint">Include the table's required primary key fields and use DynamoDB JSON attribute shapes.</span>
        </>}
        {error && <ErrorState message={error} />}
        <div className="editor-actions"><Button type="button" onClick={onClose} disabled={busy}>Cancel</Button><Button type="submit" variant="primary" busy={busy}>Add item</Button></div>
      </form>
  </ItemDialogModal>
}
