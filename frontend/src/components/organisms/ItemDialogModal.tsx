import type { MouseEventHandler, ReactNode } from 'react'
import { X } from 'lucide-react'

export function ItemDialogModal({ icon, title, description, keyStatus, keyDescription, busy, onClose, closeLabel, className = '', children }: {
  icon: ReactNode
  title: string
  description: string
  keyStatus: string
  keyDescription: string
  busy: boolean
  onClose: () => void
  closeLabel: string
  className?: string
  children: ReactNode
}) {
  const handleBackdropMouseDown: MouseEventHandler<HTMLDivElement> = (event) => {
    if (event.target === event.currentTarget && !busy) onClose()
  }

  return <div className="modal-backdrop" role="presentation" onMouseDown={handleBackdropMouseDown}>
    <section className={`editor-modal ${className}`} role="dialog" aria-modal="true" aria-labelledby="item-dialog-title">
      <div className="editor-header"><span className="editor-mark">{icon}</span><div><h2 id="item-dialog-title">{title}</h2><p>{description}</p></div><button className="icon-button" aria-label={closeLabel} disabled={busy} onClick={onClose}><X size={18} /></button></div>
      <div className="key-lock"><span className="lock-mark">⌑</span><span><strong>{keyStatus}</strong><small>{keyDescription}</small></span></div>
      {children}
    </section>
  </div>
}
