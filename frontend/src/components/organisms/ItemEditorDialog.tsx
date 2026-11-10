import { useEffect, useState, type FormEvent } from 'react'
import { Code2, X } from 'lucide-react'
import { Button } from '@/components/atoms/Button'
import { ErrorState } from '@/components/molecules/Feedback'
import { getItemKey, isItem, type Item, type TableDescription } from '@/types/dynamodb'

export function ItemEditorDialog({ item, metadata, busy, onClose, onSave }: {
  item: Item
  metadata: TableDescription
  busy: boolean
  onClose: () => void
  onSave: (key: Item, nextItem: Item) => Promise<string | null>
}) {
  const [json, setJson] = useState('')
  const [error, setError] = useState('')
  useEffect(() => { setJson(JSON.stringify(item, null, 2)); setError('') }, [item])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    let edited: unknown
    try { edited = JSON.parse(json) } catch { setError('Enter valid JSON before saving.'); return }
    if (!isItem(edited)) { setError('Every item field must use a valid DynamoDB attribute value, such as {"S":"text"} or {"N":"42"}.'); return }
    const originalKey = getItemKey(item, metadata)
    const editedKey = getItemKey(edited, metadata)
    if (!originalKey || !editedKey) { setError('The complete primary key must remain present in the item.'); return }
    if (JSON.stringify(originalKey) !== JSON.stringify(editedKey)) { setError('Primary key attributes cannot be changed while editing.'); return }
    const saveError = await onSave(originalKey, edited)
    if (saveError) setError(saveError)
  }

  return <div className="modal-backdrop" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget && !busy) onClose() }}>
    <section className="editor-modal" role="dialog" aria-modal="true" aria-labelledby="editor-title">
      <div className="editor-header"><span className="editor-mark"><Code2 size={18} /></span><div><h2 id="editor-title">Edit item</h2><p>Preserve DynamoDB types and the full primary key.</p></div><button className="icon-button" aria-label="Close editor" disabled={busy} onClick={onClose}><X size={18} /></button></div>
      <div className="key-lock"><span className="lock-mark">⌑</span><span><strong>Primary key locked</strong><small>{metadata.partitionKey.name}{metadata.sortKey ? ` + ${metadata.sortKey.name}` : ''}</small></span></div>
      <form onSubmit={(event) => void submit(event)}>
        <label className="field" htmlFor="item-json"><span className="field-label">DynamoDB JSON</span><textarea id="item-json" className="input code-textarea" spellCheck={false} value={json} onChange={(event) => setJson(event.target.value)} aria-describedby="item-json-help" /></label>
        <span id="item-json-help" className="field-hint editor-hint">Use DynamoDB JSON attribute shapes: S, N, BOOL, NULL, M, L, SS, NS, BS, or B.</span>
        {error && <ErrorState message={error} />}
        <div className="editor-actions"><Button type="button" onClick={onClose} disabled={busy}>Cancel</Button><Button type="submit" variant="primary" busy={busy}>Save changes</Button></div>
      </form>
    </section>
  </div>
}
