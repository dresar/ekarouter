import { useEffect, useState, FormEvent } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Plus,
  Trash2,
  Server,
  AlertCircle,
  ExternalLink,
  Radio,
  Shield,
  Globe,
  Pencil,
  ChevronUp,
  ChevronDown,
  CheckCircle2,
  XCircle,
  Clock,
  ToggleLeft,
  ToggleRight,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { BottomSheet } from '../../components/ui/BottomSheet.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { api } from '../../api/client.ts'
import { Provider, Account, PlatformProvider, ProxyProfile } from '../../types/api.ts'

export function ProviderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [provider, setProvider] = useState<Provider | null>(null)
  const [platProvider, setPlatProvider] = useState<PlatformProvider | null>(null)
  const [accounts, setAccounts] = useState<Account[]>([])
  const [proxies, setProxies] = useState<ProxyProfile[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [testResult, setTestResult] = useState<{ status: 'idle' | 'testing' | 'success' | 'error'; message?: string; latency?: number }>({ status: 'idle' })

  const [showAddSheet, setShowAddSheet] = useState(false)
  const [showEditSheet, setShowEditSheet] = useState(false)
  const [editingAccount, setEditingAccount] = useState<Account | null>(null)

  const [addForm, setAddForm] = useState({
    name: '',
    auth_type: 'apikey',
    priority: 1,
    api_key: '',
    proxy_pool_id: '',
  })
  const [editForm, setEditForm] = useState({
    name: '',
    priority: 1,
    enabled: true,
    proxy_pool_id: '',
    api_key: '',
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set())

  const loadData = async () => {
    if (!id) return
    setIsLoading(true)
    setError(null)
    try {
      const [allProvidersRes, platformProvidersRes, allAccountsRes, proxiesRes] = await Promise.allSettled([
        api.get<Provider[]>('/api/providers'),
        api.get<PlatformProvider[]>('/api/v1/providers'),
        api.get<Account[]>('/api/accounts'),
        api.get<ProxyProfile[]>('/api/proxies'),
      ])

      const allProviders: Provider[] = allProvidersRes.status === 'fulfilled' && Array.isArray(allProvidersRes.value) ? allProvidersRes.value : []
      const platformProviders: PlatformProvider[] = platformProvidersRes.status === 'fulfilled' && Array.isArray(platformProvidersRes.value) ? platformProvidersRes.value : []
      const allAccounts: Account[] = allAccountsRes.status === 'fulfilled' && Array.isArray(allAccountsRes.value) ? allAccountsRes.value : []
      const allProxies: ProxyProfile[] = proxiesRes.status === 'fulfilled' && Array.isArray(proxiesRes.value) ? proxiesRes.value : []

      setProxies(allProxies)

      let found = allProviders.find((p) => p.id === id || p.key === id)
      let platFound = platformProviders.find((p) => p.id === id)

      if (!found && platFound) {
        found = {
          id: platFound.id,
          key: platFound.id,
          name: platFound.name,
          kind: platFound.auth_type || 'bearer',
          base_url: platFound.base_url,
          enabled: platFound.enabled ?? true,
        }
      } else if (!found) {
        try {
          const single = await api.get<PlatformProvider>(`/api/v1/providers/${id}`)
          if (single && single.id) {
            platFound = single
            found = {
              id: single.id,
              key: single.id,
              name: single.name,
              kind: single.auth_type || 'bearer',
              base_url: single.base_url,
              enabled: true,
            }
          }
        } catch {
          const adminMatch = allProviders.find((p) => p.key === id)
          if (adminMatch) found = adminMatch
        }
      }

      if (!found) {
        setError('Provider not found')
        return
      }

      setProvider(found)
      setPlatProvider(platFound || null)

      const providerAccounts = allAccounts.filter(
        (a) => a.provider_id === id || a.provider_id === found?.key || a.provider_id === found?.id
      )
      setAccounts(providerAccounts.sort((a, b) => b.priority - a.priority))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load provider details'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [id])

  const handleTest = async () => {
    if (!provider) return
    setTestResult({ status: 'testing' })
    const start = performance.now()
    try {
      await api.post(`/api/v1/providers/${provider.id}/health`, {})
      setTestResult({ status: 'success', latency: Math.round(performance.now() - start), message: 'Connection healthy' })
    } catch {
      setTestResult({ status: 'error', latency: Math.round(performance.now() - start), message: 'Connection failed' })
    }
  }

  const handleAddSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!provider || !addForm.name.trim() || !addForm.api_key.trim()) return
    setIsSubmitting(true)
    setFormError(null)
    try {
      await api.post('/api/accounts', {
        provider_id: provider.id,
        name: addForm.name.trim(),
        auth_type: addForm.auth_type,
        priority: Number(addForm.priority),
        api_key: addForm.api_key.trim(),
        proxy_pool_id: addForm.proxy_pool_id || undefined,
      })
      setAddForm({ name: '', auth_type: 'apikey', priority: 1, api_key: '', proxy_pool_id: '' })
      setShowAddSheet(false)
      await loadData()
    } catch (err: unknown) {
      setFormError(err instanceof Error ? err.message : 'Failed to add connection')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleEditSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!editingAccount) return
    setIsSubmitting(true)
    setFormError(null)
    try {
      await api.put(`/api/accounts/${editingAccount.id}`, {
        name: editForm.name,
        priority: Number(editForm.priority),
        enabled: editForm.enabled,
        proxy_pool_id: editForm.proxy_pool_id || undefined,
        api_key: editForm.api_key || undefined,
      })
      setShowEditSheet(false)
      setEditingAccount(null)
      await loadData()
    } catch (err: unknown) {
      setFormError(err instanceof Error ? err.message : 'Failed to update connection')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async (accountId: string) => {
    try {
      await api.delete(`/api/accounts/${accountId}`)
      setAccounts((prev) => prev.filter((a) => a.id !== accountId))
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to delete')
    }
  }

  const handleToggle = async (acc: Account) => {
    try {
      await api.patch(`/api/accounts/${acc.id}`, { enabled: !acc.enabled })
      setAccounts((prev) =>
        prev.map((a) => (a.id === acc.id ? { ...a, enabled: !a.enabled } : a))
      )
    } catch {
      /* silent */
    }
  }

  const openEdit = (acc: Account) => {
    setEditingAccount(acc)
    setEditForm({
      name: acc.name,
      priority: acc.priority,
      enabled: acc.enabled,
      proxy_pool_id: acc.proxy_pool_id || '',
      api_key: '',
    })
    setFormError(null)
    setShowEditSheet(true)
  }

  const toggleExpand = (id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const getStateVariant = (acc: Account) => {
    if (!acc.enabled) return 'disabled'
    if (acc.state === 'active') return 'healthy'
    if (acc.state === 'cooling_down') return 'cooling_down'
    return 'disabled'
  }

  const getStateLabel = (acc: Account) => {
    if (!acc.enabled) return 'disabled'
    return acc.state || 'unknown'
  }

  const proxyById = (proxyId?: string) => proxies.find((p) => p.id === proxyId)

  return (
    <div className="space-y-5 pb-24">
      <PageHeader
        title={provider ? provider.name : 'Provider Details'}
        description={`Provider ID: ${id || ''}`}
        breadcrumbs={[
          { label: 'Providers', to: '/providers' },
          { label: provider?.name || 'Details' },
        ]}
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="compact"
              onClick={() => navigate('/providers')}
              leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}
            >
              Back
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadData} />}

      {provider && (
        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] p-4 flex flex-col sm:flex-row sm:items-center gap-4">
          <div className="flex items-center gap-4 flex-1 min-w-0">
            <ProviderLogo providerId={provider.id} name={provider.name} size="xl" />
            <div className="min-w-0">
              <div className="flex items-center gap-2.5 flex-wrap">
                <h2 className="text-[17px] font-bold text-[var(--text-primary)]">{provider.name}</h2>
                {platProvider?.doc_url && (
                  <a
                    href={platProvider.doc_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-1 text-[11.5px] text-[var(--brand-text)] hover:underline"
                  >
                    <ExternalLink className="w-3 h-3" />
                    Get API Key
                  </a>
                )}
              </div>
              <p className="text-[12px] font-mono text-[var(--text-muted)] mt-0.5 truncate">
                {accounts.length} connections
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <Button
              variant="secondary"
              size="compact"
              onClick={handleTest}
              isLoading={testResult.status === 'testing'}
              leftIcon={<Radio className="w-3.5 h-3.5" />}
            >
              Test Connection
            </Button>
          </div>
        </div>
      )}

      {testResult.status !== 'idle' && (
        <div
          className={`px-4 py-2.5 rounded-[7px] flex items-center gap-2.5 text-[12px] border ${
            testResult.status === 'success'
              ? 'bg-emerald-950/20 border-emerald-600/30 text-emerald-300'
              : testResult.status === 'error'
              ? 'bg-rose-950/20 border-rose-600/30 text-rose-300'
              : 'bg-[var(--bg-panel)] border-[var(--border-subtle)] text-[var(--text-muted)]'
          }`}
        >
          {testResult.status === 'success' ? (
            <CheckCircle2 className="w-4 h-4 shrink-0" />
          ) : testResult.status === 'error' ? (
            <XCircle className="w-4 h-4 shrink-0" />
          ) : (
            <Clock className="w-4 h-4 shrink-0 animate-spin" />
          )}
          <span>{testResult.message || 'Testing…'}</span>
          {testResult.latency !== undefined && (
            <span className="ml-auto font-mono text-[10.5px]">{testResult.latency}ms</span>
          )}
        </div>
      )}

      <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--border-subtle)]">
          <div className="flex items-center gap-2">
            <h3 className="text-[13.5px] font-semibold text-[var(--text-primary)]">
              Connections
            </h3>
            <span className="text-[11px] font-mono text-[var(--text-muted)] bg-[var(--bg-panel)] px-1.5 py-0.5 rounded">
              {accounts.length}
            </span>
          </div>

          <Button
            variant="primary"
            size="compact"
            onClick={() => {
              setAddForm({ name: '', auth_type: 'apikey', priority: accounts.length + 1, api_key: '', proxy_pool_id: '' })
              setFormError(null)
              setShowAddSheet(true)
            }}
            leftIcon={<Plus className="w-3.5 h-3.5" />}
          >
            Add Connection
          </Button>
        </div>

        {isLoading ? (
          <div className="py-12 flex items-center justify-center">
            <div className="w-6 h-6 rounded-full border-2 border-[var(--border-strong)] border-t-[var(--brand-primary)] animate-spin" />
          </div>
        ) : accounts.length === 0 ? (
          <div className="py-14 text-center">
            <Server className="w-8 h-8 text-[var(--text-muted)] mx-auto mb-2 opacity-40" />
            <p className="text-[13px] text-[var(--text-secondary)] font-medium">No connections configured</p>
            <p className="text-[11.5px] text-[var(--text-muted)] mt-1">
              Add an API key or credential to start routing traffic.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-[var(--border-subtle)]">
            {accounts.map((acc, idx) => {
              const isExpanded = expandedIds.has(acc.id)
              const proxy = proxyById(acc.proxy_pool_id)
              const hasError = !!acc.last_error

              return (
                <div key={acc.id} className="group">
                  <div className="flex items-center gap-3 px-4 py-3">
                    <div className="flex items-center gap-2 shrink-0">
                      <button
                        type="button"
                        onClick={() => toggleExpand(acc.id)}
                        className="p-0.5 text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
                      >
                        {isExpanded ? (
                          <ChevronUp className="w-3.5 h-3.5" />
                        ) : (
                          <ChevronDown className="w-3.5 h-3.5" />
                        )}
                      </button>
                    </div>

                    <div className="flex-1 min-w-0 flex flex-col sm:flex-row sm:items-center gap-2">
                      <div className="flex items-center gap-2 min-w-0">
                        <span className="text-[13px] font-semibold text-[var(--text-primary)] truncate">
                          {acc.name}
                        </span>
                      </div>

                      <div className="flex items-center flex-wrap gap-1.5">
                        <StatusBadge variant={getStateVariant(acc)}>
                          {getStateLabel(acc)}
                        </StatusBadge>

                        <span className="text-[10.5px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-muted)]">
                          API Key
                        </span>

                        {acc.proxy_pool_id && (
                          <span className="text-[10.5px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[var(--brand-subtle)] border border-[var(--brand-primary)]/30 text-[var(--brand-text)] flex items-center gap-1">
                            <Globe className="w-2.5 h-2.5" />
                            Proxy
                          </span>
                        )}

                        {hasError && (
                          <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[4px] bg-rose-950/20 border border-rose-600/30 text-rose-400 max-w-[180px] truncate">
                            {acc.last_error?.slice(0, 40)}…
                          </span>
                        )}

                        <span className="text-[10px] font-mono text-[var(--text-muted)]">
                          #{idx + 1}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-0.5 shrink-0">
                      <button
                        type="button"
                        title="Set proxy for this connection"
                        onClick={() => openEdit(acc)}
                        className="px-2 py-1.5 text-[10.5px] font-medium text-[var(--text-muted)] hover:text-[var(--status-warning)] hover:bg-[var(--bg-panel)] rounded-[5px] transition-colors flex items-center gap-1"
                      >
                        <Globe className="w-3.5 h-3.5" />
                        <span className="hidden sm:block">Proxy</span>
                      </button>

                      <button
                        type="button"
                        title="Edit connection"
                        onClick={() => openEdit(acc)}
                        className="px-2 py-1.5 text-[10.5px] font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] rounded-[5px] transition-colors flex items-center gap-1"
                      >
                        <Pencil className="w-3.5 h-3.5" />
                        <span className="hidden sm:block">Edit</span>
                      </button>

                      <InlineConfirm
                        trigger={
                          <button
                            type="button"
                            title="Delete connection"
                            className="px-2 py-1.5 text-[10.5px] font-medium text-[var(--text-muted)] hover:text-[var(--status-danger)] hover:bg-rose-950/20 rounded-[5px] transition-colors flex items-center gap-1"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                            <span className="hidden sm:block">Delete</span>
                          </button>
                        }
                        confirmText="Delete?"
                        onConfirm={() => handleDelete(acc.id)}
                      />

                      <button
                        type="button"
                        title={acc.enabled ? 'Disable' : 'Enable'}
                        onClick={() => handleToggle(acc)}
                        className="px-2 py-1.5 rounded-[5px] transition-colors text-[var(--text-muted)] hover:bg-[var(--bg-panel)]"
                      >
                        {acc.enabled ? (
                          <ToggleRight className="w-4.5 h-4.5 text-[var(--status-success)]" />
                        ) : (
                          <ToggleLeft className="w-4.5 h-4.5" />
                        )}
                      </button>
                    </div>
                  </div>

                  {isExpanded && (
                    <div className="px-10 pb-3 pt-1 grid grid-cols-1 sm:grid-cols-3 gap-3 bg-[var(--bg-panel)]/30 border-t border-[var(--border-subtle)]/50">
                      <div>
                        <span className="text-[10px] font-semibold uppercase text-[var(--text-muted)] block mb-0.5">
                          Masked Key
                        </span>
                        <span className="text-[11.5px] font-mono text-[var(--text-secondary)]">
                          {acc.masked_secret || '••••••••••••'}
                        </span>
                      </div>
                      <div>
                        <span className="text-[10px] font-semibold uppercase text-[var(--text-muted)] block mb-0.5">
                          Priority
                        </span>
                        <span className="text-[12px] font-mono text-[var(--brand-text)] font-semibold">
                          #{acc.priority}
                        </span>
                      </div>
                      <div>
                        <span className="text-[10px] font-semibold uppercase text-[var(--text-muted)] block mb-0.5">
                          Proxy
                        </span>
                        {proxy ? (
                          <a
                            href={`${proxy.scheme}://${proxy.host}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-[11.5px] font-mono text-[var(--brand-text)] hover:underline truncate block max-w-[220px]"
                          >
                            {proxy.scheme}://{proxy.host}
                          </a>
                        ) : acc.proxy_url ? (
                          <span className="text-[11.5px] font-mono text-[var(--text-secondary)] truncate block max-w-[220px]">
                            {acc.proxy_url}
                          </span>
                        ) : (
                          <span className="text-[11.5px] text-[var(--text-muted)]">Direct (no proxy)</span>
                        )}
                      </div>
                      {acc.last_error && (
                        <div className="col-span-full">
                          <span className="text-[10px] font-semibold uppercase text-rose-400 block mb-0.5">
                            Last Error
                          </span>
                          <span className="text-[11px] font-mono text-rose-300/80 break-all">{acc.last_error}</span>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </div>

      {platProvider?.capabilities && platProvider.capabilities.length > 0 && (
        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] p-4">
          <h4 className="text-[12px] font-semibold text-[var(--text-primary)] uppercase tracking-wider mb-3">
            Capabilities
          </h4>
          <div className="flex flex-wrap gap-1.5">
            {platProvider.capabilities.map((cap) => (
              <span
                key={cap}
                className="px-2.5 py-1 text-[11px] font-mono rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-secondary)]"
              >
                {cap}
              </span>
            ))}
          </div>
        </div>
      )}

      <BottomSheet
        isOpen={showAddSheet}
        onClose={() => setShowAddSheet(false)}
        title={`Add Connection — ${provider?.name || ''}`}
        description="API key or credential will be encrypted with AES-256-GCM"
        maxWidth="max-w-2xl"
      >
        <form onSubmit={handleAddSubmit} className="space-y-4">
          {formError && (
            <div className="flex items-center gap-2 p-3 rounded-[6px] bg-rose-950/20 border border-rose-600/30 text-rose-300 text-[12px]">
              <AlertCircle className="w-4 h-4 shrink-0" />
              {formError}
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                Connection Name *
              </label>
              <input
                type="text"
                required
                value={addForm.name}
                onChange={(e) => setAddForm({ ...addForm, name: e.target.value })}
                placeholder="e.g. Key 52"
                className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
              />
            </div>

            <div>
              <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                Priority (higher = more preferred)
              </label>
              <input
                type="number"
                min="1"
                max="100"
                value={addForm.priority}
                onChange={(e) => setAddForm({ ...addForm, priority: parseInt(e.target.value) || 1 })}
                className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
              />
            </div>
          </div>

          <div>
            <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
              API Key / Access Token *
            </label>
            <input
              type="password"
              required
              autoComplete="new-password"
              value={addForm.api_key}
              onChange={(e) => setAddForm({ ...addForm, api_key: e.target.value })}
              placeholder="sk-... or AIza..."
              className="w-full h-9 px-3 text-[13px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
            />
          </div>

          <div>
            <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
              Proxy (optional)
            </label>
            <select
              value={addForm.proxy_pool_id}
              onChange={(e) => setAddForm({ ...addForm, proxy_pool_id: e.target.value })}
              className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
            >
              <option value="">Direct (no proxy)</option>
              {proxies.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} — {p.scheme}://{p.host}
                </option>
              ))}
            </select>
          </div>

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-[var(--border-subtle)]">
            <Button
              type="button"
              variant="ghost"
              size="compact"
              onClick={() => setShowAddSheet(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isSubmitting}
              leftIcon={<Shield className="w-3.5 h-3.5" />}
            >
              Save Connection
            </Button>
          </div>
        </form>
      </BottomSheet>

      <BottomSheet
        isOpen={showEditSheet}
        onClose={() => setShowEditSheet(false)}
        title={`Edit — ${editingAccount?.name || ''}`}
        description="Update connection name, priority, proxy, or rotate API key"
        maxWidth="max-w-2xl"
      >
        <form onSubmit={handleEditSubmit} className="space-y-4">
          {formError && (
            <div className="flex items-center gap-2 p-3 rounded-[6px] bg-rose-950/20 border border-rose-600/30 text-rose-300 text-[12px]">
              <AlertCircle className="w-4 h-4 shrink-0" />
              {formError}
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                Connection Name
              </label>
              <input
                type="text"
                value={editForm.name}
                onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
              />
            </div>

            <div>
              <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                Priority
              </label>
              <input
                type="number"
                min="1"
                max="100"
                value={editForm.priority}
                onChange={(e) => setEditForm({ ...editForm, priority: parseInt(e.target.value) || 1 })}
                className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
              />
            </div>
          </div>

          <div>
            <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
              Proxy
            </label>
            <select
              value={editForm.proxy_pool_id}
              onChange={(e) => setEditForm({ ...editForm, proxy_pool_id: e.target.value })}
              className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
            >
              <option value="">Direct (no proxy)</option>
              {proxies.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} — {p.scheme}://{p.host}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
              Rotate API Key (leave blank to keep current)
            </label>
            <input
              type="password"
              autoComplete="new-password"
              value={editForm.api_key}
              onChange={(e) => setEditForm({ ...editForm, api_key: e.target.value })}
              placeholder="Enter new key to replace existing…"
              className="w-full h-9 px-3 text-[13px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
            />
          </div>

          <div className="flex items-center gap-3 p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
            <label className="text-[12.5px] font-medium text-[var(--text-secondary)] flex-1">
              Connection Enabled
            </label>
            <button
              type="button"
              onClick={() => setEditForm({ ...editForm, enabled: !editForm.enabled })}
              className="transition-colors"
            >
              {editForm.enabled ? (
                <ToggleRight className="w-6 h-6 text-[var(--status-success)]" />
              ) : (
                <ToggleLeft className="w-6 h-6 text-[var(--text-muted)]" />
              )}
            </button>
          </div>

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-[var(--border-subtle)]">
            <Button
              type="button"
              variant="ghost"
              size="compact"
              onClick={() => setShowEditSheet(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isSubmitting}
            >
              Save Changes
            </Button>
          </div>
        </form>
      </BottomSheet>
    </div>
  )
}
