import { AlertCircle, CheckCircle2, Inbox, LoaderCircle } from 'lucide-react'
import type { ReactNode } from 'react'

export enum NoticeTone {
  Success = 'success',
  Error = 'error',
  Info = 'info',
}

export function LoadingState({ label = 'Loading…' }: { label?: string }) {
  return <div className="state-panel" role="status"><LoaderCircle className="spinner state-icon" /><span>{label}</span></div>
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return <div className="feedback feedback-error" role="alert"><AlertCircle size={18} /><span>{message}</span>{onRetry && <button type="button" className="text-button" onClick={onRetry}>Try again</button>}</div>
}

export function EmptyState({ title, description, action }: { title: string; description: string; action?: ReactNode }) {
  return <div className="empty-state"><span className="empty-icon"><Inbox size={22} /></span><h3>{title}</h3><p>{description}</p>{action}</div>
}

export function InlineNotice({ children, tone = NoticeTone.Success }: { children: ReactNode; tone?: NoticeTone }) {
  return <div className={`feedback feedback-${tone}`} role={tone === NoticeTone.Error ? 'alert' : 'status'}>{tone === NoticeTone.Success ? <CheckCircle2 size={17} /> : <AlertCircle size={17} />}{children}</div>
}
