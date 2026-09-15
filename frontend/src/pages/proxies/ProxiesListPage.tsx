import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Globe, Plus, Trash2, Activity, RefreshCw, CheckCircle, XCircle } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { ProxyProfile } from '../../types/api.ts'

interface TestStatus {
  testing: boolean
  ok?: boolean
  latencyMs?: number
  error?: string
}

export function ProxiesListPage() {
  const [proxies, setProxies] = useState<ProxyProfile[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [testStatuses, setTestStatuses] = useState<Record<string, TestStatus>>({})

  const fetchProxies = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.get<ProxyProfile[]>('/api/proxies')
      setProxies(data || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load proxies'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchProxies()
  }, [])

  const handleTestConnection = async (id: string) => {
    setTestStatuses((prev) => ({ ...prev, [id]: { testing: true } }))
    try {
      const res = await api.post<{ ok: boolean; latency_ms?: number; error?: string }>(
        `/api/proxies/${id}/test`
      )
      setTestStatuses((prev) => ({
        ...prev,
        [id]: { testing: false, ok: res.ok, latencyMs: res.latency_ms, error: res.error },
      }))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Connection test failed'
      setTestStatuses((prev) => ({
        ...prev,
        [id]: { testing: false, ok: false, error: msg },
      }))
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/api/proxies/${id}`)
      setProxies((prev) => prev.filter((p) => p.id !== id))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete proxy profile'
      setError(msg)
    }
  }

  const columns: Column<ProxyProfile>[] = [
    {
      key: 'name',
      title: 'Profile Name',
      render: (item) => (
        <div className="flex items-center gap-2">
          <Globe className="w-3.5 h-3.5 text-[var(--brand-text)]" />
          <span className="font-semibold text-[13px] text-[var(--text-primary)]">{item.name}</span>
        </div>
      ),
    },
    {
      key: 'scheme',
      title: 'Protocol',
      render: (item) => (
        <span className="font-mono uppercase text-[10.5px] font-semibold text-[var(--brand-text)] bg-[var(--brand-subtle)] px-2 py-0.5 rounded border border-[var(--border-subtle)]">
          {item.scheme}
        </span>
      ),
    },
    {
      key: 'host',
      title: 'Egress Host & Port',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--text-secondary)]">
          {item.host}:{item.port}
        </span>
      ),
    },
    {
      key: 'latency',
      title: 'Live Connectivity',
      render: (item) => {
        const test = testStatuses[item.id]
        if (test?.testing) {
          return (
            <span className="text-[11px] font-mono text-[var(--text-muted)] inline-flex items-center gap-1.5 animate-pulse">
              <span className="h-2 w-2 rounded-full bg-[var(--brand-primary)]" />
              Testing...
            </span>
          )
        }
        if (test?.ok) {
          return (
            <span className="text-[11px] font-mono text-[var(--status-success)] inline-flex items-center gap-1">
              <CheckCircle className="w-3.5 h-3.5" />
              <span>Online{test.latencyMs !== undefined ? ` (${test.latencyMs}ms)` : ''}</span>
            </span>
          )
        }
        if (test?.ok === false) {
          return (
            <span className="text-[11px] font-mono text-[var(--status-danger)] inline-flex items-center gap-1">
              <XCircle className="w-3.5 h-3.5" />
              <span>Failed</span>
            </span>
          )
        }
        return (
          <StatusBadge variant={item.enabled ? 'healthy' : 'disabled'}>
            {item.enabled ? 'Configured' : 'Inactive'}
          </StatusBadge>
        )
      },
    },
    {
      key: 'actions',
      title: 'Actions',
      width: '160px',
      render: (item) => {
        const isTestingThis = testStatuses[item.id]?.testing
        return (
          <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
            <Button
              variant="secondary"
              size="compact"
              isLoading={isTestingThis}
              onClick={() => handleTestConnection(item.id)}
              leftIcon={<Activity className="w-3 h-3" />}
            >
              Test
            </Button>
            <InlineConfirm
              trigger={
                <button
                  type="button"
                  title="Delete proxy"
                  className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              }
              confirmText="Delete?"
              onConfirm={() => handleDelete(item.id)}
            />
          </div>
        )
      },
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Outbound Proxies"
        description="Manage egress proxies (HTTP/SOCKS5/Relay) to route provider calls through isolated network tunnels."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Outbound Proxies' },
        ]}
        metadata={
          <span>
            {proxies.length} proxy profiles configured
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={fetchProxies}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Link to="/proxies/new">
              <Button variant="primary" size="compact" leftIcon={<Plus className="w-3.5 h-3.5" />}>
                Tambah
              </Button>
            </Link>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchProxies} />}

      <DataTable
        columns={columns}
        data={proxies}
        isLoading={isLoading}
        keyExtractor={(p) => p.id}
        emptyMessage="No outbound proxies configured. Click 'Add Proxy Profile' to register an egress proxy."
      />
    </div>
  )
}
