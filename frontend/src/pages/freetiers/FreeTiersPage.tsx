import { useEffect, useState } from 'react'
import {
  ExternalLink,
  RefreshCw,
  Filter,
  CreditCard,
  Search,
  Server,
  Zap,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { FreeTierItem } from '../../types/api.ts'

export function FreeTiersPage() {
  const [items, setItems] = useState<FreeTierItem[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [verifiedOnly, setVerifiedOnly] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [lastUpdated, setLastUpdated] = useState<string>('-')
  const [error, setError] = useState<string | null>(null)

  const loadData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [tiersRes, catsRes] = await Promise.all([
        api.get<any>('/api/free-tiers').catch(() => []),
        api.get<any>('/api/free-tiers/categories').catch(() => []),
      ])

      const parsedItems: FreeTierItem[] = Array.isArray(tiersRes)
        ? tiersRes
        : tiersRes && Array.isArray(tiersRes.entries)
        ? tiersRes.entries
        : tiersRes && Array.isArray(tiersRes.data)
        ? tiersRes.data
        : []

      const parsedCats: string[] = Array.isArray(catsRes)
        ? catsRes
        : catsRes && Array.isArray(catsRes.categories)
        ? catsRes.categories
        : []

      setItems(parsedItems)
      setCategories(parsedCats)
      setLastUpdated(
        new Date().toLocaleTimeString('en-US', {
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
        })
      )
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load free tier catalog'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      await api.post('/api/free-tiers/refresh')
      await loadData()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Refresh failed'
      setError(msg)
    } finally {
      setIsRefreshing(false)
    }
  }

  const handleClearFilters = () => {
    setSelectedCategory('all')
    setVerifiedOnly(false)
    setSearchQuery('')
  }

  const safeItems = Array.isArray(items) ? items : []
  const filteredItems = safeItems.filter((item) => {
    const matchesCategory = selectedCategory === 'all' || item.category === selectedCategory
    const isVerified = Boolean(item.verified || item.status === 'verified')
    const matchesVerified = !verifiedOnly || isVerified
    const pName = item.provider_name || ''
    const mName = item.model_name || item.free_tier_status || ''
    const qDesc = item.free_quota_description || item.free_quota || ''
    const matchesSearch =
      !searchQuery ||
      pName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      mName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      qDesc.toLowerCase().includes(searchQuery.toLowerCase())
    return matchesCategory && matchesVerified && matchesSearch
  })

  return (
    <div className="space-y-4">
      <PageHeader
        title="Free Tiers"
        description="Explore zero-cost AI provider tiers and quota limits."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Free Tiers' },
        ]}
        metadata={
          <span className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-[var(--text-muted)]" />
            <span>Last updated: {lastUpdated}</span>
          </span>
        }
        actions={
          <Button
            variant="primary"
            size="compact"
            onClick={handleRefresh}
            isLoading={isRefreshing || isLoading}
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
          >
            Refresh Catalog
          </Button>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadData} />}

      <div className="flex flex-col md:flex-row items-center justify-between gap-3 bg-[var(--bg-surface)] p-2.5 rounded-[8px] border border-[var(--border-subtle)]">
        <div className="flex items-center gap-4 w-full md:w-auto flex-wrap">
          <div className="flex items-center gap-2">
            <Filter className="w-4 h-4 text-[var(--brand-primary)]" />
            <span className="text-[12px] font-medium text-[var(--text-primary)]">Category</span>
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              className="px-2.5 py-1 text-[12px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
            >
              <option value="all">All Categories</option>
              {categories.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </div>

          <label className="flex items-center gap-2 cursor-pointer select-none text-[12px] text-[var(--text-secondary)]">
            <input
              type="checkbox"
              checked={verifiedOnly}
              onChange={(e) => setVerifiedOnly(e.target.checked)}
              className="rounded w-3.5 h-3.5 text-[var(--brand-primary)] accent-blue-600"
            />
            <span>Verified zero-cost tiers only</span>
          </label>
        </div>

        <div className="relative w-full md:w-72">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)] pointer-events-none" />
          <input
            type="text"
            placeholder="Search providers or models..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-8 pr-3 py-1.5 text-[12px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)]"
          />
        </div>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="h-32 rounded-[8px] bg-[var(--bg-card)] border border-[var(--border-subtle)] p-4 animate-pulse"
            />
          ))}
        </div>
      ) : filteredItems.length === 0 ? (
        <div className="py-20 px-4 rounded-[8px] bg-[var(--bg-card)] border border-[var(--border-subtle)] flex flex-col items-center justify-center text-center">
          <div className="relative mb-4 flex items-center justify-center">
            <div className="w-16 h-16 rounded-xl bg-[var(--bg-panel)] border border-[var(--border-strong)] flex items-center justify-center text-[var(--text-muted)]">
              <Server className="w-8 h-8 text-[var(--text-muted)] opacity-60" />
            </div>
            <div className="absolute -bottom-1 -right-1 w-7 h-7 rounded-full bg-[var(--bg-surface)] border border-[var(--border-strong)] flex items-center justify-center">
              <Search className="w-3.5 h-3.5 text-[var(--brand-text)]" />
            </div>
          </div>

          <h3 className="text-[15px] font-semibold text-[var(--text-primary)]">
            No free tier offerings found
          </h3>
          <p className="text-[12.5px] text-[var(--text-muted)] max-w-md mt-1.5 leading-relaxed">
            We couldn't find any verified free tier offerings matching your current filter. Try adjusting your filters or refresh the catalog to fetch the latest data.
          </p>

          <div className="flex items-center gap-2.5 mt-5">
            <Button
              variant="primary"
              size="compact"
              onClick={handleRefresh}
              isLoading={isRefreshing}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh Catalog
            </Button>
            <Button
              variant="secondary"
              size="compact"
              onClick={handleClearFilters}
              leftIcon={<Filter className="w-3.5 h-3.5" />}
            >
              Clear Filters
            </Button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {filteredItems.map((tier) => (
            <div
              key={tier.id}
              className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3.5 flex flex-col justify-between transition-all hover:border-[var(--brand-primary)]/40 hover:bg-[var(--bg-panel)]/40"
            >
              <div>
                <div className="flex items-center justify-between gap-2 mb-2">
                  <div className="flex items-center gap-2 min-w-0">
                    <ProviderLogo providerId={tier.provider_name} name={tier.provider_name} size="sm" />
                    <span className="font-semibold text-[13px] text-[var(--text-primary)] truncate">
                      {tier.provider_name}
                    </span>
                  </div>
                  {(tier.verified || tier.status === 'verified') && (
                    <StatusBadge variant="healthy">
                      Verified
                    </StatusBadge>
                  )}
                </div>

                <div className="font-mono text-[11.5px] text-[var(--brand-text)] font-semibold mb-1.5 flex items-center gap-1">
                  <Zap className="w-3 h-3 text-sky-400 shrink-0" />
                  <span className="truncate">{tier.model_name || tier.free_tier_status?.replace(/_/g, ' ') || 'Free Tier Available'}</span>
                </div>

                <p className="text-[11.5px] text-[var(--text-secondary)] mb-3 leading-relaxed line-clamp-3">
                  {tier.free_quota_description || tier.free_quota || 'Active community free tier offering'}
                </p>
              </div>

              <div className="pt-2.5 border-t border-[var(--border-subtle)] flex items-center justify-between text-[11px]">
                <div className="flex items-center gap-2 text-[var(--text-muted)] font-mono">
                  {(tier.requires_credit_card || tier.payment_required) && (
                    <span className="inline-flex items-center gap-1 text-amber-400">
                      <CreditCard className="w-3 h-3" />
                      <span>CC Req</span>
                    </span>
                  )}
                  {tier.rpm_limit && (
                    <span>{tier.rpm_limit} RPM</span>
                  )}
                  {tier.tpm_limit && (
                    <span>&bull; {tier.tpm_limit} TPM</span>
                  )}
                </div>

                {(tier.documentation_url || tier.docs_url || tier.official_website) && (
                  <a
                    href={tier.documentation_url || tier.docs_url || tier.official_website}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 font-medium"
                  >
                    <span>Docs</span>
                    <ExternalLink className="w-3 h-3" />
                  </a>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
