import { describe, expect, it } from 'vitest'
import { AttributeType, getItemKey, isAttributeValue, isItem } from '@/types/dynamodb'
import type { TableDescription } from '@/types/dynamodb'

describe('DynamoDB JSON validation', () => {
  it('accepts supported scalar, collection and nested attribute values', () => {
    expect(isAttributeValue({ S: 'hello' })).toBe(true)
    expect(isAttributeValue({ N: '42.0' })).toBe(true)
    expect(isAttributeValue({ BOOL: false })).toBe(true)
    expect(isAttributeValue({ NULL: true })).toBe(true)
    expect(isAttributeValue({ M: { nested: { L: [{ B: 'c29tZQ==' }, { SS: ['a', 'b'] }] } } })).toBe(true)
  })

  it('rejects malformed or ambiguous typed values', () => {
    expect(isAttributeValue({ S: 'x', N: '1' })).toBe(false)
    expect(isAttributeValue({ BOOL: 'true' })).toBe(false)
    expect(isAttributeValue({ L: [{ Unknown: 'x' }] })).toBe(false)
    expect(isItem({ value: 'plain string' })).toBe(false)
  })

  it('extracts complete partition and sort keys without changing DynamoDB types', () => {
    const metadata: TableDescription = { tableName: 'events', partitionKey: { name: 'pk', type: AttributeType.String }, sortKey: { name: 'sk', type: AttributeType.Number } }
    const key = getItemKey({ pk: { S: 'USER#1' }, sk: { N: '17' }, status: { BOOL: true } }, metadata)
    expect(key).toEqual({ pk: { S: 'USER#1' }, sk: { N: '17' } })
    expect(getItemKey({ pk: { S: 'USER#1' } }, metadata)).toBeNull()
  })
})
