import type { ReactNode } from 'react'
import { AlertTriangle, X } from 'lucide-react'
import { Button } from '@/components/atoms/Button'

export function ConfirmationDialog({ title, description, confirmLabel, busy, onCancel, onConfirm }: {
  title: string
  description: ReactNode
  confirmLabel: string
  busy: boolean
  onCancel: () => void
  onConfirm: () => void
}) {
  return <div className="modal-backdrop" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget && !busy) onCancel() }}>
    <section className="editor-modal confirmation-modal" role="alertdialog" aria-modal="true" aria-labelledby="confirmation-title" aria-describedby="confirmation-description">
      <div className="editor-header"><span className="confirmation-mark"><AlertTriangle size={18} /></span><div><h2 id="confirmation-title">{title}</h2><p id="confirmation-description">{description}</p></div><button type="button" className="icon-button" aria-label="Close confirmation" disabled={busy} onClick={onCancel}><X size={18} /></button></div>
      <div className="editor-actions"><Button type="button" onClick={onCancel} disabled={busy}>Cancel</Button><Button type="button" variant="danger" busy={busy} onClick={onConfirm}>{confirmLabel}</Button></div>
    </section>
  </div>
}
