import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { LoaderCircle } from 'lucide-react'

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: 'sm' | 'md'
  busy?: boolean
  children: ReactNode
}

export function Button({ variant = 'secondary', size = 'md', busy = false, className = '', children, disabled, ...props }: ButtonProps) {
  return <button className={`button button-${variant} button-${size} ${className}`} disabled={disabled || busy} {...props}>
    {busy && <LoaderCircle aria-hidden="true" className="spinner" size={15} />}{children}
  </button>
}
