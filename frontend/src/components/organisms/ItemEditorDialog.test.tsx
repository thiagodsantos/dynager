import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ItemEditorDialog } from '@/components/organisms/ItemEditorDialog'
import { AttributeType, type Item, type TableDescription } from '@/types/dynamodb'

const metadata: TableDescription = {
  tableName: 'events',
  partitionKey: { name: 'pk', type: AttributeType.String },
  sortKey: { name: 'sk', type: AttributeType.Number },
}
const item: Item = { pk: { S: 'USER#1' }, sk: { N: '17' }, active: { BOOL: true } }

describe('item editor', () => {
  it('highlights JSON keys, strings, numbers, and literals while retaining an editable textarea', () => {
    render(<ItemEditorDialog item={item} metadata={metadata} busy={false} onClose={vi.fn()} onSave={vi.fn()} />)

    const editor = screen.getByLabelText('DynamoDB JSON')
    expect(editor).toHaveValue(JSON.stringify(item, null, 2))
    fireEvent.change(editor, { target: { value: JSON.stringify({ pk: 'USER#1', retries: 17, active: true }, null, 2) } })

    expect(document.querySelector('.json-token-key')).toHaveTextContent('pk')
    expect(document.querySelector('.json-token-string')).toHaveTextContent('USER#1')
    expect(document.querySelector('.json-token-number')).toHaveTextContent('17')
    expect(document.querySelector('.json-token-literal')).toHaveTextContent('true')
  })

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
