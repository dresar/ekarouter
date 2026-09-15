import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Plus,
  Search,
  RefreshCw,
  Play,
  CheckCircle2,
  ExternalLink,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { Provider, PlatformProvider, Account, VaultCredential } from '../../types/api.ts'
import { getProviderApiKeyUrl } from '../../utils/providerUrls.ts'

interface UnifiedProvider {
  id: string
  key: string
  name: string
  category: 'oauth' | 'free_tier' | 'apikey' | 'developer' | 'cloud' | 'tools'
  kind: string
  base_url: string
  doc_url?: string
  description?: string
  enabled: boolean
  free_tier_status?: string
  capabilities?: string[]
  connections_count: number
}

const CATEGORIES = [
  { id: 'all', label: 'All' },
  { id: 'oauth', label: 'OAuth' },
  { id: 'free_tier', label: 'Free Tier' },
  { id: 'apikey', label: 'API Key' },
  { id: 'cloud', label: 'Cloud' },
  { id: 'tools', label: 'Tools' },
]

export function ProvidersListPage() {
  const navigate = useNavigate()
  const [providers, setProviders] = useState<UnifiedProvider[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [testingCategory, setTestingCategory] = useState<string | null>(null)
  const [testMessage, setTestMessage] = useState<string | null>(null)

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

      const isOAuthProvider = (id: string, cat?: string) => {
        const oauthIds = [
          'antigravity', 'gemini-agy', 'claude', 'qoder', 'codex',
          'github-copilot', 'github', 'cursor', 'kilocode', 'cline',
          'clinepass', 'codebuddy-intl', 'codebuddy-cn', 'kimi',
          'grok-cli', 'xai', 'xiaomi-mimo', 'zed', 'windsurf', 'trae',
        ]
        return cat === 'oauth' || oauthIds.includes(id)
      }

      const isFreeTierProvider = (id: string, cat?: string) => {
        const freeIds = [
          'opencode', 'gemini-cli', 'kiro', 'openrouter', 'nvidia',
          'ollama', 'vertex', 'gemini', 'poolside', 'byteplus',
          'kimchi', 'api-airforce', 'bazaarlink', 'kilo-gateway',
          'mimo-free', 'mmf', 'devin-cli',
        ]
        return cat === 'free_tier' || cat === 'free' || freeIds.includes(id)
      }

      for (const p of platformProviders) {
        let cat: UnifiedProvider['category'] = 'apikey'
        if (isOAuthProvider(p.id, p.category)) {
          cat = 'oauth'
        } else if (isFreeTierProvider(p.id, p.category)) {
          cat = 'free_tier'
        } else if (p.category === 'cloud' || p.category === 'developer' || p.category === 'storage') {
          cat = 'cloud'
        } else if (p.category === 'tools' || p.category === 'security' || p.category === 'automation') {
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
          connections_count: relatedAccounts.length + relatedCredentials.length,
        })
      }

      for (const p of adminProviders) {
        const relatedAccounts = accounts.filter((a) => a.provider_id === p.id)
        const relatedCredentials = credentials.filter((c) => c.provider_id === p.id)

        if (unifiedMap.has(p.id)) {
          const existing = unifiedMap.get(p.id)!
          existing.enabled = p.enabled
          existing.base_url = p.base_url || existing.base_url
          existing.connections_count = Math.max(existing.connections_count, relatedAccounts.length + relatedCredentials.length)
        } else {
          let cat: UnifiedProvider['category'] = 'apikey'
          if (isOAuthProvider(p.id, p.kind)) {
            cat = 'oauth'
          } else if (isFreeTierProvider(p.id, p.kind)) {
            cat = 'free_tier'
          }

          unifiedMap.set(p.id, {
            id: p.id,
            key: p.key || p.id,
            name: p.name || p.id,
            category: cat,
            kind: p.kind || 'openai',
            base_url: p.base_url,
            enabled: p.enabled,
            connections_count: relatedAccounts.length + relatedCredentials.length,
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

  const handleTestAll = async (categoryName: string) => {
    setTestingCategory(categoryName)
    setTestMessage(null)
    try {
      const res = await api.post<{ message?: string; tested?: number; healthy?: number }>('/api/accounts/test-all', {})
      const healthy = res?.healthy ?? 0
      const tested = res?.tested ?? 0
      setTestMessage(`Batch health check complete: ${healthy}/${tested} healthy accounts`)
      await fetchData()
      setTimeout(() => setTestMessage(null), 4000)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Test failed'
      setError(msg)
    } finally {
      setTestingCategory(null)
    }
  }

  const matchesSearch = (p: UnifiedProvider) => {
    if (!search.trim()) return true
    const q = search.toLowerCase()
    return (
      p.name.toLowerCase().includes(q) ||
      p.id.toLowerCase().includes(q) ||
      p.base_url.toLowerCase().includes(q)
    )
  }

  const oauthProviders = providers.filter((p) => p.category === 'oauth' && matchesSearch(p))
  const freeTierProviders = providers.filter((p) => p.category === 'free_tier' && matchesSearch(p))
  const apiKeyProviders = providers.filter(
    (p) => (p.category === 'apikey' || p.category === 'developer' || p.category === 'cloud' || p.category === 'tools') && matchesSearch(p)
  )

  const totalCount = providers.length
  const connectedCount = providers.filter((p) => p.connections_count > 0).length
  const totalConnections = providers.reduce((acc, p) => acc + p.connections_count, 0)

  const renderCard = (p: UnifiedProvider) => {
    const isConnected = p.connections_count > 0
    const isReady = p.free_tier_status === 'available' && p.connections_count === 0 && (p.id === 'opencode' || p.id.includes('free'))

    return (
      <button
        key={p.id}
        type="button"
        onClick={() => navigate(`/providers/${p.id}`)}
        className="group relative flex items-center justify-between p-3.5 rounded-[10px] bg-[var(--bg-card)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] hover:bg-[var(--bg-panel)]/50 transition-all cursor-pointer text-left w-full shadow-xs"
      >
        <div className="flex items-center gap-3 min-w-0">
          <ProviderLogo providerId={p.id} name={p.name} size="md" />

          <div className="flex flex-col min-w-0">
            <span className="text-[13px] font-semibold text-[var(--text-primary)] truncate leading-tight group-hover:text-[var(--brand-text)] transition-colors">
              {p.name}
            </span>

            <div className="flex items-center gap-1.5 mt-1">
              {!p.enabled ? (
                <div className="flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-slate-500" />
                  <span className="text-[11px] font-medium text-[var(--text-muted)]">
                    Disabled
                  </span>
                </div>
              ) : isConnected ? (
                <div className="flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-[var(--status-success)] shadow-xs" />
                  <span className="text-[11px] font-medium text-[var(--status-success)]">
                    {p.connections_count} Connected
                  </span>
                </div>
              ) : isReady ? (
                <div className="flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                  <span className="text-[11px] font-medium text-emerald-400">
                    Ready
                  </span>
                </div>
              ) : (
                <span className="text-[11px] font-medium text-[var(--text-muted)]">
                  No connections
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-1.5 shrink-0 pl-2">
          <a
            href={getProviderApiKeyUrl(p.id, p.doc_url)}
            target="_blank"
            rel="noopener noreferrer"
            onClick={(e) => e.stopPropagation()}
            title="Get API Key"
            className="p-1.5 rounded-[5px] text-[#ea580c] hover:text-[#f97316] hover:bg-orange-500/10 transition-colors cursor-pointer inline-flex items-center gap-1 text-[11px] font-medium"
          >
            <ExternalLink className="w-3.5 h-3.5" />
            <span className="hidden sm:inline">Get Key</span>
          </a>
        </div>
      </button>
    )
  }

  return (
    <div className="space-y-6">
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
            <Button
              variant="primary"
              size="compact"
              onClick={() => navigate('/providers/new')}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
            >
              Add Provider
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchData} />}

      {testMessage && (
        <div className="flex items-center gap-2 px-3 py-2 rounded-[6px] bg-[var(--status-success-bg)] border border-[var(--status-success)] text-[var(--status-success)] text-[12px]">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{testMessage}</span>
        </div>
      )}

      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <div className="relative flex-1 min-w-0 max-w-xs">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-[var(--text-muted)]" />
          <input
            type="search"
            placeholder="Cari provider..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full h-8 pl-8 pr-3 text-[12.5px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
          />
        </div>

        <div className="flex items-center gap-1.5 flex-wrap">
          {CATEGORIES.map((cat) => (
            <button
              key={cat.id}
              type="button"
              onClick={() => setSelectedCategory(cat.id)}
              className={`h-7 px-3 text-[11.5px] font-medium rounded-[5px] transition-colors cursor-pointer ${
                selectedCategory === cat.id
                  ? 'bg-[#282a34] text-[#f3f4f6] border border-[#3e4354]'
                  : 'bg-[var(--bg-panel)] text-[var(--text-secondary)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] hover:text-[var(--text-primary)]'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>
      </div>

      {(selectedCategory === 'all' || selectedCategory === 'oauth') && oauthProviders.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-[15px] font-semibold text-[var(--text-primary)] tracking-tight">
              OAuth Providers
            </h2>
            <Button
              variant="secondary"
              size="compact"
              onClick={() => handleTestAll('oauth')}
              isLoading={testingCategory === 'oauth'}
              leftIcon={<Play className="w-3 h-3" />}
            >
              Test All
            </Button>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
            {oauthProviders.map(renderCard)}
          </div>
        </div>
      )}

      {(selectedCategory === 'all' || selectedCategory === 'free_tier') && freeTierProviders.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-[15px] font-semibold text-[var(--text-primary)] tracking-tight">
              Free Tier Providers
            </h2>
            <Button
              variant="secondary"
              size="compact"
              onClick={() => handleTestAll('free')}
              isLoading={testingCategory === 'free'}
              leftIcon={<Play className="w-3 h-3" />}
            >
              Test All
            </Button>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
            {freeTierProviders.map(renderCard)}
          </div>
        </div>
      )}

      {(selectedCategory === 'all' || selectedCategory === 'apikey' || selectedCategory === 'cloud' || selectedCategory === 'tools') && apiKeyProviders.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-[15px] font-semibold text-[var(--text-primary)] tracking-tight">
              API Key Providers
            </h2>
            <Button
              variant="secondary"
              size="compact"
              onClick={() => handleTestAll('apikey')}
              isLoading={testingCategory === 'apikey'}
              leftIcon={<Play className="w-3 h-3" />}
            >
              Test All
            </Button>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
            {apiKeyProviders.map(renderCard)}
          </div>
        </div>
      )}

      {oauthProviders.length === 0 && freeTierProviders.length === 0 && apiKeyProviders.length === 0 && !isLoading && (
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
    </div>
  )
}
