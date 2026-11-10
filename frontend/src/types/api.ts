import type { Item, ItemPage, TableDescription, TableSummary } from '@/types/dynamodb'

export interface ConnectionInput {
  name: string
  region: string
  profile?: string
  endpoint?: string
}

export interface Connection extends Omit<ConnectionInput, 'endpoint'> {
  id: string
  endpointConfigured?: boolean
}

export interface ConnectionResponse extends ConnectionInput {
  id: string
}

export interface TestConnectionResult {
  success: boolean
  message: string
}

export interface ConnectionsResponse { connections: ConnectionResponse[] }
export interface TablesResponse { tables: TableSummary[] }
export type { Item, ItemPage, TableDescription, TableSummary }
