import { useCallback, useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowRight, Cloud, Database, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { api } from '@/lib/api'
import type { Connection } from '@/types/api'
import { Button, ButtonVariant } from '@/components/atoms/Button'
import { StatusBadge, StatusTone } from '@/components/atoms/StatusBadge'
import { ErrorState, EmptyState, LoadingState } from '@/components/molecules/Feedback'
import { ConfirmationDialog } from '@/components/organisms/ConfirmationDialog'
import { PageHeader } from '@/components/templates/PageHeader'

export function ConnectionsPage() {
  const [connections, setConnections] = useState<Connection[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [removing, setRemoving] = useState<string | null>(null)
  const [connectionToRemove, setConnectionToRemove] = useState<Connection | null>(null)
  const navigate = useNavigate()

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try { setConnections(await api.listConnections()) }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Could not load connections.') }
    finally { setLoading(false) }
  }, [])

  useEffect(() => { void load() }, [load])

  async function remove(connection: Connection) {
    setRemoving(connection.id)
    try {
      await api.deleteConnection(connection.id)
      setConnections((current) => current.filter((entry) => entry.id !== connection.id))
      setConnectionToRemove(null)
      toast.success(`Connection “${connection.name}” removed.`)
    } catch (reason) {
      toast.error(reason instanceof Error ? reason.message : 'Could not remove connection.')
    } finally { setRemoving(null) }
  }

  return <>
    <PageHeader eyebrow="Your workspace" title="Connections" description="Manage your DynamoDB access points and jump into a table." actions={<><Button onClick={() => { void load() }} aria-label="Refresh connections"><RefreshCw size={15} />Refresh</Button><Button variant={ButtonVariant.Primary} onClick={() => navigate('/connections/new')}><Plus size={16} />Add connection</Button></>} />
    <section className="metric-strip" aria-label="Workspace summary">
      <div className="metric-card"><span className="metric-icon icon-violet"><Database size={18} /></span><span className="metric-copy"><small>Connections</small><strong>{loading ? '—' : connections.length}</strong></span><span className="metric-note">configured</span></div>
      <div className="metric-card"><span className="metric-icon icon-green"><Cloud size={18} /></span><span className="metric-copy"><small>Credential source</small><strong>External</strong></span><span className="metric-note">never stored here</span></div>
      <div className="metric-card metric-wide"><span className="metric-icon icon-blue"><span className="pulse-dot" /></span><span className="metric-copy"><small>Connection mode</small><strong>Backend managed</strong></span><span className="metric-note">AWS SDK credential chain</span></div>
    </section>
    <div className="section-heading"><div><h2>Configured connections</h2><p>Connections contain only region and optional profile/endpoint settings.</p></div><StatusBadge>{connections.length} {connections.length === 1 ? 'connection' : 'connections'}</StatusBadge></div>
    {error && <ErrorState message={error} onRetry={() => { void load() }} />}
    {loading && <LoadingState label="Loading connections…" />}
        {!loading && connections.length === 0 && <EmptyState title="No connections yet" description="Add a region and let the backend use your AWS credential chain or named profile." action={<Button variant={ButtonVariant.Primary} onClick={() => navigate('/connections/new')}><Plus size={16} />Create a connection</Button>} />}
        {!loading && connections.length > 0 && <div className="connection-grid">
      {connections.map((connection) => <article className="connection-card" key={connection.id}>
        <div className="connection-card-top"><span className="connection-icon"><Database size={19} /></span><StatusBadge tone={StatusTone.Success}>Configured</StatusBadge><button type="button" className="icon-button danger-icon" aria-label={`Remove ${connection.name}`} disabled={removing === connection.id} onClick={() => setConnectionToRemove(connection)}><Trash2 size={16} /></button></div>
        <h3>{connection.name}</h3><p className="connection-region"><span className="region-mark">⌖</span>{connection.region}</p>
        <div className="connection-meta"><span>{connection.profile ? `Profile · ${connection.profile}` : 'AWS default credential chain'}</span>{connection.endpointConfigured && <span className="endpoint-tag">Custom endpoint configured</span>}</div>
        <Link className="connection-link" to={`/connections/${encodeURIComponent(connection.id)}/tables`}>Explore tables<ArrowRight size={16} /></Link>
      </article>)}
    </div>}
    {connectionToRemove && <ConfirmationDialog title="Remove connection?" description={`Remove “${connectionToRemove.name}” from this workspace? This only removes the saved connection settings.`} confirmLabel="Remove connection" busy={removing === connectionToRemove.id} onCancel={() => setConnectionToRemove(null)} onConfirm={() => { void remove(connectionToRemove) }} />}
  </>
}
