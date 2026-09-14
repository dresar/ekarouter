import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { GitFork, Plus, Trash2, Edit, RefreshCw, Shuffle, ArrowRight } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { EmptyState } from '../../components/ui/EmptyState.tsx'
import { api } from '../../api/client.ts'
import { Route } from '../../types/api.ts'

export function RoutingListPage() {
  const navigate = useNavigate()
  const [routes, setRoutes] = useState<Route[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchRoutes = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.get<Route[]>('/api/routes')
      setRoutes(data || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load routes'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchRoutes()
  }, [])

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/api/routes/${id}`)
      setRoutes((prev) => prev.filter((r) => r.id !== id))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete route'
      setError(msg)
    }
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Routing Combos"
        description="Configure ingress model aliases, priority-based failover chains, and load balancing policies."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Routing Combos' },
        ]}
        metadata={
          <span>
            {routes.length} routing combos active
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={fetchRoutes}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Link to="/routing/new">
              <Button variant="primary" size="compact" leftIcon={<Plus className="w-3.5 h-3.5" />}>
                New Route
              </Button>
            </Link>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchRoutes} />}

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {Array.from({ length: 6 }).map((_, idx) => (
            <div
              key={idx}
              className="h-36 rounded-[8px] bg-[var(--bg-card)] border border-[var(--border-subtle)] p-4 animate-pulse"
            />
          ))}
        </div>
      ) : routes.length === 0 ? (
        <EmptyState
          icon={<GitFork className="w-5 h-5" />}
          title="No Routing Combos Defined"
          description="Create a model routing combo to point model names like gpt-4o or claude-code to upstream provider accounts."
          action={
            <Link to="/routing/new">
              <Button variant="primary" size="compact" leftIcon={<Plus className="w-3.5 h-3.5" />}>
                Create First Route
              </Button>
            </Link>
          }
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {routes.map((route) => (
            <div
              key={route.id}
              className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3.5 flex flex-col justify-between transition-all hover:border-[var(--brand-primary)]/40 hover:bg-[var(--bg-panel)]/30"
            >
              <div>
                <div className="flex items-center justify-between gap-2 mb-2">
                  <span className="font-mono text-[13px] font-bold text-[var(--text-primary)] tracking-tight truncate">
                    {route.name}
                  </span>
                  <StatusBadge variant={route.enabled ? 'healthy' : 'disabled'}>
                    {route.enabled ? 'Enabled' : 'Disabled'}
                  </StatusBadge>
                </div>

                <div className="flex items-center gap-1.5 text-[11px] text-[var(--text-muted)] mb-3">
                  <Shuffle className="w-3 h-3 text-[var(--brand-text)]" />
                  <span className="capitalize font-medium text-[var(--text-secondary)]">
                    {route.strategy.replace('_', ' ')} Strategy
                  </span>
                  <span>&bull;</span>
                  <span>{route.items?.length || route.item_count || 0} targets</span>
                </div>

                {route.items && route.items.length > 0 && (
                  <div className="space-y-1.5 pt-2 border-t border-[var(--border-subtle)]">
                    <span className="text-[10px] font-semibold tracking-wider text-[var(--text-muted)] uppercase block">
                      Failover Sequence
                    </span>
                    <div className="space-y-1">
                      {route.items.slice(0, 3).map((item, idx) => (
                        <div
                          key={item.id || idx}
                          className="flex items-center justify-between text-[11px] bg-[var(--bg-panel)]/70 px-2 py-1 rounded-[5px] border border-[var(--border-subtle)]"
                        >
                          <div className="flex items-center gap-2 text-[var(--text-secondary)]">
                            <span className="font-mono text-[var(--brand-text)] font-semibold">
                              #{item.priority}
                            </span>
                            <ArrowRight className="w-2.5 h-2.5 text-[var(--text-muted)]" />
                            <ProviderLogo providerId={item.provider_id} size="sm" />
                            <span className="font-mono truncate max-w-[110px]">
                              {item.provider_id}
                            </span>
                          </div>
                          <span className="text-[10px] text-[var(--text-muted)] font-mono">
                            w:{item.weight}
                          </span>
                        </div>
                      ))}
                      {route.items.length > 3 && (
                        <span className="text-[10.5px] text-[var(--text-muted)] block text-center">
                          +{route.items.length - 3} more targets
                        </span>
                      )}
                    </div>
                  </div>
                )}
              </div>

              <div className="mt-3 pt-2.5 border-t border-[var(--border-subtle)] flex items-center justify-between">
                <button
                  type="button"
                  onClick={() => navigate(`/routing/${route.id}`)}
                  className="text-[11.5px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 font-medium"
                >
                  <Edit className="w-3.5 h-3.5" />
                  <span>Configure</span>
                </button>

                <InlineConfirm
                  trigger={
                    <button
                      type="button"
                      className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] hover:bg-[var(--status-danger)]/10 transition-colors"
                      title="Delete Route"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  }
                  confirmText="Delete?"
                  onConfirm={() => handleDelete(route.id)}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
