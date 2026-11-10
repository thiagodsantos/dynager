import type { Connection, ConnectionInput, ConnectionResponse, ConnectionsResponse, Item, ItemPage, TablesResponse, TableDescription, TestConnectionResult } from '@/types/api'
import type { AttributeValue } from '@/types/dynamodb'

const API_VERSION = 'v1'
const API_BASE = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '')

export class ApiError extends Error {
  constructor(message: string, public readonly status?: number) {
    super(message)
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('X-API-Version', API_VERSION)
  if (init.body) headers.set('Content-Type', 'application/json')
  let response: Response
  try {
    response = await fetch(`${API_BASE}${path}`, { ...init, headers })
  } catch {
    throw new ApiError('Could not reach the API. Check that the service is running and try again.')
  }
  if (!response.ok) {
    let message = `Request failed (${response.status})`
    try {
      const body: unknown = await response.json()
      if (body && typeof body === 'object' && 'message' in body && typeof body.message === 'string') message = body.message
      else if (body && typeof body === 'object' && 'error' in body) {
        const error = body.error
        if (typeof error === 'string') message = error
        else if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string') message = error.message
      }
    } catch { /* Use the status message when the response is not JSON. */ }
    throw new ApiError(message, response.status)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

function sanitizeConnection(connection: ConnectionResponse): Connection {
  return {
    id: connection.id,
    name: connection.name,
    region: connection.region,
    ...(typeof connection.profile === 'string' ? { profile: connection.profile } : {}),
    ...(typeof connection.endpoint === 'string' && connection.endpoint.length > 0 ? { endpointConfigured: true } : {}),
  }
}

const segment = (value: string) => encodeURIComponent(value)
const connectionPath = (id: string) => `/api/connections/${segment(id)}`
const tablePath = (id: string, tableName: string) => `${connectionPath(id)}/tables/${segment(tableName)}`

export const api = {
  listConnections: async () => (await request<ConnectionsResponse>('/api/connections')).connections.map(sanitizeConnection),
  createConnection: async (input: ConnectionInput) => sanitizeConnection(await request<ConnectionResponse>('/api/connections', { method: 'POST', body: JSON.stringify(input) })),
  testConnection: (input: ConnectionInput) => request<TestConnectionResult>('/api/connections/test', { method: 'POST', body: JSON.stringify(input) }),
  deleteConnection: (id: string) => request<void>(connectionPath(id), { method: 'DELETE' }),
  listTables: async (id: string) => (await request<TablesResponse>(`${connectionPath(id)}/tables`)).tables,
  getTable: (id: string, tableName: string) => request<TableDescription>(tablePath(id, tableName)),
  listItems: (id: string, tableName: string, options: { limit: number; nextToken?: string; filter?: { attribute: string; operator: 'eq'; value: AttributeValue } }) => {
    const query = new URLSearchParams({ limit: String(options.limit) })
    if (options.nextToken) query.set('nextToken', options.nextToken)
    if (options.filter) query.set('filter', JSON.stringify(options.filter))
    return request<ItemPage>(`${tablePath(id, tableName)}/items?${query.toString()}`)
  },
  updateItem: (id: string, tableName: string, key: Item, item: Item) => request<void>(`${tablePath(id, tableName)}/items`, {
    method: 'PUT', body: JSON.stringify({ key, item }),
  }),
  deleteItem: (id: string, tableName: string, key: Item) => request<void>(`${tablePath(id, tableName)}/items`, {
    method: 'DELETE', body: JSON.stringify({ key }),
  }),
}
