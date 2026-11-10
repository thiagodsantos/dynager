import { afterEach, describe, expect, it, vi } from 'vitest'
import { api } from '@/lib/api'

afterEach(() => vi.unstubAllGlobals())

describe('API client', () => {
  it('sends the API version header on calls and encodes filter JSON in the query', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ items: [], count: 0, scannedCount: 0 }), { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)
    await api.listItems('local id', 'user events', { limit: 25, filter: { attribute: 'status', operator: 'eq', value: { S: 'active' } } })
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect((init.headers as Headers).get('X-API-Version')).toBe('v1')
    expect(url).toContain('/api/connections/local%20id/tables/user%20events/items?')
    expect(new URL(url, 'http://localhost').searchParams.get('filter')).toBe(JSON.stringify({ attribute: 'status', operator: 'eq', value: { S: 'active' } }))
  })

  it('surfaces the message from the API error envelope', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { code: 'invalid_argument', message: 'Check the region.' } }), { status: 400 }))
    vi.stubGlobal('fetch', fetchMock)
    await expect(api.listConnections()).rejects.toThrow('Check the region.')
  })

  it('sends only the defined safe connection settings in the create payload', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ id: 'one', name: 'Dev', region: 'us-east-1', profile: 'dev', endpoint: 'http://user:password@localhost:8000', secretAccessKey: 'must-not-be-retained' }), { status: 201 }))
    vi.stubGlobal('fetch', fetchMock)
    const connection = await api.createConnection({ name: 'Dev', region: 'us-east-1', profile: 'dev' })
    expect(connection).toEqual({ id: 'one', name: 'Dev', region: 'us-east-1', profile: 'dev', endpointConfigured: true })
    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(JSON.parse(String(init.body))).toEqual({ name: 'Dev', region: 'us-east-1', profile: 'dev' })
    expect((init.headers as Headers).get('X-API-Version')).toBe('v1')
  })
})
