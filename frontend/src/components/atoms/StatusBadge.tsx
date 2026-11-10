import type { ReactNode } from 'react'

export enum StatusTone {
  Neutral = 'neutral',
  Success = 'success',
  Warning = 'warning',
  Danger = 'danger',
}

export function StatusBadge({ children, tone = StatusTone.Neutral }: { children: ReactNode; tone?: StatusTone }) {
  return <span className={`status-badge status-${tone}`}><span className="status-dot" />{children}</span>
}
