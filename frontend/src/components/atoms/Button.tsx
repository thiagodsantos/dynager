import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { LoaderCircle } from 'lucide-react'

export enum ButtonVariant {
  Primary = 'primary',
  Secondary = 'secondary',
  Ghost = 'ghost',
  Danger = 'danger',
}

export enum ButtonSize {
  Small = 'sm',
  Medium = 'md',
}
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  busy?: boolean
  children: ReactNode
}

export function Button({ type = 'button', variant = ButtonVariant.Secondary, size = ButtonSize.Medium, busy = false, className = '', children, disabled, ...props }: ButtonProps) {
  // The wrapper defaults the type but forwards submit/reset when explicitly requested.
  // eslint-disable-next-line react/button-has-type
  return <button type={type} className={`button button-${variant} button-${size} ${className}`} disabled={disabled || busy} {...props}>
    {busy && <LoaderCircle aria-hidden="true" className="spinner" size={15} />}{children}
  </button>
}
