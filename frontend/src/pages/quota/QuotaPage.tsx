import { useEffect, useState, useMemo } from 'react'
import {
  RefreshCw,
  Hourglass,
  Ban,
  CheckCircle2,
  ToggleLeft,
  ToggleRight,
  EyeOff,
  Pencil,
  Trash2,
  AlertCircle,
  ChevronDown,
  LayoutGrid,
  Check,
  X,
} from 'lucide-react'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { AccountQuota } from '../../types/api.ts'

function formatResetCountdown(resetAt?: string): string {
  if (!resetAt) return '-'
  try {
    const target = new Date(resetAt).getTime()
    const now = Date.now()
    const diffMs = target - now
    if (diffMs <= 0) return 'in 0m'
    const totalMinutes = Math.floor(diffMs / 60000)
    const days = Math.floor(totalMinutes / (24 * 60))
    const hours = Math.floor((totalMinutes % (24 * 60)) / 60)
    const minutes = totalMinutes % 60
    if (days > 0) {
      return `in ${days}d ${hours}h ${minutes}m`
    }
    if (hours > 0) {
      return `in ${hours}h ${minutes}m`
    }
    return `in ${minutes}m`
  } catch {
    return '-'
  }
}

function getQuotaColor(pct: number) {
  if (pct > 70) {
    return {
      dot: 'bg-emerald-500',
      text: 'text-emerald-500 dark:text-emerald-400',
      bar: 'bg-emerald-500',
    }
  }
  if (pct >= 30) {
    return {
      dot: 'bg-amber-500',
      text: 'text-amber-500 dark:text-amber-400',
      bar: 'bg-amber-500',
    }
  }
  return {
    dot: 'bg-rose-500',
    text: 'text-rose-500 dark:text-rose-400',
    bar: 'bg-rose-500',
  }
}

export function QuotaPage() {
  const [quotas, setQuotas] = useState<AccountQuota[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [refreshingIds, setRefreshingIds] = useState<Record<string, boolean>>({})
  const [error, setError] = useState<string | null>(null)

  const [providerFilter, setProviderFilter] = useState('all')
  const [accountFilter, setAccountFilter] = useState('all')
  const [expiringFirst, setExpiringFirst] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [countdown, setCountdown] = useState(30)
  const [hiddenRows, setHiddenRows] = useState<Record<string, boolean>>({})

  const [editingAccount, setEditingAccount] = useState<AccountQuota | null>(null)
  const [editName, setEditName] = useState('')
  const [isSubmittingEdit, setIsSubmittingEdit] = useState(false)

  const [providerDropdownOpen, setProviderDropdownOpen] = useState(false)
  const [accountDropdownOpen, setAccountDropdownOpen] = useState(false)

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
    if (!autoRefresh) return
    const interval = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          fetchQuotas(false)
          return 30
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(interval)
  }, [autoRefresh])

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement
      if (!target.closest('.provider-dropdown-menu')) {
        setProviderDropdownOpen(false)
      }
      if (!target.closest('.account-dropdown-menu')) {
        setAccountDropdownOpen(false)
      }
    }
    document.addEventListener('click', handleClickOutside)
    return () => document.removeEventListener('click', handleClickOutside)
  }, [])

  const handleRefreshSingle = async (accountId: string) => {
    setRefreshingIds((prev) => ({ ...prev, [accountId]: true }))
    try {
      const updated = await api.post<AccountQuota>(`/api/quota/${accountId}/refresh`, {})
      if (updated && updated.account_id) {
        setQuotas((prev) => prev.map((q) => (q.account_id === accountId ? updated : q)))
      }
    } catch {
      await fetchQuotas(true)
    } finally {
      setRefreshingIds((prev) => ({ ...prev, [accountId]: false }))
    }
  }

  const handleToggleAccount = async (accountId: string, currentEnabled: boolean) => {
    const next = !currentEnabled
    setQuotas((prev) =>
      prev.map((q) => (q.account_id === accountId ? { ...q, is_enabled: next } : q))
    )
    try {
      await api.patch(`/api/accounts/${accountId}`, { enabled: next })
    } catch {
      setQuotas((prev) =>
        prev.map((q) => (q.account_id === accountId ? { ...q, is_enabled: currentEnabled } : q))
      )
    }
  }

  const handleDeleteAccount = async (accountId: string, accountName: string) => {
    if (!window.confirm(`Delete account "${accountName}"?`)) return
    try {
      await api.delete(`/api/accounts/${accountId}`)
      setQuotas((prev) => prev.filter((q) => q.account_id !== accountId))
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Failed to delete account')
    }
  }

  const handleTurnOffEmpty = async () => {
    const emptyAccounts = quotas.filter(
      (q) => (q.is_enabled ?? true) && (q.overall_remaining <= 0 || !!q.error)
    )
    for (const acc of emptyAccounts) {
      try {
        await api.patch(`/api/accounts/${acc.account_id}`, { enabled: false })
      } catch {}
    }
    setQuotas((prev) =>
      prev.map((q) =>
        q.overall_remaining <= 0 || q.error ? { ...q, is_enabled: false } : q
      )
    )
  }

  const handleTurnOnAvailable = async () => {
    const availableAccounts = quotas.filter(
      (q) => !(q.is_enabled ?? true) && q.overall_remaining > 0 && !q.error
    )
    for (const acc of availableAccounts) {
      try {
        await api.patch(`/api/accounts/${acc.account_id}`, { enabled: true })
      } catch {}
    }
    setQuotas((prev) =>
      prev.map((q) =>
        q.overall_remaining > 0 && !q.error ? { ...q, is_enabled: true } : q
      )
    )
  }

  const toggleHideRow = (accountId: string, modelId: string) => {
    const key = `${accountId}:${modelId}`
    setHiddenRows((prev) => ({ ...prev, [key]: !prev[key] }))
  }

  const handleOpenEdit = (acc: AccountQuota) => {
    setEditingAccount(acc)
    setEditName(acc.account_name)
  }

  const handleSaveEdit = async () => {
    if (!editingAccount) return
    setIsSubmittingEdit(true)
    try {
      await api.patch(`/api/accounts/${editingAccount.account_id}`, {
        name: editName,
      })
      setQuotas((prev) =>
        prev.map((q) =>
          q.account_id === editingAccount.account_id
            ? { ...q, account_name: editName }
            : q
        )
      )
      setEditingAccount(null)
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Failed to update account')
    } finally {
      setIsSubmittingEdit(false)
    }
  }

  const uniqueProviders = useMemo(() => {
    const set = new Set<string>()
    quotas.forEach((q) => {
      if (q.provider_name) set.add(q.provider_name)
      else if (q.provider_id) set.add(q.provider_id)
    })
    return Array.from(set).sort()
  }, [quotas])

  const uniqueAccounts = useMemo(() => {
    return quotas.map((q) => ({
      id: q.account_id,
      name: q.account_name,
      email: q.email,
    }))
  }, [quotas])

  const displayedQuotas = useMemo(() => {
    let result = quotas.filter((q) => {
      if (providerFilter !== 'all') {
        const prov = (q.provider_name || q.provider_id).toLowerCase()
        if (prov !== providerFilter.toLowerCase()) return false
      }
      if (accountFilter !== 'all') {
        if (q.account_id !== accountFilter) return false
      }
      return true
    })

    if (expiringFirst) {
      result = [...result].sort((a, b) => {
        const timeA = a.reset_at ? new Date(a.reset_at).getTime() : Infinity
        const timeB = b.reset_at ? new Date(b.reset_at).getTime() : Infinity
        return timeA - timeB
      })
    }

    return result
  }, [quotas, providerFilter, accountFilter, expiringFirst])

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2.5 pt-1">
        <div className="flex items-center gap-2 flex-wrap">
          <div className="relative provider-dropdown-menu">
            <button
              type="button"
              onClick={() => {
                setProviderDropdownOpen(!providerDropdownOpen)
                setAccountDropdownOpen(false)
              }}
              className="h-8 px-3 text-[12px] font-medium rounded-[6px] border border-[#2e3344] bg-[#1a1c24] text-[#d1d5db] hover:border-[#3e4354] hover:text-white flex items-center gap-2 cursor-pointer transition-colors"
            >
              <LayoutGrid className="w-3.5 h-3.5 text-[#9ca3af]" />
              <span>
                {providerFilter === 'all' ? 'All Providers' : providerFilter}
              </span>
              <ChevronDown className="w-3.5 h-3.5 text-[#9ca3af]" />
            </button>
            {providerDropdownOpen && (
              <div className="absolute left-0 mt-1.5 w-64 max-h-80 overflow-y-auto rounded-[10px] bg-[#181a20] border border-[#2b2f3d] shadow-2xl py-1.5 z-40">
                <button
                  type="button"
                  onClick={() => {
                    setProviderFilter('all')
                    setProviderDropdownOpen(false)
                  }}
                  className="w-full flex items-center justify-between px-3.5 py-2 text-[12.5px] hover:bg-white/[0.04] cursor-pointer transition-colors"
                >
                  <div className="flex items-center gap-2.5">
                    <LayoutGrid className="w-4 h-4 text-orange-500" />
                    <span className={providerFilter === 'all' ? 'text-orange-500 font-semibold' : 'text-white'}>
                      All providers
                    </span>
                  </div>
                  {providerFilter === 'all' && <Check className="w-4 h-4 text-orange-500" />}
                </button>
                {uniqueProviders.map((p) => {
                  const isSelected = providerFilter.toLowerCase() === p.toLowerCase()
                  return (
                    <button
                      key={p}
                      type="button"
                      onClick={() => {
                        setProviderFilter(p)
                        setProviderDropdownOpen(false)
                      }}
                      className="w-full flex items-center justify-between px-3.5 py-2 text-[12.5px] hover:bg-white/[0.04] cursor-pointer transition-colors"
                    >
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div className="w-5 h-5 rounded-[4px] bg-[#222533] shrink-0 flex items-center justify-center overflow-hidden">
                          <ProviderLogo providerId={p.toLowerCase()} name={p} size="sm" />
                        </div>
                        <span className={`truncate capitalize ${isSelected ? 'text-orange-500 font-semibold' : 'text-[#e5e7eb]'}`}>
                          {p}
                        </span>
                      </div>
                      {isSelected && <Check className="w-4 h-4 text-orange-500 shrink-0" />}
                    </button>
                  )
                })}
              </div>
            )}
          </div>

          <div className="relative account-dropdown-menu">
            <button
              type="button"
              onClick={() => {
                setAccountDropdownOpen(!accountDropdownOpen)
                setProviderDropdownOpen(false)
              }}
              className="h-8 px-3 text-[12px] font-medium rounded-[6px] border border-[#2e3344] bg-[#1a1c24] text-[#d1d5db] hover:border-[#3e4354] hover:text-white flex items-center gap-2 cursor-pointer transition-colors"
            >
              <span className="truncate max-w-[130px]">
                {accountFilter === 'all'
                  ? 'All accounts'
                  : uniqueAccounts.find((a) => a.id === accountFilter)?.name ||
                    'Selected account'}
              </span>
              <ChevronDown className="w-3.5 h-3.5 text-[#9ca3af]" />
            </button>
            {accountDropdownOpen && (
              <div className="absolute left-0 mt-1.5 w-64 max-h-80 overflow-y-auto rounded-[10px] bg-[#181a20] border border-[#2b2f3d] shadow-2xl py-1.5 z-40">
                <button
                  type="button"
                  onClick={() => {
                    setAccountFilter('all')
                    setAccountDropdownOpen(false)
                  }}
                  className="w-full flex items-center justify-between px-3.5 py-2 text-[12.5px] hover:bg-white/[0.04] cursor-pointer transition-colors"
                >
                  <span className={accountFilter === 'all' ? 'text-orange-500 font-semibold' : 'text-white'}>
                    All accounts
                  </span>
                  {accountFilter === 'all' && <Check className="w-4 h-4 text-orange-500" />}
                </button>
                {uniqueAccounts.map((a) => {
                  const isSelected = accountFilter === a.id
                  return (
                    <button
                      key={a.id}
                      type="button"
                      onClick={() => {
                        setAccountFilter(a.id)
                        setAccountDropdownOpen(false)
                      }}
                      className="w-full flex items-center justify-between px-3.5 py-2 text-[12.5px] hover:bg-white/[0.04] cursor-pointer transition-colors"
                    >
                      <span className={`truncate ${isSelected ? 'text-orange-500 font-semibold' : 'text-[#e5e7eb]'}`}>
                        {a.name || a.email}
                      </span>
                      {isSelected && <Check className="w-4 h-4 text-orange-500 shrink-0" />}
                    </button>
                  )
                })}
              </div>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <button
            type="button"
            onClick={() => setExpiringFirst(!expiringFirst)}
            className={`h-8 px-3 text-[12px] font-medium rounded-[6px] border flex items-center gap-1.5 cursor-pointer transition-colors ${
              expiringFirst
                ? 'border-amber-500/40 bg-amber-500/10 text-amber-400'
                : 'border-[#2e3344] bg-[#1a1c24] text-[#d1d5db] hover:border-[#3e4354] hover:text-white'
            }`}
          >
            <Hourglass className="w-3.5 h-3.5 text-amber-400" />
            <span>Expiring first</span>
          </button>

          <button
            type="button"
            onClick={handleTurnOffEmpty}
            className="h-8 px-3 text-[12px] font-medium rounded-[6px] border border-rose-500/40 bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 flex items-center gap-1.5 cursor-pointer transition-colors"
          >
            <Ban className="w-3.5 h-3.5 text-rose-400" />
            <span>Turn off Empty</span>
          </button>

          <button
            type="button"
            onClick={handleTurnOnAvailable}
            className="h-8 px-3 text-[12px] font-medium rounded-[6px] border border-emerald-500/40 bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 flex items-center gap-1.5 cursor-pointer transition-colors"
          >
            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
            <span>Turn on Available</span>
          </button>

          <button
            type="button"
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={`h-8 px-3 text-[12px] font-medium rounded-[6px] border flex items-center gap-1.5 cursor-pointer transition-colors ${
              autoRefresh
                ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-400'
                : 'border-[#2e3344] bg-[#1a1c24] text-[#9ca3af] hover:text-white'
            }`}
          >
            {autoRefresh ? (
              <ToggleRight className="w-4 h-4 text-emerald-400" />
            ) : (
              <ToggleLeft className="w-4 h-4 text-[#9ca3af]" />
            )}
            <span>Auto-refresh</span>
            {autoRefresh && (
              <span className="text-[11px] font-mono tabular-nums opacity-80">
                ({countdown}s)
              </span>
            )}
          </button>

          <button
            type="button"
            onClick={() => fetchQuotas(true)}
            disabled={isRefreshing || isLoading}
            className="h-8 w-8 rounded-[6px] border border-[#2e3344] bg-[#1a1c24] text-[#d1d5db] hover:border-[#3e4354] hover:text-white flex items-center justify-center cursor-pointer transition-colors disabled:opacity-50"
            title="Refresh all"
          >
            <RefreshCw
              className={`w-3.5 h-3.5 ${
                isRefreshing || isLoading ? 'animate-spin' : ''
              }`}
            />
          </button>
        </div>
      </div>

      {error && <ErrorBanner message={error} onRetry={() => fetchQuotas(true)} />}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3.5">
        {displayedQuotas.map((item) => {
          const isEnabled = item.is_enabled ?? true
          const isBusy = refreshingIds[item.account_id] || false
          const visibleQuotas = item.quotas.filter(
            (q) => !hiddenRows[`${item.account_id}:${q.id}`]
          )

          return (
            <div
              key={item.account_id}
              className={`rounded-[10px] border border-[#282b3a] bg-[#171922] transition-colors overflow-hidden ${
                !isEnabled ? 'opacity-60' : ''
              }`}
            >
              <div className="px-3.5 py-2.5 border-b border-[#232635] flex items-center justify-between gap-2">
                <div className="flex items-center gap-2.5 min-w-0">
                  <div className="w-8 h-8 rounded-[6px] shrink-0 flex items-center justify-center bg-[#202330] overflow-hidden">
                    <ProviderLogo
                      providerId={item.provider_id}
                      name={item.provider_name}
                      size="sm"
                    />
                  </div>
                  <div className="min-w-0">
                    <h3 className="text-[13px] font-semibold text-white truncate capitalize leading-tight">
                      {item.provider_name || item.provider_id}
                    </h3>
                    <p className="text-[11px] text-[#9ca3af] truncate leading-tight mt-0.5">
                      {item.email || item.account_name}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-1.5 shrink-0">
                  <button
                    type="button"
                    onClick={() => handleRefreshSingle(item.account_id)}
                    disabled={isBusy}
                    title="Refresh quota"
                    className="w-7 h-7 rounded-[5px] flex items-center justify-center text-[#9ca3af] hover:text-white hover:bg-white/5 transition-colors cursor-pointer disabled:opacity-50"
                  >
                    <RefreshCw
                      className={`w-3.5 h-3.5 ${isBusy ? 'animate-spin' : ''}`}
                    />
                  </button>

                  <button
                    type="button"
                    onClick={() => handleOpenEdit(item)}
                    title="Edit account"
                    className="w-7 h-7 rounded-[5px] flex items-center justify-center text-[#9ca3af] hover:text-white hover:bg-white/5 transition-colors cursor-pointer"
                  >
                    <Pencil className="w-3.5 h-3.5" />
                  </button>

                  <button
                    type="button"
                    onClick={() =>
                      handleDeleteAccount(item.account_id, item.account_name)
                    }
                    title="Delete account"
                    className="w-7 h-7 rounded-[5px] flex items-center justify-center text-rose-500/80 hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>

                  <div className="pl-1">
                    <button
                      type="button"
                      role="switch"
                      aria-checked={isEnabled}
                      onClick={() =>
                        handleToggleAccount(item.account_id, isEnabled)
                      }
                      className={`w-8 h-4.5 rounded-full transition-colors relative cursor-pointer flex items-center p-0.5 ${
                        isEnabled ? 'bg-orange-500' : 'bg-[#374151]'
                      }`}
                    >
                      <span
                        className={`w-3.5 h-3.5 rounded-full bg-white transition-transform ${
                          isEnabled ? 'translate-x-3.5' : 'translate-x-0'
                        }`}
                      />
                    </button>
                  </div>
                </div>
              </div>

              <div className="px-3.5 py-3">
                {item.error ? (
                  <div className="py-7 px-4 text-center">
                    <AlertCircle className="w-6 h-6 text-rose-500 mx-auto mb-2" />
                    <p className="text-[12px] text-rose-300 font-mono break-all leading-relaxed">
                      {item.error}
                    </p>
                  </div>
                ) : item.quotas && item.quotas.length > 0 ? (
                  <div className="space-y-1.5">
                    <div className="text-[10px] text-[#9ca3af] font-medium mb-1">
                      {item.quotas.length} quota
                      {item.quotas.length > 1 ? 's' : ''}
                    </div>
                    {visibleQuotas.map((q) => {
                      const colors = getQuotaColor(q.remaining_percentage)
                      const countdownStr = formatResetCountdown(q.reset_at)
                      return (
                        <div
                          key={q.id}
                          className="flex items-center gap-2.5 py-1 border-b border-white/[0.04] last:border-0 hover:bg-white/[0.02] px-1 rounded-[4px] transition-colors"
                        >
                          <div className="flex items-center gap-1.5 w-36 sm:w-44 shrink-0 min-w-0">
                            <span
                              className={`w-2 h-2 rounded-full shrink-0 ${colors.dot}`}
                            />
                            <span
                              className="text-[11.5px] font-medium text-[#e5e7eb] truncate"
                              title={q.name}
                            >
                              {q.name}
                            </span>
                          </div>

                          <div className="flex-1 min-w-0 space-y-1">
                            <div className="h-1 rounded-full bg-[#272a39] overflow-hidden">
                              <div
                                className={`h-full rounded-full transition-all duration-300 ${colors.bar}`}
                                style={{
                                  width: `${Math.min(
                                    q.remaining_percentage,
                                    100
                                  )}%`,
                                }}
                              />
                            </div>
                            <div className="flex items-center justify-between text-[10px] leading-tight">
                              <span className="text-[#9ca3af] font-mono">
                                {q.used.toLocaleString()} /{' '}
                                {q.total.toLocaleString()}
                              </span>
                              <span
                                className={`font-medium font-mono ${colors.text}`}
                              >
                                {Math.round(q.remaining_percentage)}%
                              </span>
                            </div>
                          </div>

                          <div className="shrink-0 min-w-[72px] text-right">
                            <span className="text-[11px] font-medium text-[#e5e7eb] font-mono">
                              {countdownStr}
                            </span>
                          </div>

                          <button
                            type="button"
                            onClick={() => toggleHideRow(item.account_id, q.id)}
                            className="p-1 text-[#6b7280] hover:text-[#d1d5db] transition-colors cursor-pointer shrink-0"
                            title="Hide row"
                          >
                            <EyeOff className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      )
                    })}
                  </div>
                ) : item.message ? (
                  <div className="py-7 px-4 text-center">
                    <p className="text-[11.5px] text-[#9ca3af] leading-relaxed max-w-sm mx-auto">
                      {item.message}
                    </p>
                  </div>
                ) : (
                  <div className="py-7 px-4 text-center">
                    <p className="text-[11.5px] text-[#9ca3af]">
                      Quota tracking not supported for this provider
                    </p>
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>

      {editingAccount && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-xs flex items-center justify-center p-4 z-50">
          <div className="w-full max-w-md rounded-[10px] border border-[#2e3344] bg-[#1a1c24] p-5 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-[#2e3344] pb-3">
              <h3 className="text-[14px] font-semibold text-white">
                Edit Connection
              </h3>
              <button
                type="button"
                onClick={() => setEditingAccount(null)}
                className="text-[#9ca3af] hover:text-white cursor-pointer p-1"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-[11px] font-medium text-[#9ca3af] uppercase tracking-wide mb-1">
                  Account Name
                </label>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="w-full h-8 px-3 text-[12.5px] rounded-[6px] bg-[#12141a] border border-[#2e3344] text-white focus:outline-none focus:border-orange-500 transition-colors"
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-[#9ca3af] uppercase tracking-wide mb-1">
                  Provider
                </label>
                <input
                  type="text"
                  disabled
                  value={
                    editingAccount.provider_name || editingAccount.provider_id
                  }
                  className="w-full h-8 px-3 text-[12.5px] rounded-[6px] bg-[#12141a]/50 border border-[#2e3344] text-[#6b7280] cursor-not-allowed"
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 pt-2 border-t border-[#2e3344]">
              <button
                type="button"
                onClick={() => setEditingAccount(null)}
                className="h-7 px-3 text-[11.5px] font-medium rounded-[5px] border border-[#2e3344] text-[#d1d5db] hover:text-white hover:border-[#3e4354] transition-colors cursor-pointer"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleSaveEdit}
                disabled={isSubmittingEdit}
                className="h-7 px-3 text-[11.5px] font-medium rounded-[5px] bg-orange-500 hover:bg-orange-600 text-white transition-colors cursor-pointer disabled:opacity-50"
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
