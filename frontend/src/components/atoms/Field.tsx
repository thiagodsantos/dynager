import type { InputHTMLAttributes, ReactNode, TextareaHTMLAttributes } from 'react'

interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  hint?: string
  trailing?: ReactNode
}

export function Field({ label, hint, id, trailing, ...props }: FieldProps) {
  const fieldId = id ?? `field-${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`
  return <label className="field" htmlFor={fieldId}>
    <span className="field-label">{label}</span>
    <span className="input-wrap"><input id={fieldId} className="input" {...props} />{trailing}</span>
    {hint && <span className="field-hint">{hint}</span>}
  </label>
}

interface TextAreaFieldProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string
  hint?: string
}

export function TextAreaField({ label, hint, id, ...props }: TextAreaFieldProps) {
  const fieldId = id ?? `field-${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`
  return <label className="field" htmlFor={fieldId}>
    <span className="field-label">{label}</span>
    <textarea id={fieldId} className="input textarea" {...props} />
    {hint && <span className="field-hint">{hint}</span>}
  </label>
}
