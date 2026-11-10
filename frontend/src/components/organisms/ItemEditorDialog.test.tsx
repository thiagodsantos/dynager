import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ItemEditorDialog } from '@/components/organisms/ItemEditorDialog'
import type { Item, TableDescription } from '@/types/dynamodb'

const metadata: TableDescription = {
  tableName: 'events',
  partitionKey: { name: 'pk', type: 'S' },
  sortKey: { name: 'sk', type: 'N' },
}
const item: Item = { pk: { S: 'USER#1' }, sk: { N: '17' }, active: { BOOL: true } }

describe('item editor', () => {
  it('preserves DynamoDB types and refuses primary-key changes', async () => {
    const onSave = vi.fn().mockResolvedValue(null)
    render(<ItemEditorDialog item={item} metadata={metadata} busy={false} onClose={vi.fn()} onSave={onSave} />)
    const editor = screen.getByLabelText('DynamoDB JSON')
    fireEvent.change(editor, { target: { value: JSON.stringify({ ...item, pk: { S: 'USER#2' } }) } })
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }))

    expect(await screen.findByText('Primary key attributes cannot be changed while editing.')).toBeInTheDocument()
    expect(onSave).not.toHaveBeenCalled()
  })

  it('submits the original full key and edited typed item', async () => {
    const onSave = vi.fn().mockResolvedValue(null)
    render(<ItemEditorDialog item={item} metadata={metadata} busy={false} onClose={vi.fn()} onSave={onSave} />)
    const updated = { ...item, active: { BOOL: false }, retries: { N: '3' } }
    fireEvent.change(screen.getByLabelText('DynamoDB JSON'), { target: { value: JSON.stringify(updated) } })
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }))

    expect(onSave).toHaveBeenCalledWith({ pk: { S: 'USER#1' }, sk: { N: '17' } }, updated)
  })
})
