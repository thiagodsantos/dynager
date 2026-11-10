import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { AddItemDialog } from '@/components/organisms/AddItemDialog'
import { AttributeType, type TableDescription } from '@/types/dynamodb'

const metadata: TableDescription = {
  tableName: 'events',
  partitionKey: { name: 'pk', type: AttributeType.String },
  sortKey: { name: 'sk', type: AttributeType.Number },
}

describe('add item dialog', () => {
  it('starts with the type-and-value builder and creates a typed item', async () => {
    const onCreate = vi.fn().mockResolvedValue(null)
    render(<AddItemDialog metadata={metadata} busy={false} onClose={vi.fn()} onCreate={onCreate} />)

    expect(screen.getByRole('tab', { name: 'Select type and value' })).toHaveAttribute('aria-selected', 'true')
    fireEvent.change(screen.getByLabelText('Attribute value 1'), { target: { value: 'USER#1' } })
    fireEvent.change(screen.getByLabelText('Attribute value 2'), { target: { value: '17' } })
    fireEvent.change(screen.getByLabelText('Attribute name 3'), { target: { value: 'active' } })
    fireEvent.change(screen.getByLabelText('Attribute type 3'), { target: { value: 'BOOL' } })
    fireEvent.change(screen.getByLabelText('Attribute value 3'), { target: { value: 'true' } })
    fireEvent.click(screen.getByRole('button', { name: 'Add item' }))

    await waitFor(() => expect(onCreate).toHaveBeenCalledWith({ pk: { S: 'USER#1' }, sk: { N: '17' }, active: { BOOL: true } }))
  })

  it('supports highlighted raw JSON and validates its primary key', async () => {
    const onCreate = vi.fn().mockResolvedValue(null)
    render(<AddItemDialog metadata={metadata} busy={false} onClose={vi.fn()} onCreate={onCreate} />)
    fireEvent.click(screen.getByRole('tab', { name: 'Raw JSON' }))
    const editor = screen.getByLabelText('DynamoDB JSON')
    fireEvent.change(editor, { target: { value: JSON.stringify({ pk: { S: 'USER#2' }, sk: { N: '18' }, count: { N: '2' } }) } })

    expect(document.querySelector('.json-token-key')).toHaveTextContent('pk')
    fireEvent.click(screen.getByRole('button', { name: 'Add item' }))
    await waitFor(() => expect(onCreate).toHaveBeenCalledWith({ pk: { S: 'USER#2' }, sk: { N: '18' }, count: { N: '2' } }))
  })

  it('requires the table primary key before creating an item', async () => {
    const onCreate = vi.fn().mockResolvedValue(null)
    render(<AddItemDialog metadata={metadata} busy={false} onClose={vi.fn()} onCreate={onCreate} />)
    fireEvent.click(screen.getByRole('tab', { name: 'Raw JSON' }))
    fireEvent.change(screen.getByLabelText('DynamoDB JSON'), { target: { value: JSON.stringify({ status: { S: 'active' } }) } })
    fireEvent.click(screen.getByRole('button', { name: 'Add item' }))

    expect(await screen.findByText(/Include the required primary key attributes/)).toBeInTheDocument()
    expect(onCreate).not.toHaveBeenCalled()
  })
})
