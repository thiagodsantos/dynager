import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Check, Cloud, ShieldCheck, Wifi } from 'lucide-react'
import { toast } from 'react-toastify'
import { api } from '@/lib/api'
import type { ConnectionInput } from '@/types/api'
import { Button, ButtonVariant } from '@/components/atoms/Button'
import { Field } from '@/components/atoms/Field'
import { ErrorState } from '@/components/molecules/Feedback'
import { PageHeader } from '@/components/templates/PageHeader'

const initialForm: ConnectionInput = { name: '', region: '', profile: '', endpoint: '' }

export function ConnectionSetupPage() {
  const [form, setForm] = useState(initialForm)
  const [error, setError] = useState('')
  const [testing, setTesting] = useState(false)
  const [saving, setSaving] = useState(false)
  const navigate = useNavigate()

  function update(field: keyof ConnectionInput, value: string) {
    setForm((current) => ({ ...current, [field]: value }))
    setError('')
  }

  function validate(): ConnectionInput {
    const name = form.name.trim()
    const region = form.region.trim()
    if (!name || !region) throw new Error('Connection name and AWS region are required.')
    const endpoint = form.endpoint?.trim() ?? ''
    if (endpoint) {
      let parsed: URL
      try { parsed = new URL(endpoint) } catch { throw new Error('Enter a valid endpoint URL, including http:// or https://.') }
      if (parsed.username || parsed.password) throw new Error('Endpoint URLs cannot contain embedded usernames or passwords.')
      if (parsed.search || parsed.hash) throw new Error('Endpoint URLs cannot contain query strings or fragments.')
      if (!['http:', 'https:'].includes(parsed.protocol)) throw new Error('Endpoint must use HTTP or HTTPS.')
    }
    const profile = form.profile?.trim() ?? ''
    return { name, region, ...(profile ? { profile } : {}), ...(endpoint ? { endpoint } : {}) }
  }

  async function testConnection() {
    setError('')
    let input: ConnectionInput
    try { input = validate() } catch (reason) { setError(reason instanceof Error ? reason.message : 'Check the connection details.'); return }
    setTesting(true)
    try {
      const result = await api.testConnection(input)
      if (result.success) toast.success(result.message)
      else toast.error(result.message)
    } catch (reason) { toast.error(reason instanceof Error ? reason.message : 'Connection test failed.') }
    finally { setTesting(false) }
  }

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    let input: ConnectionInput
    try { input = validate() } catch (reason) { setError(reason instanceof Error ? reason.message : 'Check the connection details.'); return }
    setSaving(true)
    try {
      const connection = await api.createConnection(input)
      toast.success('Connection saved successfully.')
      navigate(`/connections/${encodeURIComponent(connection.id)}/tables`)
    } catch (reason) { toast.error(reason instanceof Error ? reason.message : 'Could not save connection.') }
    finally { setSaving(false) }
  }

  return <>
    <PageHeader eyebrow="Connection settings" title="Connect to DynamoDB" description="Choose a region and let the backend resolve credentials securely." actions={<Link className="button button-secondary button-md" to="/connections"><ArrowLeft size={15} />All connections</Link>} />
    <div className="setup-layout">
      <form className="panel setup-form" onSubmit={(event) => { void save(event) }}>
        <div className="panel-title-row"><div><span className="step-indicator">01</span><div><h2>Connection details</h2><p>Tell dynager where your DynamoDB service lives.</p></div></div><span className="form-required">* Required</span></div>
        <div className="form-grid">
          <Field label="Connection name" placeholder="e.g. Development" autoComplete="off" required value={form.name} onChange={(event) => update('name', event.target.value)} />
          <Field label="AWS region" placeholder="e.g. us-east-1" autoComplete="off" required value={form.region} onChange={(event) => update('region', event.target.value)} hint="The region where your DynamoDB tables are hosted." />
          <Field label="AWS profile" placeholder="Optional · default" autoComplete="off" value={form.profile} onChange={(event) => update('profile', event.target.value)} hint="Uses the AWS SDK credential chain when left blank." />
          <Field label="Custom endpoint" placeholder="http://127.0.0.1:8001" autoComplete="url" value={form.endpoint} onChange={(event) => update('endpoint', event.target.value)} hint="For the seeded dev database use port 8001. Port 8080 is the DynamoDB Manager API." />
        </div>
        <div className="security-callout"><ShieldCheck size={19} /><div><strong>Your AWS credentials are never entered here.</strong><p>dynager uses the backend’s AWS SDK credential chain. Access keys and session tokens are never sent to or stored by this browser.</p></div></div>
        {error && <ErrorState message={error} />}

        <div className="form-actions"><Button type="button" onClick={() => { void testConnection() }} busy={testing}><Wifi size={16} />Test connection</Button><Button type="submit" variant={ButtonVariant.Primary} busy={saving}><Check size={16} />Save connection</Button></div>
      </form>
      <aside className="setup-aside">
        <div className="info-card info-card-purple"><span className="info-icon"><Cloud size={19} /></span><p className="eyebrow">SECURE BY DEFAULT</p><h3>No secrets in the browser</h3><p>Credentials are resolved by your backend using its configured AWS credential provider. Connection settings contain only a name, region, optional profile, and endpoint.</p><div className="info-check"><ShieldCheck size={15} />No access key fields</div><div className="info-check"><ShieldCheck size={15} />No browser credential storage</div></div>
        <div className="setup-tip"><span className="tip-bulb">✦</span><p><strong>Local development?</strong><br />Run <code>make dev-up</code>, then connect to DynamoDB Local at <code>http://127.0.0.1:8001</code>.</p></div>
      </aside>
    </div>
  </>
}
