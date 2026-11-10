import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { LoaderCircle } from 'lucide-react'

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: 'sm' | 'md'
  busy?: boolean
  children: ReactNode
}

export function Button({ type = 'button', variant = 'secondary', size = 'md', busy = false, className = '', children, disabled, ...props }: ButtonProps) {
  // The wrapper defaults the type but forwards submit/reset when explicitly requested.
  // eslint-disable-next-line react/button-has-type
  return <button type={type} className={`button button-${variant} button-${size} ${className}`} disabled={disabled || busy} {...props}>
    {busy && <LoaderCircle aria-hidden="true" className="spinner" size={15} />}{children}
  </button>
}
