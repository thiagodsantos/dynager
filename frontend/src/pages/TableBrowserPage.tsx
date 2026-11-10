import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { toast } from 'react-toastify'
import { ArrowDownWideNarrow, ArrowLeft, ArrowRight, ChevronDown, Database, Filter, KeyRound, Pencil, Plus, RefreshCw, Search, Trash2, X } from 'lucide-react'
import { api } from '@/lib/api'
import { getAttributeKind, getAttributePreview } from '@/lib/format'
import { getItemKey, isAttributeValue, type AttributeValue, type Item, type ItemPage, type TableDescription } from '@/types/dynamodb'
import type { Connection, TableSummary } from '@/types/api'
import { Button } from '@/components/atoms/Button'
import { Field, TextAreaField } from '@/components/atoms/Field'
import { StatusBadge } from '@/components/atoms/StatusBadge'
import { EmptyState, ErrorState, LoadingState } from '@/components/molecules/Feedback'
import { ConfirmationDialog } from '@/components/organisms/ConfirmationDialog'
import { AddItemDialog } from '@/components/organisms/AddItemDialog'
import { ItemEditorDialog } from '@/components/organisms/ItemEditorDialog'
import { PageHeader } from '@/components/templates/PageHeader'

const PAGE_SIZE = 25
interface ItemFilter { attribute: string; operator: 'eq'; value: AttributeValue }

export function TableBrowserPage() {
  const { connectionId = '', tableName } = useParams()
  const navigate = useNavigate()
  const [connection, setConnection] = useState<Connection | null>(null)
  const [tables, setTables] = useState<TableSummary[]>([])
  const [tablesConnectionId, setTablesConnectionId] = useState('')
  const [tablesLoading, setTablesLoading] = useState(true)
  const [tablesError, setTablesError] = useState('')
  const [metadata, setMetadata] = useState<TableDescription | null>(null)
  const [metaLoading, setMetaLoading] = useState(false)
  const [metaError, setMetaError] = useState('')
  const [page, setPage] = useState<ItemPage | null>(null)
  const [pageLoading, setPageLoading] = useState(false)
  const [itemsError, setItemsError] = useState('')
  const [tokens, setTokens] = useState<Array<string | undefined>>([undefined])
  const [pageIndex, setPageIndex] = useState(0)
  const [filterName, setFilterName] = useState('')
  const [filterValue, setFilterValue] = useState('')
  const [activeFilter, setActiveFilter] = useState<ItemFilter | undefined>()
  const [selectedItem, setSelectedItem] = useState<Item | null>(null)
  const [addItemOpen, setAddItemOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Item | null>(null)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const [filterError, setFilterError] = useState('')


  useEffect(() => {
    let current = true
    setTablesLoading(true)
    setTablesError('')
    setTables([])
    setTablesConnectionId('')
    setConnection(null)
    Promise.all([api.listConnections(), api.listTables(connectionId)]).then(([connections, availableTables]) => {
      if (!current) return
      setConnection(connections.find((entry) => entry.id === connectionId) ?? null)
      setTables(availableTables)
      setTablesConnectionId(connectionId)
    }).catch((reason: unknown) => { if (current) setTablesError(reason instanceof Error ? reason.message : 'Could not load tables.') })
      .finally(() => { if (current) setTablesLoading(false) })
    return () => { current = false }
  }, [connectionId])

  useEffect(() => {
    if (!tablesLoading && tablesConnectionId === connectionId && !tableName && tables.length > 0) navigate(`/connections/${encodeURIComponent(connectionId)}/tables/${encodeURIComponent(tables[0].name)}`, { replace: true })
  }, [connectionId, navigate, tableName, tables, tablesConnectionId, tablesLoading])

  useEffect(() => {
    setActiveFilter(undefined)
    setFilterName('')
    setFilterValue('')
    setTokens([undefined])
    setPageIndex(0)
    setItemsError('')
    if (!tableName) { setMetadata(null); return undefined }
    let current = true
    setMetaLoading(true)
    setMetaError('')
    setMetadata(null)
    setPage(null)
    api.getTable(connectionId, tableName).then((data) => { if (current) setMetadata(data) })
      .catch((reason: unknown) => { if (current) setMetaError(reason instanceof Error ? reason.message : 'Could not load table metadata.') })
      .finally(() => { if (current) setMetaLoading(false) })
    return () => { current = false }
  }, [connectionId, tableName])

  const loadItems = useCallback(async () => {
    if (!tableName || !metadata || metadata.tableName !== tableName) return
    setPageLoading(true)
    setItemsError('')
    try {
      const result = await api.listItems(connectionId, tableName, { limit: PAGE_SIZE, nextToken: tokens[pageIndex], filter: activeFilter })
      setPage(result)
    } catch (reason) {
      setItemsError(reason instanceof Error ? reason.message : 'Could not load items.')
    } finally { setPageLoading(false) }
  }, [activeFilter, connectionId, metadata, pageIndex, tableName, tokens])

  useEffect(() => { loadItems() }, [loadItems])

  function applyFilter() {
    setFilterError('')
    if (!filterName.trim()) { setFilterError('Enter an attribute name to filter by.'); return }
    let value: unknown
    try { value = JSON.parse(filterValue) } catch { setFilterError('Filter value must be valid DynamoDB JSON, e.g. {"S":"active"}.'); return }
    if (!isAttributeValue(value)) { setFilterError('Filter value must be a single valid DynamoDB attribute value.'); return }
    setTokens([undefined])
    setPageIndex(0)
    setActiveFilter({ attribute: filterName.trim(), operator: 'eq', value })
  }

  function clearFilter() {
    setFilterName('')
    setFilterValue('')
    setFilterError('')
    setActiveFilter(undefined)
    setTokens([undefined])
    setPageIndex(0)
  }

  function goNext() {
    if (!page?.nextToken) return
    const nextIndex = pageIndex + 1
    setTokens((current) => [...current.slice(0, nextIndex), page.nextToken])
    setPageIndex(nextIndex)
  }

  function goPrevious() { if (pageIndex > 0) setPageIndex((current) => current - 1) }

  async function saveItem(key: Item, item: Item): Promise<string | null> {
    if (!tableName) return 'The selected table is no longer available.'
    setSaving(true)
    try {
      await api.updateItem(connectionId, tableName, key, item)
      setSelectedItem(null)
      toast.success('Item updated successfully.')
      await loadItems()
      return null
    } catch (reason) {
      return reason instanceof Error ? reason.message : 'Could not update item.'
    } finally { setSaving(false) }
  }

  async function createItem(item: Item): Promise<string | null> {
    if (!tableName || !metadata) return 'The selected table is no longer available.'
    const key = getItemKey(item, metadata)
    if (!key) return 'The item must include the complete primary key.'
    setSaving(true)
    setItemsError('')
    try {
      await api.updateItem(connectionId, tableName, key, item)
      setAddItemOpen(false)
      toast.success('Item added successfully.')
      const isAlreadyFirstPage = pageIndex === 0 && !tokens[0]
      setTokens([undefined])
      setPageIndex(0)
      if (isAlreadyFirstPage) await loadItems()
      return null
    } catch (reason) {
      return reason instanceof Error ? reason.message : 'Could not add item.'
    } finally { setSaving(false) }
  }

  async function removeItem(item: Item) {
    if (!tableName || !metadata) return
    const key = getItemKey(item, metadata)
    if (!key) { toast.error('This item is missing part of its primary key and cannot be deleted.'); setDeleteTarget(null); return }
    setDeleting(true)
    try {
      await api.deleteItem(connectionId, tableName, key)
      setDeleteTarget(null)
      toast.success('Item deleted successfully.')
      await loadItems()
    } catch (reason) { toast.error(reason instanceof Error ? reason.message : 'Could not delete item.') }
    finally { setDeleting(false) }
  }

  const columns = useMemo(() => [...new Set(page?.items.flatMap((item) => Object.keys(item)) ?? [])], [page])
  const connectionLabel = connection?.name ?? 'Connection'
  const pathToTables = `/connections/${encodeURIComponent(connectionId)}/tables`
  const emptyItemsTitle = activeFilter ? 'No items match this filter' : 'This table is empty'
  const emptyItemsDescription = activeFilter
    ? 'Try another typed value or clear the filter to see all scanned items.'
    : 'There are no records in this table yet.'
  const emptyItemsAction = activeFilter ? <Button onClick={clearFilter}>Clear filter</Button> : undefined

  return <>
    <div className="back-link-row"><Link className="back-link" to="/connections"><ArrowLeft size={15} />Connections</Link><span>/</span><span>{connectionLabel}</span></div>
    <PageHeader eyebrow="DynamoDB explorer" title={tableName ?? 'Tables'} description={connection ? `${connection.name} · ${connection.region}${connection.profile ? ` · profile ${connection.profile}` : ''}` : 'Browse tables and inspect their records.'} actions={<><Button onClick={() => { if (tableName) loadItems(); else navigate(pathToTables) }} aria-label="Refresh data"><RefreshCw size={15} />Refresh</Button>{tableName && metadata && <Button variant="primary" onClick={() => setAddItemOpen(true)}><Plus size={15} />Add item</Button>}</>} />
    {tablesError && <ErrorState message={tablesError} onRetry={() => navigate(pathToTables)} />}
    {tablesLoading ? <LoadingState label="Loading tables…" /> : <div className="browser-layout">
      <aside className="table-rail panel">
        <div className="rail-heading"><div><span className="eyebrow">DATABASE</span><h2>Tables</h2></div><StatusBadge>{tables.length}</StatusBadge></div>
        {tables.length === 0 ? <div className="rail-empty">No tables found for this connection.</div> : <nav className="table-nav" aria-label="DynamoDB tables">{tables.map((table) => <Link key={table.name} className={`table-nav-item ${tableName === table.name ? 'table-nav-active' : ''}`} to={`${pathToTables}/${encodeURIComponent(table.name)}`}><Database size={15} /><span>{table.name}</span><ChevronDown className="table-nav-chevron" size={13} /></Link>)}</nav>}
        <div className="rail-foot"><span className="secure-indicator" />Read access is backend-managed</div>
      </aside>
      <section className="browser-main">
        {!tableName ? <div className="panel table-welcome"><EmptyState title="Select a table" description="Choose a table from the list to inspect its key schema and browse records." /></div> : <>
          {(metaLoading || !metadata) && !metaError && <LoadingState label="Loading table schema…" />}
          {metaError && <ErrorState message={metaError} onRetry={() => navigate(`${pathToTables}/${encodeURIComponent(tableName)}`)} />}
          {metadata && <>
            <div className="schema-strip">
              <div className="schema-title"><span className="schema-icon"><Database size={17} /></span><span><small>TABLE</small><strong>{metadata.tableName}</strong></span></div>
              <div className="schema-key"><KeyRound size={15} /><span><small>Partition key</small><strong>{metadata.partitionKey.name}</strong></span><code>{metadata.partitionKey.type}</code></div>
              {metadata.sortKey && <div className="schema-key"><ArrowDownWideNarrow size={15} /><span><small>Sort key</small><strong>{metadata.sortKey.name}</strong></span><code>{metadata.sortKey.type}</code></div>}
              {!metadata.sortKey && <div className="schema-key schema-no-sort"><span><small>Key schema</small><strong>Partition key only</strong></span></div>}
            </div>
            <div className="panel filter-panel">
              <div className="filter-heading"><span className="filter-icon"><Filter size={16} /></span><div><h2>Filter items</h2><p>Equality filters scan the table. Broad scans may consume read capacity.</p></div><span className="scan-note">SCAN FILTER</span></div>
              <div className="filter-controls"><Field label="Attribute" aria-label="Filter attribute" placeholder="e.g. status" value={filterName} onChange={(event) => setFilterName(event.target.value)} /><TextAreaField label="Typed value (DynamoDB JSON)" aria-label="Typed filter value" rows={1} placeholder='{"S":"active"}' value={filterValue} onChange={(event) => setFilterValue(event.target.value)} /><div className="filter-buttons"><Button variant="primary" onClick={applyFilter}><Search size={15} />Apply</Button>{activeFilter && <Button onClick={clearFilter}><X size={15} />Clear</Button>}</div></div>
              {filterError && <ErrorState message={filterError} />}
              {activeFilter && <div className="active-filter"><span>Active: <strong>{activeFilter.attribute}</strong> equals <code>{JSON.stringify(activeFilter.value)}</code></span><button type="button" aria-label="Clear filter" className="text-button" onClick={clearFilter}>Clear</button></div>}
            </div>
            <section className="panel items-panel">
              <div className="items-toolbar"><div><div className="items-title"><h2>Items</h2>{page && <StatusBadge>{page.count} returned</StatusBadge>}</div><p>{page ? `${page.scannedCount.toLocaleString()} items scanned · page ${pageIndex + 1}` : 'Browse and edit typed DynamoDB records.'}</p></div><span className="page-size"><span className="page-size-dot" />25 per page</span></div>

              {itemsError && <ErrorState message={itemsError} onRetry={() => loadItems()} />}
              {pageLoading && <LoadingState label="Loading items…" />}
              {!pageLoading && page?.items.length === 0 && <EmptyState title={emptyItemsTitle} description={emptyItemsDescription} action={emptyItemsAction} />}
              {!pageLoading && page && page.items.length > 0 && <>
                <div className="items-table-wrap"><table className="items-table"><thead><tr><th className="row-number">#</th>{columns.map((column) => <th key={column}>{column}{column === metadata.partitionKey.name && <span className="column-key" title="Partition key">PK</span>}{column === metadata.sortKey?.name && <span className="column-key" title="Sort key">SK</span>}</th>)}<th className="actions-column">Actions</th></tr></thead><tbody>{page.items.map((item, index) => <tr key={JSON.stringify(getItemKey(item, metadata) ?? item)}><td className="row-number">{index + 1 + pageIndex * PAGE_SIZE}</td>{columns.map((column) => <td key={column}>{item[column] ? <span className="attribute-cell"><span className="attribute-kind">{getAttributeKind(item[column])}</span><span className="attribute-preview" title={JSON.stringify(item[column])}>{getAttributePreview(item[column])}</span></span> : <span className="missing-value">—</span>}</td>)}<td><div className="row-actions"><button type="button" className="icon-button row-edit" aria-label={`Edit item ${index + 1}`} onClick={() => setSelectedItem(item)}><Pencil size={15} /></button><button type="button" className="icon-button danger-icon" aria-label={`Delete item ${index + 1}`} disabled={deleting} onClick={() => setDeleteTarget(item)}><Trash2 size={15} /></button></div></td></tr>)}</tbody></table></div>
                <div className="pagination"><span>Showing <strong>{page.items.length ? pageIndex * PAGE_SIZE + 1 : 0}–{pageIndex * PAGE_SIZE + page.items.length}</strong> of scanned records</span><div><Button size="sm" onClick={goPrevious} disabled={pageIndex === 0 || pageLoading}><ArrowLeft size={14} />Previous</Button><span className="page-number">{pageIndex + 1}</span><Button size="sm" onClick={goNext} disabled={!page.nextToken || pageLoading}>Next<ArrowRight size={14} /></Button></div></div>
              </>}
            </section>
          </>}
        </>}
      </section>
    </div>}
    {deleteTarget && tableName && <ConfirmationDialog title="Delete item?" description={`Permanently delete this item from ${tableName}? This action cannot be undone.`} confirmLabel="Delete item" busy={deleting} onCancel={() => setDeleteTarget(null)} onConfirm={() => removeItem(deleteTarget)} />}
    {addItemOpen && metadata && <AddItemDialog metadata={metadata} busy={saving} onClose={() => setAddItemOpen(false)} onCreate={createItem} />}
    {selectedItem && metadata && <ItemEditorDialog item={selectedItem} metadata={metadata} busy={saving} onClose={() => setSelectedItem(null)} onSave={saveItem} />}
  </>
}
