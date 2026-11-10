import { useState, type FormEvent } from 'react'
import { toast } from 'react-toastify'
import { Code2, Plus, Trash2 } from 'lucide-react'
import { Button, ButtonVariant } from '@/components/atoms/Button'
import { JsonCodeEditor } from '@/components/atoms/JsonCodeEditor'
import { ErrorState } from '@/components/molecules/Feedback'
import { ItemDialogModal } from '@/components/organisms/ItemDialogModal'
import { AttributeType, getItemKey, isAttributeValue, isItem, type AttributeValue, type Item, type TableDescription } from '@/types/dynamodb'

export enum EntryMode {
  Builder = 'builder',
  Json = 'json',
}

type AttributeDraft = { id: string; name: string; kind: AttributeType; value: string; required?: boolean }

let nextAttributeId = 0

function createAttributeId() {
  nextAttributeId += 1
  return `attribute-${nextAttributeId}`
}
const kinds = Object.values(AttributeType)

function initialAttributes(metadata: TableDescription): AttributeDraft[] {
  return [
    { id: createAttributeId(), name: metadata.partitionKey.name, kind: metadata.partitionKey.type, value: '', required: true },
    ...(metadata.sortKey ? [{ id: createAttributeId(), name: metadata.sortKey.name, kind: metadata.sortKey.type, value: '', required: true }] : []),
    { id: createAttributeId(), name: '', kind: AttributeType.String, value: '' },
  ]
}

function parseDraftValue(kind: AttributeType, raw: string): AttributeValue {
  if (kind === AttributeType.String) return { S: raw }
  if (kind === AttributeType.Number) return { N: raw }
  if (kind === AttributeType.Binary) return { B: raw }
  if (kind === AttributeType.Boolean) return { BOOL: raw === 'true' }
  if (kind === AttributeType.Null) return { NULL: true }
  let value: unknown
  try { value = JSON.parse(raw) } catch { throw new Error(`${kind} values must be valid JSON.`) }
  if (kind === AttributeType.Map) {
    if (!isItem(value)) throw new Error('Map values must be a DynamoDB JSON object.')
    return { M: value }
  }
  if (kind === AttributeType.List) {
    if (!Array.isArray(value) || !value.every(isAttributeValue)) throw new Error('List values must be an array of DynamoDB attribute values.')
    return { L: value }
  }
  if (!Array.isArray(value) || !value.every((entry) => typeof entry === 'string')) throw new Error(`${kind} values must be a JSON array of strings.`)
  if (kind === AttributeType.StringSet) return { SS: value }
  if (kind === AttributeType.NumberSet) return { NS: value }
  return { BS: value }
}

function valuePlaceholder(kind: AttributeType) {
  if (kind === AttributeType.Boolean) return 'true or false'
  if (kind === AttributeType.Null) return 'No value required'
  if (kind === AttributeType.Map) return '{"field":{"S":"value"}}'
  if (kind === AttributeType.List) return '[{"S":"value"}]'
  if ([AttributeType.StringSet, AttributeType.NumberSet, AttributeType.BinarySet].includes(kind)) return '["value1","value2"]'
  if (kind === AttributeType.Number) return 'e.g. 42'
  if (kind === AttributeType.Binary) return 'Base64-encoded value'
  return 'Enter a value'
}

export function AddItemDialog({ metadata, busy, onClose, onCreate }: {
  metadata: TableDescription
  busy: boolean
  onClose: () => void
  onCreate: (item: Item) => Promise<string | null>
}) {
  const [mode, setMode] = useState<EntryMode>(EntryMode.Builder)
  const [attributes, setAttributes] = useState(() => initialAttributes(metadata))
  const [json, setJson] = useState('{}')
  const [error, setError] = useState('')

  function updateAttribute(index: number, changes: Partial<AttributeDraft>) {
    setAttributes((current) => current.map((attribute, currentIndex) => currentIndex === index ? { ...attribute, ...changes } : attribute))
  }

  function renderValueControl(attribute: AttributeDraft, index: number) {
    const label = `Attribute value ${index + 1}`
    const id = `attribute-value-${attribute.id}`
    if (attribute.kind === AttributeType.Boolean) {
      return <select id={id} className="input" aria-label={label} value={attribute.value} onChange={(event) => updateAttribute(index, { value: event.target.value })}><option value="">Choose…</option><option value="true">true</option><option value="false">false</option></select>
    }
    if (attribute.kind === AttributeType.Null) {
      return <input id={id} className="input" aria-label={label} value="null" disabled />
    }
    return <input id={id} className="input" aria-label={label} value={attribute.value} placeholder={valuePlaceholder(attribute.kind)} onChange={(event) => updateAttribute(index, { value: event.target.value })} />
  }

  function buildItem(): Item {
    const item: Item = {}
    attributes.forEach((attribute) => {
      if (!attribute.name.trim()) {
        if (attribute.value.trim()) throw new Error('Enter an attribute name for each value.')
        return
      }
      const name = attribute.name.trim()
      if (Object.prototype.hasOwnProperty.call(item, name)) throw new Error(`Attribute name "${name}" is duplicated.`)
      if (!attribute.value.trim() && attribute.kind !== AttributeType.Null) throw new Error(`Enter a value for "${name}".`)
      item[name] = parseDraftValue(attribute.kind, attribute.value)
    })
    return item
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    let item: unknown
    try {
      item = mode === EntryMode.Builder ? buildItem() : JSON.parse(json)
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
    const hasInvalidKey = expectedKeys.some((attribute) => {
      const value = item[attribute.name]
      if (!isAttributeValue(value) || Object.keys(value).length !== 1 || !Object.prototype.hasOwnProperty.call(value, attribute.type)) {
        setError(`Primary key "${attribute.name}" must use DynamoDB type ${attribute.type}.`)
        return true
      }
      if ((attribute.type === 'S' && !('S' in value && value.S.length > 0)) || (attribute.type === 'N' && !('N' in value && value.N.length > 0)) || (attribute.type === 'B' && !('B' in value && value.B.length > 0))) {
        setError(`Primary key "${attribute.name}" cannot be empty.`)
        return true
      }
      return false
    })
    if (hasInvalidKey) return
    const createError = await onCreate(item)
    if (createError) toast.error(createError)
  }

  return <ItemDialogModal icon={<Plus size={18} />} title="Add item" description={`Create a new record in ${metadata.tableName}.`} keyStatus="Primary key required" keyDescription={`${metadata.partitionKey.name}${metadata.sortKey ? ` + ${metadata.sortKey.name}` : ''}`} busy={busy} onClose={onClose} closeLabel="Close add item dialog" className="add-item-modal">
      <div className="item-entry-tabs" role="tablist" aria-label="Item entry method">
        <button type="button" role="tab" aria-selected={mode === EntryMode.Builder} className={mode === EntryMode.Builder ? 'item-entry-tab active' : 'item-entry-tab'} onClick={() => { setMode(EntryMode.Builder); setError('') }}>Select type and value</button>
        <button type="button" role="tab" aria-selected={mode === EntryMode.Json} className={mode === EntryMode.Json ? 'item-entry-tab active' : 'item-entry-tab'} onClick={() => { setMode(EntryMode.Json); setError('') }}><Code2 size={14} />Raw JSON</button>
      </div>
      <form onSubmit={submit}>
        {mode === EntryMode.Builder ? <div className="attribute-builder">
          {attributes.map((attribute, index) => <div className="attribute-builder-row" key={attribute.id}>
            <label className="field attribute-name-field" htmlFor={`attribute-name-${attribute.id}`}><span className="field-label">Attribute name{attribute.required ? ' · key' : ''}</span><input id={`attribute-name-${attribute.id}`} className="input" aria-label={`Attribute name ${index + 1}`} value={attribute.name} readOnly={attribute.required} placeholder="e.g. status" onChange={(event) => updateAttribute(index, { name: event.target.value })} /></label>
            <label className="field attribute-kind-field" htmlFor={`attribute-kind-${attribute.id}`}><span className="field-label">Type</span><select id={`attribute-kind-${attribute.id}`} className="input" aria-label={`Attribute type ${index + 1}`} value={attribute.kind} disabled={attribute.required} onChange={(event) => updateAttribute(index, { kind: event.target.value as AttributeType, value: '' })}>{kinds.map((kind) => <option key={kind} value={kind}>{kind}</option>)}</select></label>
            <label className="field attribute-value-field" htmlFor={`attribute-value-${attribute.id}`}><span className="field-label">Value</span>{renderValueControl(attribute, index)}</label>
            <button type="button" className="icon-button attribute-remove" aria-label={`Remove attribute ${index + 1}`} disabled={attribute.required} onClick={() => setAttributes((current) => current.filter((_, currentIndex) => currentIndex !== index))}><Trash2 size={15} /></button>
          </div>)}
          <Button type="button" onClick={() => setAttributes((current) => [...current, { id: createAttributeId(), name: '', kind: AttributeType.String, value: '' }])}><Plus size={14} />Add attribute</Button>
          <p className="field-hint editor-hint">For maps, lists, and sets, enter a JSON value using DynamoDB attribute shapes.</p>
        </div> : <>
          <div className="field"><span className="field-label">DynamoDB JSON</span><JsonCodeEditor id="new-item-json" value={json} onChange={setJson} describedBy="new-item-json-help" /></div>
          <span id="new-item-json-help" className="field-hint editor-hint">Include the table&apos;s required primary key fields and use DynamoDB JSON attribute shapes.</span>
        </>}
        {error && <ErrorState message={error} />}
        <div className="editor-actions"><Button type="button" onClick={onClose} disabled={busy}>Cancel</Button><Button type="submit" variant={ButtonVariant.Primary} busy={busy}>Add item</Button></div>
      </form>
  </ItemDialogModal>
}
