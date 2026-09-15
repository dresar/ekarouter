import { useEffect, useState, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Plus, Trash2, ExternalLink, RefreshCw, AlertCircle, Save, Flame, Cpu, Code, RotateCw, BarChart2, ArrowDownRight } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ProviderSelect } from '../../components/ui/ProviderSelect.tsx'
import { SearchableSelect } from '../../components/ui/SearchableSelect.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { CredentialPool, Provider } from '../../types/api.ts'

export function PoolsListPage() {
  const navigate = useNavigate()
  const [pools, setPools] = useState<CredentialPool[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showAddForm, setShowAddForm] = useState(false)
  const [newPool, setNewPool] = useState({
    name: '',
    provider_id: 'openai',
    environment: 'production',
    strategy: 'round_robin',
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const fetchPools = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [poolsRes, provRes] = await Promise.allSettled([
        api.get<CredentialPool[]>('/api/credential-pools'),
        api.get<Provider[]>('/api/providers'),
      ])
      if (poolsRes.status === 'fulfilled') setPools(poolsRes.value || [])
      if (provRes.status === 'fulfilled') setProviders(provRes.value || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load credential pools'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchPools()
  }, [])

  const handleCreatePool = async (e: FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    setFormError(null)

    try {
      await api.post('/api/credential-pools', {
        name: newPool.name.trim(),
        provider_id: newPool.provider_id.trim(),
        environment: newPool.environment,
        strategy: newPool.strategy,
      })
      setNewPool({
        name: '',
        provider_id: 'openai',
        environment: 'production',
        strategy: 'round_robin',
      })
      setShowAddForm(false)
      await fetchPools()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create pool'
      setFormError(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/api/credential-pools/${id}`)
      setPools((prev) => prev.filter((p) => p.id !== id))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete pool'
      setError(msg)
    }
  }

  const columns: Column<CredentialPool>[] = [
    {
      key: 'name',
      title: 'Pool Name',
      render: (item) => (
        <div className="flex items-center gap-2.5">
          <ProviderLogo providerId={item.provider_id} size="sm" />
          <div className="flex flex-col">
            <Link
              to={`/credential-pools/${item.id}`}
              className="font-semibold text-[13px] text-[var(--text-primary)] hover:text-[var(--brand-text)] hover:underline"
            >
              {item.name}
            </Link>
            <span className="text-[10.5px] font-mono text-[var(--text-muted)]">{item.provider_id}</span>
          </div>
        </div>
      ),
    },
    {
      key: 'environment',
      title: 'Environment',
      render: (item) => (
        <span className="capitalize font-mono text-[11px] text-[var(--text-secondary)]">
          {item.environment}
        </span>
      ),
    },
    {
      key: 'strategy',
      title: 'Rotation Strategy',
      render: (item) => (
        <span className="font-mono text-[11px] text-[var(--brand-text)]">
          {item.strategy || 'round_robin'}
        </span>
      ),
    },
    {
      key: 'status',
      title: 'Status',
      render: (item) => (
        <StatusBadge variant={item.status === 'active' ? 'healthy' : 'disabled'}>
          {item.status || 'Active'}
        </StatusBadge>
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
            onClick={() => navigate(`/credential-pools/${item.id}`)}
            title="Inspect pool members"
            className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
          >
            <ExternalLink className="w-4 h-4" />
          </button>
          <InlineConfirm
            trigger={
              <button
                type="button"
                className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] hover:bg-[var(--status-danger)]/10 transition-colors"
                title="Delete pool"
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
        title="HA Pools"
        description="Pool multiple credentials for automatic rotation, load balancing, and cooldown recovery."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'HA Pools' },
        ]}
        metadata={
          <span>
            {pools.length} active credential pools
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={fetchPools}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Button
              variant="primary"
              size="compact"
              onClick={() => setShowAddForm(!showAddForm)}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
            >
              {showAddForm ? 'Close Form' : 'New Pool'}
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchPools} />}

      {showAddForm && (
        <form
          onSubmit={handleCreatePool}
          className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3 animate-in fade-in duration-150"
        >
          <div className="flex items-center justify-between border-b border-[var(--border-subtle)] pb-2">
            <span className="text-[13px] font-semibold text-[var(--text-primary)]">
              Create Credential Pool
            </span>
          </div>

          {formError && (
            <div className="p-2.5 rounded bg-[var(--status-danger)]/10 text-[var(--status-danger)] text-[12px] flex items-center gap-1.5">
              <AlertCircle className="w-3.5 h-3.5" />
              <span>{formError}</span>
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-4 gap-3">
            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Pool Name *
              </label>
              <input
                type="text"
                required
                value={newPool.name}
                onChange={(e) => setNewPool({ ...newPool, name: e.target.value })}
                placeholder="Nama"
                className="w-full px-2.5 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
              />
            </div>

            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Provider *
              </label>
              <ProviderSelect
                value={newPool.provider_id}
                onChange={(val) => setNewPool({ ...newPool, provider_id: val })}
                providers={providers.map((p) => ({
                  id: p.id,
                  name: p.name,
                  kind: p.kind,
                }))}
              />
            </div>

            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Environment *
              </label>
              <SearchableSelect
                value={newPool.environment}
                onChange={(val) => setNewPool({ ...newPool, environment: val })}
                options={[
                  { value: 'production', label: 'Production', icon: <Flame className="w-3.5 h-3.5 text-red-400" /> },
                  { value: 'staging', label: 'Staging', icon: <Cpu className="w-3.5 h-3.5 text-amber-400" /> },
                  { value: 'development', label: 'Development', icon: <Code className="w-3.5 h-3.5 text-blue-400" /> },
                ]}
              />
            </div>

            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Rotation Strategy *
              </label>
              <SearchableSelect
                value={newPool.strategy}
                onChange={(val) => setNewPool({ ...newPool, strategy: val })}
                options={[
                  { value: 'round_robin', label: 'Round Robin', icon: <RotateCw className="w-3.5 h-3.5 text-blue-400" />, sublabel: 'Distribute evenly' },
                  { value: 'least_used', label: 'Least Used', icon: <BarChart2 className="w-3.5 h-3.5 text-purple-400" />, sublabel: 'Balance by load' },
                  { value: 'priority', label: 'Priority Order', icon: <ArrowDownRight className="w-3.5 h-3.5 text-emerald-400" />, sublabel: 'Tier priority' },
                ]}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-[var(--border-subtle)]">
            <Button
              type="button"
              variant="ghost"
              size="compact"
              onClick={() => setShowAddForm(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isSubmitting}
              leftIcon={<Save className="w-3.5 h-3.5" />}
            >
              Save Pool
            </Button>
          </div>
        </form>
      )}

      <DataTable
        columns={columns}
        data={pools}
        isLoading={isLoading}
        keyExtractor={(p) => p.id}
        emptyMessage="No high-availability pools configured. Click 'New Pool' to create one."
      />
    </div>
  )
}
