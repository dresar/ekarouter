import { useEffect, useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Save, Plus, Trash2, AlertCircle } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'
import { Provider, Account, RouteItem } from '../../types/api.ts'

interface TargetRow {
  id: string
  provider_id: string
  account_id: string
  priority: number
  weight: number
}

export function RouteCreatePage() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [strategy, setStrategy] = useState<'priority' | 'round_robin' | 'least_used'>('priority')
  const [enabled, setEnabled] = useState(true)
  const [targets, setTargets] = useState<TargetRow[]>([])

  const [providers, setProviders] = useState<Provider[]>([])
  const [accounts, setAccounts] = useState<Account[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadMeta() {
      try {
        const [provData, accData] = await Promise.all([
          api.get<Provider[]>('/api/providers'),
          api.get<Account[]>('/api/accounts'),
        ])
        setProviders(provData || [])
        setAccounts(accData || [])

        if (provData && provData.length > 0) {
          const firstProv = provData[0]
          const matchingAcc = (accData || []).find((a) => a.provider_id === firstProv.id)
          setTargets([
            {
              id: 'target_1',
              provider_id: firstProv.id,
              account_id: matchingAcc ? matchingAcc.id : '',
              priority: 1,
              weight: 10,
            },
          ])
        }
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : 'Failed to load metadata'
        setError(msg)
      } finally {
        setIsLoading(false)
      }
    }

    loadMeta()
  }, [])

  const handleAddTarget = () => {
    const defaultProv = providers[0]?.id || ''
    const matchingAcc = accounts.find((a) => a.provider_id === defaultProv)?.id || ''
    setTargets((prev) => [
      ...prev,
      {
        id: `target_${Date.now()}`,
        provider_id: defaultProv,
        account_id: matchingAcc,
        priority: prev.length + 1,
        weight: 10,
      },
    ])
  }

  const handleRemoveTarget = (targetId: string) => {
    setTargets((prev) => prev.filter((t) => t.id !== targetId))
  }

  const handleTargetChange = (targetId: string, field: keyof TargetRow, value: unknown) => {
    setTargets((prev) =>
      prev.map((t) => {
        if (t.id !== targetId) return t
        const updated = { ...t, [field]: value }
        if (field === 'provider_id') {
          const matchingAcc = accounts.find((a) => a.provider_id === value)?.id || ''
          updated.account_id = matchingAcc
        }
        return updated
      })
    )
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      setError('Route name is required')
      return
    }
    if (targets.length === 0) {
      setError('At least one target account is required')
      return
    }

    setIsSubmitting(true)
    setError(null)

    const routeItems: Partial<RouteItem>[] = targets.map((t) => ({
      provider_id: t.provider_id,
      account_id: t.account_id || undefined,
      priority: Number(t.priority),
      weight: Number(t.weight),
      enabled: true,
    }))

    try {
      await api.post('/api/routes', {
        name: name.trim(),
        strategy,
        enabled,
        items: routeItems,
      })
      navigate('/routing')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create route'
      setError(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-3xl mx-auto space-y-5">
      <PageHeader
        title="Create Routing Combo"
        description="Bind an ingress model alias to ordered upstream targets with automatic failover."
        breadcrumbs={[
          { label: 'Routing', to: '/routing' },
          { label: 'New Route' },
        ]}
        actions={
          <Link to="/routing">
            <Button variant="ghost" size="compact" leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}>
              Back to List
            </Button>
          </Link>
        }
      />

      {error && (
        <div className="p-3 rounded-[6px] bg-[var(--status-danger)]/10 border border-[var(--status-danger)]/30 flex items-center gap-2 text-[12.5px] text-[var(--status-danger)]">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-5 space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                Route Alias / Model Slug *
              </label>
              <input
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. gpt-4o or claude-code"
                className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
              />
              <span className="text-[11px] text-[var(--text-muted)] mt-1 block">
                The model identifier sent in /v1/chat/completions
              </span>
            </div>

            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                Dispatch Strategy *
              </label>
              <select
                value={strategy}
                onChange={(e) =>
                  setStrategy(e.target.value as 'priority' | 'round_robin' | 'least_used')
                }
                className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
              >
                <option value="priority">Priority Fallback (Tier 1 then Tier 2)</option>
                <option value="round_robin">Round Robin (Distribute evenly)</option>
                <option value="least_used">Least Used (Balance load)</option>
              </select>
            </div>
          </div>

          <div className="pt-2 border-t border-[var(--border-subtle)]">
            <label className="flex items-center gap-2 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                className="rounded w-4 h-4 text-[var(--brand-primary)]"
              />
              <span className="text-[13px] font-medium text-[var(--text-primary)]">
                Enable this route immediately for gateway traffic
              </span>
            </label>
          </div>
        </div>

        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-5 space-y-3.5">
          <div className="flex items-center justify-between border-b border-[var(--border-subtle)] pb-2.5">
            <div>
              <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
                Upstream Target Accounts
              </h3>
              <p className="text-[11px] text-[var(--text-muted)]">
                Define the sequence of accounts to try upon failure or load balancing
              </p>
            </div>
            <Button
              type="button"
              variant="secondary"
              size="compact"
              onClick={handleAddTarget}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
            >
              Add Target
            </Button>
          </div>

          {isLoading ? (
            <div className="py-4 text-center text-[12px] text-[var(--text-muted)]">
              Loading available accounts...
            </div>
          ) : targets.length === 0 ? (
            <div className="py-6 text-center text-[12.5px] text-[var(--text-muted)]">
              No target accounts assigned. Click &quot;Add Target&quot; above.
            </div>
          ) : (
            <div className="space-y-2.5">
              {targets.map((target, idx) => {
                const targetAccounts = accounts.filter(
                  (a) => a.provider_id === target.provider_id
                )
                return (
                  <div
                    key={target.id}
                    className="p-3 bg-[var(--bg-panel)]/50 border border-[var(--border-strong)] rounded-[6px] grid grid-cols-1 sm:grid-cols-12 gap-2.5 items-center text-[12px]"
                  >
                    <div className="sm:col-span-1 flex items-center justify-center font-mono font-bold text-[var(--brand-text)]">
                      #{idx + 1}
                    </div>

                    <div className="sm:col-span-4">
                      <label className="block text-[10px] uppercase font-semibold text-[var(--text-muted)] mb-1">
                        Provider
                      </label>
                      <select
                        value={target.provider_id}
                        onChange={(e) =>
                          handleTargetChange(target.id, 'provider_id', e.target.value)
                        }
                        className="w-full px-2 py-1 text-[12px] rounded focus:outline-none"
                      >
                        {providers.map((p) => (
                          <option key={p.id} value={p.id}>
                            {p.name} ({p.kind})
                          </option>
                        ))}
                      </select>
                    </div>

                    <div className="sm:col-span-4">
                      <label className="block text-[10px] uppercase font-semibold text-[var(--text-muted)] mb-1">
                        Account Key
                      </label>
                      <select
                        value={target.account_id}
                        onChange={(e) =>
                          handleTargetChange(target.id, 'account_id', e.target.value)
                        }
                        className="w-full px-2 py-1 text-[12px] rounded focus:outline-none"
                      >
                        <option value="">Default Provider Credential</option>
                        {targetAccounts.map((a) => (
                          <option key={a.id} value={a.id}>
                            {a.name} (Tier {a.priority})
                          </option>
                        ))}
                      </select>
                    </div>

                    <div className="sm:col-span-1">
                      <label className="block text-[10px] uppercase font-semibold text-[var(--text-muted)] mb-1">
                        Priority
                      </label>
                      <input
                        type="number"
                        min="1"
                        max="10"
                        value={target.priority}
                        onChange={(e) =>
                          handleTargetChange(
                            target.id,
                            'priority',
                            parseInt(e.target.value) || 1
                          )
                        }
                        className="w-full px-2 py-1 text-[12px] rounded focus:outline-none"
                      />
                    </div>

                    <div className="sm:col-span-1">
                      <label className="block text-[10px] uppercase font-semibold text-[var(--text-muted)] mb-1">
                        Weight
                      </label>
                      <input
                        type="number"
                        min="1"
                        max="100"
                        value={target.weight}
                        onChange={(e) =>
                          handleTargetChange(
                            target.id,
                            'weight',
                            parseInt(e.target.value) || 10
                          )
                        }
                        className="w-full px-2 py-1 text-[12px] rounded focus:outline-none"
                      />
                    </div>

                    <div className="sm:col-span-1 flex justify-center pt-3 sm:pt-0">
                      <button
                        type="button"
                        onClick={() => handleRemoveTarget(target.id)}
                        className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] transition-colors"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>

        <div className="flex items-center justify-end gap-2 pt-2">
          <Link to="/routing">
            <Button type="button" variant="ghost" size="compact">
              Cancel
            </Button>
          </Link>
          <Button
            type="submit"
            variant="primary"
            size="compact"
            isLoading={isSubmitting}
            leftIcon={<Save className="w-3.5 h-3.5" />}
          >
            Save Routing Combo
          </Button>
        </div>
      </form>
    </div>
  )
}
