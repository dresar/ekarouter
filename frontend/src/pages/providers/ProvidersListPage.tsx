import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Plus,
  Search,
  RefreshCw,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
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
  connections_count: number
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
  const navigate = useNavigate()
  const [providers, setProviders] = useState<UnifiedProvider[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')

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

      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <div className="relative flex-1 min-w-0 max-w-xs">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-[var(--text-muted)]" />
          <input
            type="search"
            placeholder="Cari"
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

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
        {filteredProviders.map((p) => {
          const isConnected = p.connections_count > 0

          return (
            <button
              key={p.id}
              type="button"
              onClick={() => navigate(`/providers/${p.id}`)}
              className="group relative flex items-center justify-between p-3 rounded-[8px] bg-[var(--bg-card)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] hover:bg-[var(--bg-panel)]/50 transition-all cursor-pointer text-left w-full"
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
            </button>
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
    </div>
  )
}
