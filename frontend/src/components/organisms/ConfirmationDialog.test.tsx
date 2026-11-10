import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ConfirmationDialog } from '@/components/organisms/ConfirmationDialog'

describe('confirmation dialog', () => {
  it('shows the confirmation message and invokes the confirm action', () => {
    const onConfirm = vi.fn()
    render(<ConfirmationDialog title="Delete item?" description="This action cannot be undone." confirmLabel="Delete item" busy={false} onCancel={vi.fn()} onConfirm={onConfirm} />)

    expect(screen.getByRole('alertdialog', { name: 'Delete item?' })).toBeInTheDocument()
    expect(screen.getByText('This action cannot be undone.')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Delete item' }))

    expect(onConfirm).toHaveBeenCalledOnce()
  })

  it('prevents dismissal while a confirmation action is busy', () => {
    const onCancel = vi.fn()
    render(<ConfirmationDialog title="Remove connection?" description="Confirm removal." confirmLabel="Remove" busy onCancel={onCancel} onConfirm={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    fireEvent.mouseDown(screen.getByRole('presentation'))

    expect(onCancel).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: 'Remove' })).toBeDisabled()
  })
})
