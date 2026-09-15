import { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  RefreshCw,
  Search,
  Clock,
  ExternalLink,
  ShieldCheck,
  AlertTriangle,
  Layers,
  Sparkles,
  ChevronDown,
  ChevronUp,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { AccountQuota } from '../../types/api.ts'

export function QuotaPage() {
  const navigate = useNavigate()
  const [quotas, setQuotas] = useState<AccountQuota[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [selectedProvider, setSelectedProvider] = useState('all')
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [expandedCards, setExpandedCards] = useState<Record<string, boolean>>({})

  const autoRefreshTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchQuotas = async (force = false) => {
    if (force) {
      setIsRefreshing(true)
    } else if (quotas.length === 0) {
      setIsLoading(true)
    }
    setError(null)
    try {
      const data = await api.get<AccountQuota[]>(`/api/quota${force ? '?force=1' : ''}`)
      setQuotas(Array.isArray(data) ? data : [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load quota limits'
      setError(msg)
    } finally {
      setIsLoading(false)
      setIsRefreshing(false)
    }
  }

  useEffect(() => {
    fetchQuotas()
  }, [])

  useEffect(() => {
    if (autoRefreshTimerRef.current) {
      clearInterval(autoRefreshTimerRef.current)
    }
    if (autoRefresh) {
      autoRefreshTimerRef.current = setInterval(() => {
        fetchQuotas(false)
      }, 30000)
    }
    return () => {
      if (autoRefreshTimerRef.current) {
        clearInterval(autoRefreshTimerRef.current)
      }
    }
  }, [autoRefresh])

  const toggleExpand = (accountId: string) => {
    setExpandedCards((prev) => ({
      ...prev,
      [accountId]: !prev[accountId],
    }))
  }

  const handleRefreshSingle = async (accountId: string) => {
    try {
      const updated = await api.post<AccountQuota>(`/api/quota/${accountId}/refresh`, {})
      if (updated) {
        setQuotas((prev) => prev.map((q) => (q.account_id === accountId ? updated : q)))
      }
    } catch {
      await fetchQuotas(true)
    }
  }

  const formatResetTime = (resetAt?: string) => {
    if (!resetAt) return null
    try {
      const target = new Date(resetAt).getTime()
      const now = Date.now()
      const diffMs = target - now
      if (diffMs <= 0) return 'Resetting soon'
      const diffMins = Math.floor(diffMs / 60000)
      const hours = Math.floor(diffMins / 60)
      const mins = diffMins % 60
      if (hours > 24) {
        const days = Math.floor(hours / 24)
        return `Resets in ${days}d ${hours % 24}h`
      }
      if (hours > 0) {
        return `Resets in ${hours}h ${mins}m`
      }
      return `Resets in ${mins}m`
    } catch {
      return null
    }
  }

  const getProgressColor = (pct: number) => {
    if (pct > 40) return 'bg-[var(--status-success)]'
    if (pct > 15) return 'bg-amber-400'
    return 'bg-[var(--status-error)]'
  }

  const getProgressBg = (pct: number) => {
    if (pct > 40) return 'text-[var(--status-success)]'
    if (pct > 15) return 'text-amber-400'
    return 'text-[var(--status-error)]'
  }

  const providersList = Array.from(new Set(quotas.map((q) => q.provider_id))).sort()

  const filteredQuotas = quotas.filter((q) => {
    const matchesProv = selectedProvider === 'all' || q.provider_id === selectedProvider
    if (!matchesProv) return false
    if (!search.trim()) return true
    const term = search.toLowerCase()
    return (
      q.account_name.toLowerCase().includes(term) ||
      q.provider_id.toLowerCase().includes(term) ||
      (q.email && q.email.toLowerCase().includes(term)) ||
      q.quotas.some((m) => m.name.toLowerCase().includes(term))
    )
  })

  const totalTracked = quotas.length
  const healthyCount = quotas.filter((q) => q.overall_remaining > 20).length
  const depletedCount = quotas.filter((q) => q.overall_remaining <= 5).length
  const avgRemaining =
    totalTracked > 0
      ? Math.round(quotas.reduce((acc, q) => acc + q.overall_remaining, 0) / totalTracked)
      : 100

  return (
    <div className="space-y-6">
      <PageHeader
        title="Quota Tracker"
        description="Track and manage your API quota limits"
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Quota Tracker' },
        ]}
        metadata={
          <span>
            {healthyCount} healthy &bull; {depletedCount} depleted &bull; avg {avgRemaining}% remaining
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setAutoRefresh(!autoRefresh)}
              className={`h-8 px-3 text-[11.5px] font-medium rounded-[6px] border flex items-center gap-1.5 transition-colors cursor-pointer ${
                autoRefresh
                  ? 'bg-[var(--status-success-bg)] border-[var(--status-success)] text-[var(--status-success)]'
                  : 'bg-[var(--bg-panel)] border-[var(--border-subtle)] text-[var(--text-muted)]'
              }`}
            >
              <span
                className={`w-1.5 h-1.5 rounded-full ${
                  autoRefresh ? 'bg-[var(--status-success)] animate-pulse' : 'bg-slate-500'
                }`}
              />
              Auto-refresh 30s
            </button>
            <Button
              variant="secondary"
              size="compact"
              onClick={() => fetchQuotas(true)}
              isLoading={isRefreshing || isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh All
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={() => fetchQuotas(true)} />}

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div className="p-3.5 rounded-[10px] bg-[var(--bg-card)] border border-[var(--border-subtle)] shadow-xs flex flex-col">
          <span className="text-[11px] font-medium text-[var(--text-muted)] uppercase tracking-wide">
            Tracked Accounts
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-[20px] font-bold text-[var(--text-primary)] font-mono">
              {totalTracked}
            </span>
            <span className="text-[11px] text-[var(--text-muted)]">connected</span>
          </div>
        </div>

        <div className="p-3.5 rounded-[10px] bg-[var(--bg-card)] border border-[var(--border-subtle)] shadow-xs flex flex-col">
          <span className="text-[11px] font-medium text-[var(--status-success)] uppercase tracking-wide flex items-center gap-1">
            <ShieldCheck className="w-3 h-3" />
            Healthy Quota
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-[20px] font-bold text-[var(--status-success)] font-mono">
              {healthyCount}
            </span>
            <span className="text-[11px] text-[var(--text-muted)]">&gt; 20% left</span>
          </div>
        </div>

        <div className="p-3.5 rounded-[10px] bg-[var(--bg-card)] border border-[var(--border-subtle)] shadow-xs flex flex-col">
          <span className="text-[11px] font-medium text-[var(--status-error)] uppercase tracking-wide flex items-center gap-1">
            <AlertTriangle className="w-3 h-3" />
            Depleted / Low
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-[20px] font-bold text-[var(--status-error)] font-mono">
              {depletedCount}
            </span>
            <span className="text-[11px] text-[var(--text-muted)]">&le; 5% left</span>
          </div>
        </div>

        <div className="p-3.5 rounded-[10px] bg-[var(--bg-card)] border border-[var(--border-subtle)] shadow-xs flex flex-col">
          <span className="text-[11px] font-medium text-[var(--text-muted)] uppercase tracking-wide">
            Avg Headroom
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className={`text-[20px] font-bold font-mono ${getProgressBg(avgRemaining)}`}>
              {avgRemaining}%
            </span>
            <span className="text-[11px] text-[var(--text-muted)]">remaining</span>
          </div>
        </div>
      </div>

      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <div className="relative flex-1 min-w-0 max-w-xs">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-[var(--text-muted)]" />
          <input
            type="search"
            placeholder="Cari akun atau model..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full h-8 pl-8 pr-3 text-[12.5px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
          />
        </div>

        <div className="flex items-center gap-1.5 flex-wrap">
          <button
            type="button"
            onClick={() => setSelectedProvider('all')}
            className={`h-7 px-3 text-[11.5px] font-medium rounded-[5px] transition-colors cursor-pointer ${
              selectedProvider === 'all'
                ? 'bg-[#282a34] text-[#f3f4f6] border border-[#3e4354]'
                : 'bg-[var(--bg-panel)] text-[var(--text-secondary)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] hover:text-[var(--text-primary)]'
            }`}
          >
            All
          </button>
          {providersList.map((pId) => (
            <button
              key={pId}
              type="button"
              onClick={() => setSelectedProvider(pId)}
              className={`h-7 px-3 text-[11.5px] font-medium rounded-[5px] transition-colors cursor-pointer ${
                selectedProvider === pId
                  ? 'bg-[#282a34] text-[#f3f4f6] border border-[#3e4354]'
                  : 'bg-[var(--bg-panel)] text-[var(--text-secondary)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] hover:text-[var(--text-primary)]'
              }`}
            >
              {pId}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {filteredQuotas.map((item) => {
          const resetText = formatResetTime(item.reset_at)
          const isExpanded = expandedCards[item.account_id] ?? (item.quotas.length > 0 && item.quotas.length <= 4)
          const roundedPct = Math.round(item.overall_remaining)

          return (
            <div
              key={item.account_id}
              className="p-4 rounded-[12px] bg-[var(--bg-card)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] transition-all flex flex-col justify-between shadow-xs"
            >
              <div>
                <div className="flex items-start justify-between gap-3">
                  <div className="flex items-center gap-3 min-w-0">
                    <ProviderLogo providerId={item.provider_id} name={item.provider_name} size="md" />
                    <div className="flex flex-col min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="text-[14px] font-semibold text-[var(--text-primary)] truncate">
                          {item.account_name}
                        </span>
                        {item.plan && (
                          <span className="px-1.5 py-0.5 text-[10px] font-mono rounded-[4px] bg-[var(--brand-primary)]/10 text-[var(--brand-text)] border border-[var(--brand-primary)]/20">
                            {item.plan}
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-2 mt-0.5 text-[11px] text-[var(--text-muted)]">
                        <span className="font-mono">{item.provider_name || item.provider_id}</span>
                        {item.email && (
                          <>
                            <span>&bull;</span>
                            <span className="truncate">{item.email}</span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-1.5 shrink-0">
                    <button
                      type="button"
                      onClick={() => handleRefreshSingle(item.account_id)}
                      title="Refresh single quota"
                      className="p-1.5 rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors cursor-pointer"
                    >
                      <RefreshCw className="w-3.5 h-3.5" />
                    </button>
                    <button
                      type="button"
                      onClick={() => navigate(`/providers/${item.provider_id}`)}
                      title="View provider"
                      className="p-1.5 rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors cursor-pointer"
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>

                <div className="mt-4 space-y-1.5">
                  <div className="flex items-center justify-between text-[12px]">
                    <span className="font-medium text-[var(--text-secondary)]">Remaining Quota</span>
                    <span className={`font-mono font-bold ${getProgressBg(roundedPct)}`}>
                      {roundedPct}%
                    </span>
                  </div>
                  <div className="w-full h-2 rounded-full bg-[var(--bg-panel)] overflow-hidden">
                    <div
                      className={`h-full transition-all duration-300 ${getProgressColor(roundedPct)}`}
                      style={{ width: `${Math.max(2, Math.min(100, roundedPct))}%` }}
                    />
                  </div>
                </div>

                {resetText && (
                  <div className="mt-2.5 flex items-center gap-1.5 text-[11px] font-mono text-[var(--text-muted)]">
                    <Clock className="w-3 h-3 text-[var(--brand-text)]" />
                    <span>{resetText}</span>
                  </div>
                )}

                {item.quotas && item.quotas.length > 0 && (
                  <div className="mt-4 pt-3 border-t border-[var(--border-subtle)]">
                    <button
                      type="button"
                      onClick={() => toggleExpand(item.account_id)}
                      className="w-full flex items-center justify-between text-[11.5px] font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors cursor-pointer py-1"
                    >
                      <span className="flex items-center gap-1.5">
                        <Layers className="w-3.5 h-3.5" />
                        Model Limit Breakdown ({item.quotas.length})
                      </span>
                      {isExpanded ? (
                        <ChevronUp className="w-3.5 h-3.5" />
                      ) : (
                        <ChevronDown className="w-3.5 h-3.5" />
                      )}
                    </button>

                    {isExpanded && (
                      <div className="mt-2 space-y-2 pt-1">
                        {item.quotas.map((mq) => {
                          const mPct = Math.round(mq.remaining_percentage)
                          const mReset = formatResetTime(mq.reset_at)

                          return (
                            <div
                              key={mq.id}
                              className="p-2 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[11.5px] space-y-1.5"
                            >
                              <div className="flex items-center justify-between">
                                <span className="font-medium text-[var(--text-primary)] truncate max-w-[200px]">
                                  {mq.display_name || mq.name}
                                </span>
                                <div className="flex items-center gap-2">
                                  {mReset && (
                                    <span className="text-[10px] font-mono text-[var(--text-muted)]">
                                      {mReset}
                                    </span>
                                  )}
                                  <span className={`font-mono font-semibold ${getProgressBg(mPct)}`}>
                                    {mPct}%
                                  </span>
                                </div>
                              </div>
                              <div className="w-full h-1.5 rounded-full bg-[var(--bg-card)] overflow-hidden">
                                <div
                                  className={`h-full transition-all duration-300 ${getProgressColor(mPct)}`}
                                  style={{ width: `${Math.max(1, Math.min(100, mPct))}%` }}
                                />
                              </div>
                            </div>
                          )
                        })}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>

      {filteredQuotas.length === 0 && !isLoading && (
        <div className="py-16 text-center rounded-[8px] bg-[var(--bg-card)] border border-[var(--border-subtle)]">
          <Sparkles className="w-8 h-8 mx-auto text-[var(--brand-text)] opacity-80" />
          <p className="text-[13.5px] font-semibold text-[var(--text-primary)] mt-3">
            No quota limits recorded yet
          </p>
          <p className="text-[12px] text-[var(--text-muted)] mt-1 max-w-md mx-auto">
            Connect Antigravity, Gemini CLI, or AI providers to automatically track real-time quota
            limits, 5h reset windows, and rate headroom.
          </p>
          <Button
            variant="primary"
            size="compact"
            onClick={() => navigate('/providers')}
            className="mt-4"
          >
            Go to Providers
          </Button>
        </div>
      )}
    </div>
  )
}
