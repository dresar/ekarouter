import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Plus, Trash2, ExternalLink, RefreshCw, Search } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { SecretViewer } from '../../components/ui/SecretViewer.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { VaultCredential } from '../../types/api.ts'
import { formatDate } from '../../utils/formatters.ts'

export function VaultListPage() {
  const navigate = useNavigate()
  const [credentials, setCredentials] = useState<VaultCredential[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [envFilter, setEnvFilter] = useState('all')

  const fetchCredentials = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.get<VaultCredential[]>('/api/v1/credentials')
      setCredentials(data || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load vault credentials'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchCredentials()
  }, [])

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/api/v1/credentials/${id}`)
      setCredentials((prev) => prev.filter((c) => c.id !== id))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete credential'
      setError(msg)
    }
  }

  const filtered = credentials.filter((c) => {
    const matchesSearch =
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.provider_id.toLowerCase().includes(search.toLowerCase()) ||
      (c.tags && c.tags.toLowerCase().includes(search.toLowerCase()))
    const matchesEnv = envFilter === 'all' || c.environment === envFilter
    return matchesSearch && matchesEnv
  })

  const columns: Column<VaultCredential>[] = [
    {
      key: 'name',
      title: 'Credential Name',
      render: (item) => (
        <div className="flex items-center gap-2.5">
          <ProviderLogo providerId={item.provider_id} size="sm" />
          <div className="flex flex-col">
            <Link
              to={`/vault/${item.id}`}
              className="font-semibold text-[12.5px] text-[var(--text-primary)] hover:text-[var(--brand-text)] hover:underline"
            >
              {item.name}
            </Link>
            <span className="text-[10.5px] font-mono text-[var(--text-muted)]">{item.provider_id}</span>
          </div>
        </div>
      ),
    },
    {
      key: 'masked_value',
      title: 'Masked Secret',
      render: (item) => <SecretViewer value={item.masked_value} />,
    },
    {
      key: 'environment',
      title: 'Environment',
      render: (item) => (
        <span
          className={`inline-flex px-2 py-0.5 rounded-[4px] text-[10.5px] font-mono font-medium border ${
            item.environment === 'production'
              ? 'bg-rose-500/10 text-rose-400 border-rose-500/20'
              : item.environment === 'staging'
              ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
              : 'bg-blue-500/10 text-blue-400 border-blue-500/20'
          }`}
        >
          {item.environment}
        </span>
      ),
    },
    {
      key: 'health_state',
      title: 'Health',
      render: (item) => (
        <StatusBadge
          variant={
            item.health_state === 'healthy'
              ? 'healthy'
              : item.health_state === 'degraded'
              ? 'degraded'
              : item.health_state === 'unhealthy'
              ? 'unhealthy'
              : 'neutral'
          }
        >
          {item.health_state}
        </StatusBadge>
      ),
    },
    {
      key: 'last_used_at',
      title: 'Last Used',
      render: (item) => (
        <span className="text-[11px] text-[var(--text-muted)] font-mono">
          {formatDate(item.last_used_at)}
        </span>
      ),
    },
    {
      key: 'actions',
      title: 'Actions',
      width: '90px',
      render: (item) => (
        <div className="flex items-center gap-1.5" onClick={(e) => e.stopPropagation()}>
          <button
            type="button"
            onClick={() => navigate(`/vault/${item.id}`)}
            title="Inspect credential"
            className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
          >
            <ExternalLink className="w-4 h-4" />
          </button>
          <InlineConfirm
            trigger={
              <button
                type="button"
                className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] hover:bg-[var(--status-danger)]/10 transition-colors"
                title="Delete credential"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            }
            confirmText="Delete?"
            onConfirm={() => handleDelete(item.id)}
          />
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Developer Credential Vault"
        description="Encrypted AES-256-GCM secret vault with health states, rate-limit backoff, and secret rotation."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Credential Vault' },
        ]}
        metadata={
          <span>
            {credentials.length} encrypted keys registered
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={fetchCredentials}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Link to="/vault/new">
              <Button variant="primary" size="compact" leftIcon={<Plus className="w-3.5 h-3.5" />}>
                Add Secret
              </Button>
            </Link>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchCredentials} />}

      <div className="flex flex-col sm:flex-row items-center gap-2.5 bg-[var(--bg-surface)] p-2.5 rounded-[8px] border border-[var(--border-subtle)]">
        <div className="relative flex-1 w-full">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)] pointer-events-none" />
          <input
            type="text"
            placeholder="Cari"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-8 pr-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)]"
          />
        </div>

        <select
          value={envFilter}
          onChange={(e) => setEnvFilter(e.target.value)}
          className="w-full sm:w-[160px] px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
        >
          <option value="all">All Environments</option>
          <option value="production">Production</option>
          <option value="staging">Staging</option>
          <option value="development">Development</option>
        </select>
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        isLoading={isLoading}
        keyExtractor={(c) => c.id}
        emptyMessage="No credentials found in vault"
      />
    </div>
  )
}
