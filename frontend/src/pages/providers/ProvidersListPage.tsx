import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Plus,
  Search,
  RefreshCw,
  ExternalLink,
  CheckCircle2,
  AlertCircle,
  Clock,
  Shield,
  Trash2,
  Radio,
  Sliders,
  KeyRound,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { RightDrawer } from '../../components/ui/RightDrawer.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { api } from '../../api/client.ts'
import { Provider, PlatformProvider, Account, VaultCredential } from '../../types/api.ts'

interface UnifiedProvider {
  id: string
  key: string
  name: string
  category: string
  kind: string
  base_url: string
  doc_url?: string
  description?: string
  enabled: boolean
  free_tier_status?: string
  capabilities?: string[]
  supported_operations?: string[]
  connections_count: number
  accounts: Account[]
  credentials: VaultCredential[]
}

const CATEGORIES = [
  { id: 'all', label: 'All' },
  { id: 'ai', label: 'AI' },
  { id: 'coding', label: 'Coding' },
  { id: 'oauth', label: 'OAuth' },
  { id: 'free_tier', label: 'Free Tier' },
  { id: 'tools', label: 'Tools' },
  { id: 'cloud', label: 'Cloud' },
]

export function ProvidersListPage() {
  const [providers, setProviders] = useState<UnifiedProvider[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [drawerError, setDrawerError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [selectedProvider, setSelectedProvider] = useState<UnifiedProvider | null>(null)
  const [testResult, setTestResult] = useState<{ status: 'idle' | 'testing' | 'success' | 'error'; message?: string; latency?: number }>({ status: 'idle' })

  // Quick add connection form state in drawer
  const [showAddForm, setShowAddForm] = useState(false)
  const [newAccountName, setNewAccountName] = useState('')
  const [newApiKey, setNewApiKey] = useState('')
  const [isSubmittingAccount, setIsSubmittingAccount] = useState(false)

  const fetchData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [adminProvidersRes, platformProvidersRes, accountsRes, credentialsRes] = await Promise.allSettled([
        api.get<Provider[]>('/api/providers'),
        api.get<PlatformProvider[]>('/api/v1/providers'),
        api.get<Account[]>('/api/accounts'),
        api.get<VaultCredential[]>('/api/v1/credentials'),
      ])

      const adminProviders: Provider[] = adminProvidersRes.status === 'fulfilled' && Array.isArray(adminProvidersRes.value) ? adminProvidersRes.value : []
      const platformProviders: PlatformProvider[] = platformProvidersRes.status === 'fulfilled' && Array.isArray(platformProvidersRes.value) ? platformProvidersRes.value : []
      const accounts: Account[] = accountsRes.status === 'fulfilled' && Array.isArray(accountsRes.value) ? accountsRes.value : []
      const credentials: VaultCredential[] = credentialsRes.status === 'fulfilled' && Array.isArray(credentialsRes.value) ? credentialsRes.value : []

      const unifiedMap = new Map<string, UnifiedProvider>()

      // 1. Process platform registry
      for (const p of platformProviders) {
        let cat = (p.category || 'ai').toLowerCase()
        if (p.id.includes('github') || p.id.includes('code') || p.id.includes('cline') || p.id.includes('devin')) {
          cat = 'coding'
        } else if (p.free_tier_status === 'available' || p.id.includes('free')) {
          cat = 'free_tier'
        } else if (cat === 'developer' || cat === 'cloud' || cat === 'storage') {
          cat = 'cloud'
        } else if (cat === 'tool' || cat === 'tools') {
          cat = 'tools'
        }

        const relatedAccounts = accounts.filter((a) => a.provider_id === p.id)
        const relatedCredentials = credentials.filter((c) => c.provider_id === p.id)

        unifiedMap.set(p.id, {
          id: p.id,
          key: p.id,
          name: p.name,
          category: cat,
          kind: p.auth_type || 'bearer',
          base_url: p.base_url,
          doc_url: p.doc_url || p.website_url,
          description: p.description,
          enabled: p.enabled ?? true,
          free_tier_status: p.free_tier_status,
          capabilities: p.capabilities || [],
          supported_operations: p.supported_operations || [],
          connections_count: relatedAccounts.length + relatedCredentials.length,
          accounts: relatedAccounts,
          credentials: relatedCredentials,
        })
      }

      // 2. Merge admin configured providers
      for (const p of adminProviders) {
        const relatedAccounts = accounts.filter((a) => a.provider_id === p.id)
        const relatedCredentials = credentials.filter((c) => c.provider_id === p.id)

        if (unifiedMap.has(p.id)) {
          const existing = unifiedMap.get(p.id)!
          existing.enabled = p.enabled
          existing.kind = p.kind || existing.kind
          existing.base_url = p.base_url || existing.base_url
          existing.connections_count = Math.max(existing.connections_count, relatedAccounts.length + relatedCredentials.length)
          existing.accounts = relatedAccounts
          existing.credentials = relatedCredentials
        } else {
          let cat = 'ai'
          if (p.id.includes('github') || p.kind.includes('code')) cat = 'coding'
          if (p.id.includes('free')) cat = 'free_tier'

          unifiedMap.set(p.id, {
            id: p.id,
            key: p.key || p.id,
            name: p.name || p.id,
            category: cat,
            kind: p.kind || 'openai',
            base_url: p.base_url,
            doc_url: undefined,
            description: undefined,
            enabled: p.enabled,
            connections_count: relatedAccounts.length + relatedCredentials.length,
            accounts: relatedAccounts,
            credentials: relatedCredentials,
          })
        }
      }

      setProviders(Array.from(unifiedMap.values()))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load provider catalog'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const handleTestProvider = async (provider: UnifiedProvider) => {
    setTestResult({ status: 'testing' })
    const start = performance.now()
    try {
      await api.post(`/api/v1/providers/${provider.id}/health`, {})
      const latency = Math.round(performance.now() - start)
      setTestResult({
        status: 'success',
        latency,
        message: 'Connection successful',
      })
    } catch {
      const latency = Math.round(performance.now() - start)
      setTestResult({
        status: 'error',
        latency,
        message: 'Connection timed out or endpoint unreachable',
      })
    }
  }

  const handleAddAccount = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedProvider || !newAccountName.trim() || !newApiKey.trim()) return

    setIsSubmittingAccount(true)
    setDrawerError(null)
    try {
      await api.post('/api/accounts', {
        provider_id: selectedProvider.id,
        name: newAccountName.trim(),
        auth_type: 'apiKey',
        priority: 1,
        api_key: newApiKey.trim(),
      })

      setShowAddForm(false)
      setNewAccountName('')
      setNewApiKey('')
      fetchData()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to save connection'
      setDrawerError(msg)
    } finally {
      setIsSubmittingAccount(false)
    }
  }

  const handleDeleteAccount = async (accountId: string) => {
    setDrawerError(null)
    try {
      await api.delete(`/api/accounts/${accountId}`)
      fetchData()
      if (selectedProvider) {
        setSelectedProvider({
          ...selectedProvider,
          accounts: selectedProvider.accounts.filter((a) => a.id !== accountId),
          connections_count: Math.max(0, selectedProvider.connections_count - 1),
        })
      }
    } catch (err: unknown) {
      setDrawerError(err instanceof Error ? err.message : 'Failed to delete connection')
    }
  }

  const handleDeleteCredential = async (credId: string) => {
    setDrawerError(null)
    try {
      await api.delete(`/api/v1/credentials/${credId}`)
      fetchData()
      if (selectedProvider) {
        setSelectedProvider({
          ...selectedProvider,
          credentials: selectedProvider.credentials.filter((c) => c.id !== credId),
          connections_count: Math.max(0, selectedProvider.connections_count - 1),
        })
      }
    } catch (err: unknown) {
      setDrawerError(err instanceof Error ? err.message : 'Failed to delete vault credential')
    }
  }

  const filteredProviders = providers.filter((p) => {
    const matchesSearch =
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.id.toLowerCase().includes(search.toLowerCase()) ||
      p.base_url.toLowerCase().includes(search.toLowerCase())
    const matchesCat = selectedCategory === 'all' || p.category === selectedCategory
    return matchesSearch && matchesCat
  })

  const totalCount = providers.length
  const connectedCount = providers.filter((p) => p.connections_count > 0).length
  const totalConnections = providers.reduce((acc, p) => acc + p.connections_count, 0)

  return (
    <div className="space-y-4">
      <PageHeader
        title="Providers"
        description="Manage AI provider connections."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Providers' },
        ]}
        metadata={
          <span>
            {connectedCount} connected ({totalConnections} accounts) &bull; {totalCount} available
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={fetchData}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Link to="/providers/new">
              <Button variant="primary" size="compact" leftIcon={<Plus className="w-3.5 h-3.5" />}>
                Add Provider
              </Button>
            </Link>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchData} />}

      <div className="flex flex-col sm:flex-row items-center justify-between gap-3 bg-[var(--bg-surface)] p-2.5 rounded-[8px] border border-[var(--border-subtle)]">
        <div className="flex items-center gap-1 overflow-x-auto w-full sm:w-auto pb-1 sm:pb-0 scrollbar-none">
          {CATEGORIES.map((cat) => (
            <button
              key={cat.id}
              type="button"
              onClick={() => setSelectedCategory(cat.id)}
              className={`px-3 py-1 text-[11.5px] font-medium rounded-[5px] transition-colors whitespace-nowrap ${
                selectedCategory === cat.id
                  ? 'bg-[var(--brand-primary)] text-white'
                  : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)]'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>

        <div className="relative w-full sm:w-64 shrink-0">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)] pointer-events-none" />
          <input
            type="text"
            placeholder="Search provider..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-8 pr-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)]"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
        {filteredProviders.map((p) => {
          const isConnected = p.connections_count > 0
          const isSelected = selectedProvider?.id === p.id

          return (
            <div
              key={p.id}
              onClick={() => {
                setSelectedProvider(p)
                setTestResult({ status: 'idle' })
                setShowAddForm(false)
              }}
              className={`group relative flex items-center justify-between p-3 rounded-[8px] bg-[var(--bg-card)] border transition-all cursor-pointer ${
                isSelected
                  ? 'border-[var(--brand-primary)] bg-[var(--brand-subtle)]/40 ring-1 ring-[var(--brand-primary)]'
                  : 'border-[var(--border-subtle)] hover:border-[var(--border-strong)] hover:bg-[var(--bg-panel)]/50'
              }`}
            >
              <div className="flex items-center gap-3 min-w-0">
                <ProviderLogo providerId={p.id} name={p.name} size="md" />

                <div className="flex flex-col min-w-0">
                  <div className="flex items-center gap-1.5">
                    <span className="text-[13px] font-semibold text-[var(--text-primary)] truncate leading-tight group-hover:text-[var(--brand-text)] transition-colors">
                      {p.name}
                    </span>
                  </div>

                  <div className="flex items-center gap-2 mt-1">
                    <span className="text-[10.5px] font-mono text-[var(--text-muted)] uppercase">
                      {p.category}
                    </span>
                    <span className="text-[var(--border-strong)]">&bull;</span>
                    <div className="flex items-center gap-1">
                      <span
                        className={`w-1.5 h-1.5 rounded-full ${
                          !p.enabled
                            ? 'bg-slate-500'
                            : isConnected
                            ? 'bg-[var(--status-success)]'
                            : 'bg-[var(--text-muted)]'
                        }`}
                      />
                      <span className="text-[11px] font-medium text-[var(--text-secondary)]">
                        {!p.enabled
                          ? 'Disabled'
                          : isConnected
                          ? `${p.connections_count} connected`
                          : 'No connections'}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="flex items-center shrink-0 ml-2">
                <button
                  type="button"
                  title="Configure provider"
                  className="p-1 rounded text-[var(--text-muted)] group-hover:text-[var(--text-primary)] transition-colors"
                >
                  <Sliders className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          )
        })}
      </div>

      {filteredProviders.length === 0 && !isLoading && (
        <div className="py-16 text-center rounded-[8px] bg-[var(--bg-card)] border border-[var(--border-subtle)]">
          <p className="text-[13px] font-medium text-[var(--text-primary)]">No providers found</p>
          <p className="text-[12px] text-[var(--text-muted)] mt-1">
            Try adjusting your search or category filter.
          </p>
          <Button
            variant="secondary"
            size="compact"
            onClick={() => {
              setSearch('')
              setSelectedCategory('all')
            }}
            className="mt-3"
          >
            Clear Filters
          </Button>
        </div>
      )}

      {selectedProvider && (
        <RightDrawer
          isOpen={Boolean(selectedProvider)}
          onClose={() => {
            setSelectedProvider(null)
            setDrawerError(null)
          }}
          title={selectedProvider.name}
          description="Provider Connection & Credential Details"
        >
          <div className="space-y-5">
            {drawerError && (
              <div className="p-2.5 rounded-[6px] bg-[var(--status-danger)]/10 border border-[var(--status-danger)]/30 text-[12px] text-[var(--status-danger)] flex items-center justify-between">
                <span>{drawerError}</span>
                <button
                  type="button"
                  onClick={() => setDrawerError(null)}
                  className="text-[var(--status-danger)] font-bold text-xs"
                >
                  ✕
                </button>
              </div>
            )}
            <div className="flex items-center justify-between p-3 rounded-[8px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
              <div className="flex items-center gap-3">
                <ProviderLogo providerId={selectedProvider.id} name={selectedProvider.name} size="lg" />
                <div>
                  <h4 className="text-[13.5px] font-semibold text-[var(--text-primary)]">
                    {selectedProvider.name}
                  </h4>
                  <p className="text-[11px] font-mono text-[var(--text-muted)] mt-0.5 truncate max-w-[240px]">
                    {selectedProvider.base_url}
                  </p>
                </div>
              </div>

              <StatusBadge variant={selectedProvider.enabled ? 'healthy' : 'disabled'}>
                {selectedProvider.enabled ? 'Active' : 'Disabled'}
              </StatusBadge>
            </div>

            {selectedProvider.description && (
              <div className="text-[12px] text-[var(--text-secondary)] leading-relaxed">
                {selectedProvider.description}
              </div>
            )}

            <div className="flex items-center gap-2">
              <Button
                variant="secondary"
                size="compact"
                onClick={() => handleTestProvider(selectedProvider)}
                isLoading={testResult.status === 'testing'}
                leftIcon={<Radio className="w-3.5 h-3.5" />}
                className="w-full"
              >
                Test Connection
              </Button>

              {selectedProvider.doc_url && (
                <a
                  href={selectedProvider.doc_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="shrink-0"
                >
                  <Button variant="secondary" size="compact" leftIcon={<ExternalLink className="w-3.5 h-3.5" />}>
                    Docs
                  </Button>
                </a>
              )}
            </div>

            {testResult.status !== 'idle' && (
              <div
                className={`p-3 rounded-[6px] text-[11.5px] flex items-center justify-between border ${
                  testResult.status === 'success'
                    ? 'bg-emerald-950/20 border-emerald-600/30 text-emerald-300'
                    : testResult.status === 'error'
                    ? 'bg-rose-950/20 border-rose-600/30 text-rose-300'
                    : 'bg-[var(--bg-panel)] border-[var(--border-subtle)] text-[var(--text-muted)]'
                }`}
              >
                <div className="flex items-center gap-2">
                  {testResult.status === 'success' ? (
                    <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                  ) : testResult.status === 'error' ? (
                    <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />
                  ) : (
                    <Clock className="w-4 h-4 text-blue-400 shrink-0 animate-spin" />
                  )}
                  <span>{testResult.message || 'Testing connectivity...'}</span>
                </div>
                {testResult.latency !== undefined && (
                  <span className="font-mono text-[10.5px]">{testResult.latency}ms</span>
                )}
              </div>
            )}

            <div>
              <div className="flex items-center justify-between mb-2">
                <h5 className="text-[12px] font-semibold text-[var(--text-primary)] uppercase tracking-wider">
                  Connected Accounts ({selectedProvider.accounts.length})
                </h5>
                <button
                  type="button"
                  onClick={() => setShowAddForm(!showAddForm)}
                  className="text-[11.5px] font-medium text-[var(--brand-text)] hover:underline flex items-center gap-1"
                >
                  <Plus className="w-3 h-3" />
                  <span>Add Connection</span>
                </button>
              </div>

              {showAddForm && (
                <form
                  onSubmit={handleAddAccount}
                  className="mb-3 p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] space-y-2.5"
                >
                  <div>
                    <label className="text-[11px] font-medium text-[var(--text-secondary)] block mb-1">
                      Account Label
                    </label>
                    <input
                      type="text"
                      placeholder="e.g. Production Key"
                      value={newAccountName}
                      onChange={(e) => setNewAccountName(e.target.value)}
                      required
                      className="w-full px-2.5 py-1.5 text-[12px] rounded-[5px]"
                    />
                  </div>
                  <div>
                    <label className="text-[11px] font-medium text-[var(--text-secondary)] block mb-1">
                      API Key / Secret Token
                    </label>
                    <input
                      type="password"
                      placeholder="sk-..."
                      value={newApiKey}
                      onChange={(e) => setNewApiKey(e.target.value)}
                      required
                      className="w-full px-2.5 py-1.5 text-[12px] rounded-[5px]"
                    />
                  </div>
                  <div className="flex items-center justify-end gap-2 pt-1">
                    <Button
                      type="button"
                      variant="secondary"
                      size="compact"
                      onClick={() => setShowAddForm(false)}
                    >
                      Cancel
                    </Button>
                    <Button
                      type="submit"
                      variant="primary"
                      size="compact"
                      isLoading={isSubmittingAccount}
                    >
                      Save Key
                    </Button>
                  </div>
                </form>
              )}

              <div className="space-y-2">
                {selectedProvider.accounts.map((acc) => (
                  <div
                    key={acc.id}
                    className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] flex items-center justify-between"
                  >
                    <div className="flex flex-col min-w-0">
                      <div className="flex items-center gap-2">
                        <Shield className="w-3.5 h-3.5 text-[var(--brand-text)] shrink-0" />
                        <span className="text-[12px] font-medium text-[var(--text-primary)] truncate">
                          {acc.name}
                        </span>
                        <span className="text-[10px] font-mono text-[var(--text-muted)] bg-[var(--bg-card)] px-1.5 py-0.5 rounded">
                          priority {acc.priority}
                        </span>
                      </div>
                      <span className="text-[10.5px] font-mono text-[var(--text-muted)] mt-1">
                        Fingerprint: {acc.id.substring(0, 8)}...
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      <StatusBadge variant={acc.enabled ? 'healthy' : 'disabled'}>
                        {acc.enabled ? 'Active' : 'Disabled'}
                      </StatusBadge>
                      <InlineConfirm
                        trigger={
                          <button
                            type="button"
                            title="Delete connection"
                            className="p-1 text-[var(--text-muted)] hover:text-[var(--status-danger)]"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        }
                        confirmText="Remove?"
                        onConfirm={() => handleDeleteAccount(acc.id)}
                      />
                    </div>
                  </div>
                ))}

                {selectedProvider.accounts.length === 0 && !showAddForm && (
                  <div className="p-3 text-center rounded-[6px] bg-[var(--bg-panel)]/30 border border-dashed border-[var(--border-subtle)] text-[11.5px] text-[var(--text-muted)]">
                    No active credentials configured for this provider.
                  </div>
                )}
              </div>
            </div>

            {selectedProvider.credentials.length > 0 && (
              <div>
                <h5 className="text-[12px] font-semibold text-[var(--text-primary)] uppercase tracking-wider mb-2">
                  Vault Credentials ({selectedProvider.credentials.length})
                </h5>
                <div className="space-y-2">
                  {selectedProvider.credentials.map((cred) => (
                    <div
                      key={cred.id}
                      className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] flex items-center justify-between"
                    >
                      <div className="flex flex-col min-w-0">
                        <div className="flex items-center gap-2">
                          <KeyRound className="w-3.5 h-3.5 text-[var(--brand-text)] shrink-0" />
                          <span className="text-[12px] font-medium text-[var(--text-primary)] truncate">
                            {cred.name}
                          </span>
                          <span className="text-[10px] font-mono text-[var(--text-muted)] bg-[var(--bg-card)] px-1.5 py-0.5 rounded">
                            {cred.environment}
                          </span>
                        </div>
                        <span className="text-[10.5px] font-mono text-[var(--text-muted)] mt-1">
                          {cred.masked_value}
                        </span>
                      </div>

                      <div className="flex items-center gap-2">
                        <StatusBadge
                          variant={
                            cred.status === 'active'
                              ? 'healthy'
                              : cred.status === 'cooling_down'
                              ? 'cooling_down'
                              : 'disabled'
                          }
                        >
                          {cred.status}
                        </StatusBadge>
                        <InlineConfirm
                          trigger={
                            <button
                              type="button"
                              title="Delete vault credential"
                              className="p-1 text-[var(--text-muted)] hover:text-[var(--status-danger)]"
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          }
                          confirmText="Remove?"
                          onConfirm={() => handleDeleteCredential(cred.id)}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {selectedProvider.capabilities && selectedProvider.capabilities.length > 0 && (
              <div>
                <h5 className="text-[12px] font-semibold text-[var(--text-primary)] uppercase tracking-wider mb-2">
                  Capabilities
                </h5>
                <div className="flex flex-wrap gap-1.5">
                  {selectedProvider.capabilities.map((cap) => (
                    <span
                      key={cap}
                      className="px-2 py-0.5 text-[10.5px] font-mono rounded-[4px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-secondary)]"
                    >
                      {cap}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </RightDrawer>
      )}
    </div>
  )
}
