import { useEffect, useState } from 'react'
import { AlertTriangle, RotateCcw, RefreshCw, CheckCircle2 } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { Account } from '../../types/api.ts'

export function QuotaPage() {
  const [accounts, setAccounts] = useState<Account[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [resetFeedback, setResetFeedback] = useState<string | null>(null)

  const loadAccounts = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.get<Account[]>('/api/accounts')
      setAccounts(data || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load account quotas'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadAccounts()
  }, [])

  const handleResetCooldown = async (account: Account) => {
    try {
      await api.post('/api/accounts', {
        ...account,
        state: 'active',
        consecutive_failures: 0,
        cooldown_until: null,
      })
      setResetFeedback(`Circuit reset for ${account.name}`)
      setTimeout(() => setResetFeedback(null), 3000)
      await loadAccounts()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to reset circuit'
      setError(msg)
    }
  }

  const coolingDownAccounts = accounts.filter((a) => a.state === 'cooling_down')

  const columns: Column<Account>[] = [
    {
      key: 'name',
      title: 'Account Name',
      render: (item) => (
        <div className="flex items-center gap-2.5">
          <ProviderLogo providerId={item.provider_id} size="sm" />
          <div className="flex flex-col">
            <span className="font-semibold text-[13px] text-[var(--text-primary)]">{item.name}</span>
            <span className="text-[10.5px] font-mono text-[var(--text-muted)]">{item.provider_id}</span>
          </div>
        </div>
      ),
    },
    {
      key: 'priority',
      title: 'Priority Tier',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--brand-text)] font-semibold">
          Tier {item.priority}
        </span>
      ),
    },
    {
      key: 'state',
      title: 'Circuit State',
      render: (item) => (
        <StatusBadge
          variant={
            item.state === 'active'
              ? 'healthy'
              : item.state === 'cooling_down'
              ? 'cooling_down'
              : 'disabled'
          }
          pulse={item.state === 'cooling_down'}
        >
          {item.state}
        </StatusBadge>
      ),
    },
    {
      key: 'consecutive_failures',
      title: 'Consecutive Errors',
      render: (item) => (
        <span
          className={`font-mono text-[12px] font-medium ${
            (item.consecutive_failures || 0) > 0
              ? 'text-[var(--status-danger)]'
              : 'text-[var(--text-muted)]'
          }`}
        >
          {item.consecutive_failures || 0}
        </span>
      ),
    },
    {
      key: 'cooldown_until',
      title: 'Cooldown Expiry',
      render: (item) => (
        <span className="font-mono text-[11px] text-[var(--text-muted)]">
          {item.cooldown_until || '-'}
        </span>
      ),
    },
    {
      key: 'actions',
      title: 'Actions',
      width: '130px',
      render: (item) =>
        item.state === 'cooling_down' ? (
          <Button
            variant="secondary"
            size="compact"
            onClick={() => handleResetCooldown(item)}
            leftIcon={<RotateCcw className="w-3 h-3" />}
          >
            Reset Circuit
          </Button>
        ) : (
          <span className="text-[11px] text-[var(--text-muted)]">Normal</span>
        ),
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Quotas & Circuits"
        description="Monitor rate-limit backoff timers, error tripwires, and manually reset cooldown circuits."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Quotas & Circuits' },
        ]}
        metadata={
          <span>
            {coolingDownAccounts.length} accounts in backoff &bull; {accounts.length} total
          </span>
        }
        actions={
          <Button
            variant="secondary"
            size="compact"
            onClick={loadAccounts}
            isLoading={isLoading}
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
          >
            Refresh
          </Button>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadAccounts} />}

      {resetFeedback && (
        <div className="p-3 rounded-[6px] bg-emerald-950/20 border border-emerald-600/30 text-[12px] text-emerald-300 flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          <span>{resetFeedback}</span>
        </div>
      )}

      {coolingDownAccounts.length > 0 && (
        <div className="bg-amber-950/20 border border-amber-600/30 rounded-[8px] p-4 space-y-2">
          <div className="flex items-center gap-2 text-amber-300 font-semibold text-[13px]">
            <AlertTriangle className="w-4 h-4" />
            <span>Accounts Currently Tripped</span>
          </div>
          <p className="text-[12px] text-[var(--text-secondary)]">
            The following accounts encountered 429 rate limits or authentication rejections and are paused from the routing ring:
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1">
            {coolingDownAccounts.map((acc) => (
              <div
                key={acc.id}
                className="flex items-center justify-between bg-[var(--bg-card)] px-3 py-2 rounded-[6px] text-[12px] border border-[var(--border-subtle)]"
              >
                <div className="flex items-center gap-2">
                  <ProviderLogo providerId={acc.provider_id} size="sm" />
                  <div>
                    <span className="font-semibold text-[var(--text-primary)] block">
                      {acc.name}
                    </span>
                    <span className="text-[10.5px] font-mono text-[var(--text-muted)]">
                      {acc.provider_id}
                    </span>
                  </div>
                </div>
                <Button
                  variant="secondary"
                  size="compact"
                  onClick={() => handleResetCooldown(acc)}
                  leftIcon={<RotateCcw className="w-3 h-3" />}
                >
                  Reset
                </Button>
              </div>
            ))}
          </div>
        </div>
      )}

      <DataTable
        columns={columns}
        data={accounts}
        isLoading={isLoading}
        keyExtractor={(a) => a.id}
        emptyMessage="No provider accounts registered."
      />
    </div>
  )
}
