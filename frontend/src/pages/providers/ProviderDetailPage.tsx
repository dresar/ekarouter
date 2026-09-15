import { useEffect, useState, useRef, FormEvent } from 'react'
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
  Bot,
  Copy,
  Check,
  Eye,
  Network,
  ArrowLeftRight,
  Unlink,
  RefreshCw,
  Ban,
  RotateCw,
  X,
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

interface ProviderModel {
  id: string
  external_name: string
  display_name: string
  vision?: boolean
  web?: boolean
  enabled: boolean
}

const DEFAULT_GEMINI_MODELS: Omit<ProviderModel, 'id' | 'enabled'>[] = [
  { external_name: 'gemini-3.6-flash', display_name: 'Gemini 3.6 Flash', vision: true, web: true },
  { external_name: 'gemini-3.8-flash', display_name: 'Gemini 3.8 Flash', vision: true, web: true },
  { external_name: 'gemini-3.7-flash', display_name: 'Gemini 3.7 Flash', vision: true, web: true },
  { external_name: 'gemini-3.5-flash-lite', display_name: 'Gemini 3.5 Flash Lite', vision: true, web: false },
  { external_name: 'gemini-3.1-pro-preview', display_name: 'Gemini 3.1 Pro Preview', vision: true, web: true },
  { external_name: 'gemini-3.1-flash-lite-preview', display_name: 'Gemini 3.1 Flash Lite Preview', vision: true, web: false },
  { external_name: 'gemini-3-flash-preview', display_name: 'Gemini 3 Flash Preview', vision: true, web: true },
  { external_name: 'gemini-2.5-pro', display_name: 'Gemini 2.5 Pro', vision: true, web: true },
  { external_name: 'gemini-2.5-flash', display_name: 'Gemini 2.5 Flash', vision: true, web: true },
  { external_name: 'gemini-2.5-flash-lite', display_name: 'Gemini 2.5 Flash Lite', vision: true, web: false },
  { external_name: 'gemma-4-31b-it', display_name: 'Gemma 4 31B IT', vision: false, web: false },
]

export function ProviderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [provider, setProvider] = useState<Provider | null>(null)
  const [platProvider, setPlatProvider] = useState<PlatformProvider | null>(null)
  const [accounts, setAccounts] = useState<Account[]>([])
  const [proxies, setProxies] = useState<ProxyProfile[]>([])
  const [models, setModels] = useState<ProviderModel[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionSuccess, setActionSuccess] = useState<string | null>(null)

  const [testResult, setTestResult] = useState<{
    status: 'idle' | 'testing' | 'success' | 'error'
    message?: string
    latency?: number
  }>({ status: 'idle' })

  const [thinkingMode, setThinkingMode] = useState('Auto')
  const [roundRobin, setRoundRobin] = useState(true)
  const [stickyCount, setStickyCount] = useState(1)

  const [showApplyProxySheet, setShowApplyProxySheet] = useState(false)
  const [isApplyingProxy, setIsApplyingProxy] = useState(false)

  const [isTestingOneByOne, setIsTestingOneByOne] = useState(false)
  const [testingIndex, setTestingIndex] = useState(-1)
  const [testStats, setTestStats] = useState({ success: 0, failed: 0, total: 0 })
  const abortTestRef = useRef(false)

  const [showAddModel, setShowAddModel] = useState(false)
  const [customModelName, setCustomModelName] = useState('')
  const [copiedModelId, setCopiedModelId] = useState<string | null>(null)

  const [showAddSheet, setShowAddSheet] = useState(false)
  const [showEditSheet, setShowEditSheet] = useState(false)
  const [editingAccount, setEditingAccount] = useState<Account | null>(null)
  const [proxyTargetAccount, setProxyTargetAccount] = useState<Account | null>(null)
  const [showRowProxySheet, setShowRowProxySheet] = useState(false)
  const [isSavingRowProxy, setIsSavingRowProxy] = useState(false)

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
      const [allProvidersRes, platformProvidersRes, allAccountsRes, proxiesRes, modelsRes, settingsRes] = await Promise.allSettled([
        api.get<Provider[]>('/api/providers'),
        api.get<PlatformProvider[]>('/api/v1/providers'),
        api.get<Account[]>('/api/accounts'),
        api.get<ProxyProfile[]>('/api/proxies'),
        api.get<any[]>('/api/models'),
        api.get<Record<string, string>>('/api/settings'),
      ])

      const allProviders: Provider[] = allProvidersRes.status === 'fulfilled' && Array.isArray(allProvidersRes.value) ? allProvidersRes.value : []
      const platformProviders: PlatformProvider[] = platformProvidersRes.status === 'fulfilled' && Array.isArray(platformProvidersRes.value) ? platformProvidersRes.value : []
      const allAccounts: Account[] = allAccountsRes.status === 'fulfilled' && Array.isArray(allAccountsRes.value) ? allAccountsRes.value : []
      const allProxies: ProxyProfile[] = proxiesRes.status === 'fulfilled' && Array.isArray(proxiesRes.value) ? proxiesRes.value : []
      const allModels: any[] = modelsRes.status === 'fulfilled' && Array.isArray(modelsRes.value) ? modelsRes.value : []
      const settings: Record<string, string> = settingsRes.status === 'fulfilled' && settingsRes.value ? settingsRes.value : {}

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

      if (settings[`round_robin_${id}`] !== undefined) {
        setRoundRobin(settings[`round_robin_${id}`] === 'true')
      }
      if (settings[`sticky_${id}`] !== undefined) {
        setStickyCount(parseInt(settings[`sticky_${id}`]) || 1)
      }
      if (settings[`thinking_${id}`] !== undefined) {
        setThinkingMode(settings[`thinking_${id}`])
      }

      const activeProviderKey = found.key || found.id
      const providerDbModels = allModels.filter(
        (m) => m.provider_id === id || m.provider_id === activeProviderKey
      )

      if (id === 'gemini' || id === 'gemini-cli' || activeProviderKey === 'gemini' || activeProviderKey === 'gemini-cli') {
        const merged: ProviderModel[] = DEFAULT_GEMINI_MODELS.map((item) => {
          const fullId = `${activeProviderKey}/${item.external_name}`
          const existing = providerDbModels.find(
            (dbm) => dbm.id === fullId || dbm.external_name === item.external_name
          )
          const isDefaultActive = item.external_name === 'gemini-3.6-flash'
          return {
            id: existing ? existing.id : fullId,
            external_name: item.external_name,
            display_name: existing?.display_name || item.display_name,
            vision: item.vision,
            web: item.web,
            enabled: existing ? existing.enabled : isDefaultActive,
          }
        })

        providerDbModels.forEach((dbm) => {
          if (!merged.some((m) => m.external_name === dbm.external_name)) {
            merged.push({
              id: dbm.id,
              external_name: dbm.external_name,
              display_name: dbm.display_name || dbm.external_name,
              vision: true,
              web: true,
              enabled: dbm.enabled,
            })
          }
        })
        setModels(merged)
      } else {
        const mapped: ProviderModel[] = providerDbModels.map((m) => ({
          id: m.id,
          external_name: m.external_name,
          display_name: m.display_name || m.external_name,
          vision: true,
          web: true,
          enabled: m.enabled,
        }))
        setModels(mapped)
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load provider details'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    window.scrollTo(0, 0)
    document.documentElement.scrollTo(0, 0)
    document.body.scrollTo(0, 0)
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

  const handleApplyProxy = async (mode: 'rotate' | 'none' | 'single', proxyPoolId?: string) => {
    if (!provider) return
    setIsApplyingProxy(true)
    try {
      const res = await api.post<{ updated: number; mode: string }>('/api/accounts/batch-proxy', {
        provider_id: provider.id,
        mode,
        proxy_pool_id: proxyPoolId,
      })
      setShowApplyProxySheet(false)
      setActionSuccess(
        mode === 'rotate'
          ? `Rotated proxies across ${res.updated} connections.`
          : mode === 'none'
          ? `Removed proxies from ${res.updated} connections.`
          : `Assigned proxy to ${res.updated} connections.`
      )
      setTimeout(() => setActionSuccess(null), 4000)
      await loadData()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to apply proxy batch')
    } finally {
      setIsApplyingProxy(false)
    }
  }

  const handleTestOneByOne = async () => {
    if (isTestingOneByOne) {
      abortTestRef.current = true
      setIsTestingOneByOne(false)
      return
    }

    if (accounts.length === 0) return

    setIsTestingOneByOne(true)
    abortTestRef.current = false
    setTestStats({ success: 0, failed: 0, total: accounts.length })

    let successCount = 0
    let failedCount = 0

    for (let i = 0; i < accounts.length; i++) {
      if (abortTestRef.current) break
      const acc = accounts[i]
      setTestingIndex(i)

      try {
        const res = await api.post<{ healthy: boolean; latency: number; message?: string }>(
          `/api/accounts/${acc.id}/test`,
          {}
        )
        if (res.healthy) {
          successCount++
          setAccounts((prev) =>
            prev.map((a) => (a.id === acc.id ? { ...a, state: 'active', last_error: undefined } : a))
          )
        } else {
          failedCount++
          setAccounts((prev) =>
            prev.map((a) =>
              a.id === acc.id
                ? { ...a, state: 'cooling_down', last_error: res.message || 'Health check failed' }
                : a
            )
          )
        }
      } catch (err: unknown) {
        failedCount++
        const msg = err instanceof Error ? err.message : 'Request failed'
        setAccounts((prev) =>
          prev.map((a) => (a.id === acc.id ? { ...a, state: 'cooling_down', last_error: msg } : a))
        )
      }

      setTestStats({ success: successCount, failed: failedCount, total: accounts.length })
    }

    setIsTestingOneByOne(false)
    setTestingIndex(-1)
  }

  const handleToggleModel = async (model: ProviderModel) => {
    const nextState = !model.enabled
    setModels((prev) => prev.map((m) => (m.external_name === model.external_name ? { ...m, enabled: nextState } : m)))
    try {
      await api.post('/api/models', {
        id: model.id,
        provider_id: provider?.key || provider?.id || id,
        external_name: model.external_name,
        display_name: model.display_name,
        streaming: true,
      })
      await api.patch(`/api/models/${encodeURIComponent(model.id)}`, { enabled: nextState })
    } catch {}
  }

  const handleBatchToggleModels = async (enabled: boolean) => {
    if (!provider) return
    setModels((prev) => prev.map((m) => ({ ...m, enabled })))
    try {
      await api.post('/api/models/batch-toggle', {
        provider_id: provider.key || provider.id || id,
        enabled,
      })
      for (const m of models) {
        if (enabled && !m.enabled) {
          await api.post('/api/models', {
            id: m.id,
            provider_id: provider.key || provider.id || id,
            external_name: m.external_name,
            display_name: m.display_name,
            streaming: true,
          })
        }
      }
    } catch {}
  }

  const handleAddCustomModel = async (e: FormEvent) => {
    e.preventDefault()
    if (!customModelName.trim() || !provider) return
    const ext = customModelName.trim().toLowerCase()
    const fullId = `${provider.key || provider.id || id}/${ext}`
    const newModel: ProviderModel = {
      id: fullId,
      external_name: ext,
      display_name: customModelName.trim(),
      vision: true,
      web: true,
      enabled: true,
    }

    setModels((prev) => [newModel, ...prev.filter((m) => m.external_name !== ext)])
    setCustomModelName('')
    setShowAddModel(false)

    try {
      await api.post('/api/models', {
        id: fullId,
        provider_id: provider.key || provider.id || id,
        external_name: ext,
        display_name: customModelName.trim(),
        streaming: true,
      })
    } catch {}
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopiedModelId(text)
    setTimeout(() => setCopiedModelId(null), 2000)
  }

  const handleRoundRobinToggle = async () => {
    const nextVal = !roundRobin
    setRoundRobin(nextVal)
    if (provider) {
      try {
        await api.put('/api/settings', { key: `round_robin_${provider.id}`, value: String(nextVal) })
      } catch {}
    }
  }

  const handleStickyChange = async (val: number) => {
    setStickyCount(val)
    if (provider) {
      try {
        await api.put('/api/settings', { key: `sticky_${provider.id}`, value: String(val) })
      } catch {}
    }
  }

  const handleThinkingChange = async (mode: string) => {
    setThinkingMode(mode)
    if (provider) {
      try {
        await api.put('/api/settings', { key: `thinking_${provider.id}`, value: mode })
      } catch {}
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
    } catch {}
  }

  const openAccountProxy = (acc: Account) => {
    setProxyTargetAccount(acc)
    setShowRowProxySheet(true)
  }

  const handleSetAccountProxy = async (proxyPoolId: string) => {
    if (!proxyTargetAccount) return
    setIsSavingRowProxy(true)
    try {
      await api.put(`/api/accounts/${proxyTargetAccount.id}`, {
        proxy_pool_id: proxyPoolId,
      })
      setShowRowProxySheet(false)
      const selectedProxy = proxies.find((p) => p.id === proxyPoolId)
      setActionSuccess(
        proxyPoolId
          ? `Proxy ${selectedProxy?.name || ''} berhasil dipasang ke ${proxyTargetAccount.name}`
          : `Proxy dilepas. ${proxyTargetAccount.name} kini terhubung Direct (tanpa proxy).`
      )
      setTimeout(() => setActionSuccess(null), 3500)
      await loadData()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal mengatur proxy')
    } finally {
      setIsSavingRowProxy(false)
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

  const toggleExpand = (accountId: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(accountId)) next.delete(accountId)
      else next.add(accountId)
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

  const activeModels = models.filter((m) => m.enabled)
  const disabledModels = models.filter((m) => !m.enabled)

  return (
    <div className="space-y-5 pb-28">
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

      {actionSuccess && (
        <div className="px-4 py-2.5 rounded-[7px] bg-emerald-950/20 border border-emerald-600/30 text-emerald-300 text-[12.5px] flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{actionSuccess}</span>
        </div>
      )}

      {provider && (
        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4 min-w-0">
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
              <div className="flex items-center gap-2 mt-1">
                <span className="text-[11.5px] font-mono text-[var(--text-muted)]">
                  {accounts.length} connections
                </span>
                <span className="text-[11.5px] text-[var(--text-muted)]">•</span>
                <span className="text-[11.5px] font-mono text-[var(--text-secondary)]">
                  {activeModels.length} active models
                </span>
              </div>
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
              Test Provider
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
          <span>{testResult.message || 'Testing provider health…'}</span>
          {testResult.latency !== undefined && (
            <span className="ml-auto font-mono text-[10.5px]">{testResult.latency}ms</span>
          )}
        </div>
      )}

      <div className="flex flex-wrap items-center justify-between gap-3 bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] px-4 py-2.5">
        <div className="flex items-center gap-2 flex-wrap">
          <button
            type="button"
            onClick={() => setShowApplyProxySheet(true)}
            className="h-8 px-3 text-[12px] font-medium rounded-[6px] bg-[#202227] hover:bg-[#2a2d35] border border-[#363a45] text-[#e0e2eb] flex items-center gap-2 transition-colors cursor-pointer"
          >
            <Network className="w-3.5 h-3.5 text-[#9fa3b4]" />
            <span>Apply Proxy</span>
          </button>

          <button
            type="button"
            onClick={handleTestOneByOne}
            className={`h-8 px-3 text-[12px] font-medium rounded-[6px] border text-[#e0e2eb] flex items-center gap-2 transition-colors cursor-pointer ${
              isTestingOneByOne
                ? 'bg-amber-950/30 border-amber-600/50 text-amber-300'
                : 'bg-[#202227] hover:bg-[#2a2d35] border-[#363a45]'
            }`}
          >
            <RefreshCw className={`w-3.5 h-3.5 text-[#9fa3b4] ${isTestingOneByOne ? 'animate-spin' : ''}`} />
            <span>{isTestingOneByOne ? 'Stop Testing' : 'Test Connection One-by-One'}</span>
          </button>
        </div>

        <div className="flex items-center gap-4 flex-wrap">
          <div className="flex items-center gap-2">
            <span className="text-[12px] font-medium text-[#9da1b2]">Round Robin</span>
            <button
              type="button"
              onClick={handleRoundRobinToggle}
              className={`w-9 h-5 rounded-full transition-colors relative flex items-center px-0.5 cursor-pointer ${
                roundRobin ? 'bg-[#ea580c]' : 'bg-[#2d3139]'
              }`}
            >
              <div
                className={`w-4 h-4 rounded-full bg-white transition-transform ${
                  roundRobin ? 'translate-x-4' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="flex items-center gap-1.5">
            <span className="text-[12px] font-medium text-[#9da1b2]">Sticky:</span>
            <input
              type="number"
              min="1"
              max="50"
              value={stickyCount}
              onChange={(e) => handleStickyChange(parseInt(e.target.value) || 1)}
              className="w-10 h-7 text-center text-[12px] font-mono rounded-[5px] bg-[#1a1b20] border border-[#333742] text-[#e0e2eb] focus:outline-none focus:border-[#ea580c]"
            />
          </div>
        </div>
      </div>

      {isTestingOneByOne && (
        <div className="px-4 py-3 rounded-[8px] bg-[#1c1e25] border border-amber-500/30 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-[12px]">
          <div className="flex items-center gap-2 text-amber-300">
            <Clock className="w-4 h-4 animate-spin shrink-0" />
            <span>
              Testing connection {testingIndex + 1} of {accounts.length} —{' '}
              <strong className="text-white">{accounts[testingIndex]?.name || '...'}</strong>
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span className="font-mono text-emerald-400">{testStats.success} Healthy</span>
            <span className="text-[#686d80]">•</span>
            <span className="font-mono text-rose-400">{testStats.failed} Failed</span>
            <button
              type="button"
              onClick={() => {
                abortTestRef.current = true
                setIsTestingOneByOne(false)
              }}
              className="px-2 py-1 rounded-[4px] bg-rose-950/40 border border-rose-600/40 text-rose-300 hover:bg-rose-900/50 text-[11px] cursor-pointer"
            >
              Stop
            </button>
          </div>
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
              const isCurrentlyTesting = isTestingOneByOne && testingIndex === idx

              return (
                <div
                  key={acc.id}
                  className={`group transition-colors ${isCurrentlyTesting ? 'bg-amber-950/10 border-l-2 border-l-amber-400' : ''}`}
                >
                  <div className="flex items-center gap-3 px-4 py-3">
                    <div className="flex items-center gap-2 shrink-0">
                      <button
                        type="button"
                        onClick={() => toggleExpand(acc.id)}
                        className="p-0.5 text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors cursor-pointer"
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

                        {(acc.proxy_pool_id || acc.proxy_url) && (
                          <span
                            title={acc.proxy_name || acc.proxy_url}
                            className="text-[10.5px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[var(--brand-subtle)] border border-[var(--brand-primary)]/30 text-[var(--brand-text)] flex items-center gap-1 max-w-[170px] truncate"
                          >
                            <Globe className="w-2.5 h-2.5 shrink-0" />
                            <span className="truncate">{acc.proxy_name || 'Proxy'}</span>
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
                        title={acc.proxy_pool_id || acc.proxy_url ? `Proxy aktif: ${acc.proxy_name || acc.proxy_url}` : 'Pasang proxy untuk koneksi ini'}
                        onClick={() => openAccountProxy(acc)}
                        className={`px-2 py-1.5 text-[10.5px] font-medium rounded-[5px] transition-colors flex items-center gap-1 cursor-pointer ${
                          acc.proxy_pool_id || acc.proxy_url
                            ? 'bg-[#2a1d17] border border-[#ea580c]/40 text-[#f97316] hover:bg-[#382319]'
                            : 'text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] border border-transparent'
                        }`}
                      >
                        <Globe className={`w-3.5 h-3.5 ${acc.proxy_pool_id || acc.proxy_url ? 'text-[#f97316]' : 'text-[var(--text-muted)]'}`} />
                        <span className="hidden sm:block">Proxy</span>
                      </button>

                      <button
                        type="button"
                        title="Edit connection"
                        onClick={() => openEdit(acc)}
                        className="px-2 py-1.5 text-[10.5px] font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] rounded-[5px] transition-colors flex items-center gap-1 cursor-pointer"
                      >
                        <Pencil className="w-3.5 h-3.5" />
                        <span className="hidden sm:block">Edit</span>
                      </button>

                      <InlineConfirm
                        trigger={
                          <button
                            type="button"
                            title="Delete connection"
                            className="px-2 py-1.5 text-[10.5px] font-medium text-[var(--text-muted)] hover:text-[var(--status-danger)] hover:bg-rose-950/20 rounded-[5px] transition-colors flex items-center gap-1 cursor-pointer"
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
                        className="px-2 py-1.5 rounded-[5px] transition-colors text-[var(--text-muted)] hover:bg-[var(--bg-panel)] cursor-pointer"
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
                            {proxy.scheme}://{proxy.host} ({proxy.name})
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

      <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] p-4 sm:p-5 space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-1 border-b border-[var(--border-subtle)]">
          <div className="flex items-center gap-3 flex-wrap">
            <h3 className="text-[15px] font-bold text-[var(--text-primary)]">Available Models</h3>

            <div className="relative inline-block">
              <select
                value={thinkingMode}
                onChange={(e) => handleThinkingChange(e.target.value)}
                className="h-7 px-2.5 pr-6 text-[11.5px] font-medium rounded-[6px] bg-[#1c1e24] border border-[#333742] text-[#d4d6df] hover:border-[var(--brand-primary)] focus:outline-none cursor-pointer appearance-none"
              >
                <option value="Auto">Thinking: Auto</option>
                <option value="Enabled">Thinking: Enabled</option>
                <option value="High">Thinking: High</option>
                <option value="Medium">Thinking: Medium</option>
                <option value="Low">Thinking: Low</option>
                <option value="Disabled">Thinking: Disabled</option>
              </select>
              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-1.5 text-[#8a8d9a]">
                <ChevronDown className="w-3 h-3" />
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => handleBatchToggleModels(true)}
              className="h-7 px-2.5 text-[11.5px] font-medium rounded-[6px] bg-[#202227] hover:bg-[#282b33] border border-[#333742] text-[#e0e2eb] flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <RotateCw className="w-3 h-3 text-[#a0a4b5]" />
              Active All
            </button>
            <button
              type="button"
              onClick={() => handleBatchToggleModels(false)}
              className="h-7 px-2.5 text-[11.5px] font-medium rounded-[6px] bg-[#202227] hover:bg-[#282b33] border border-[#333742] text-[#e0e2eb] flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <Ban className="w-3 h-3 text-[#a0a4b5]" />
              Disable All
            </button>
          </div>
        </div>

        <div className="flex items-center flex-wrap gap-2.5">
          {activeModels.map((m) => {
            const displayModelId = `${provider?.key || provider?.id || id}/${m.external_name}`
            const isCopied = copiedModelId === displayModelId

            return (
              <div
                key={m.external_name}
                className="bg-[#17181c] border border-[#2b2e37] rounded-[7px] px-3 py-2 flex items-center gap-2.5 min-w-[220px] max-w-[340px] group transition-all"
              >
                <div className="w-7 h-7 rounded-[6px] bg-[#22242b] border border-[#333742] flex items-center justify-center shrink-0 text-[#b5b8c7]">
                  <Bot className="w-4 h-4" />
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-1.5">
                    <span className="text-[11.5px] font-mono font-medium text-[#e6e8f0] truncate">
                      {displayModelId}
                    </span>
                    <button
                      type="button"
                      onClick={() => copyToClipboard(displayModelId)}
                      title="Copy model identifier"
                      className="p-1 text-[#787c8d] hover:text-[#e6e8f0] transition-colors rounded cursor-pointer"
                    >
                      {isCopied ? (
                        <Check className="w-3 h-3 text-emerald-400" />
                      ) : (
                        <Copy className="w-3 h-3" />
                      )}
                    </button>
                  </div>

                  <div className="flex items-center gap-1.5 text-[11px] text-[#8e92a2]">
                    <span className="truncate italic">{m.display_name}</span>
                    {m.vision && (
                      <span title="Vision Capable" className="inline-flex">
                        <Eye className="w-2.5 h-2.5 shrink-0 opacity-70" />
                      </span>
                    )}
                    {m.web && (
                      <span title="Grounding / Search" className="inline-flex">
                        <Globe className="w-2.5 h-2.5 shrink-0 opacity-70" />
                      </span>
                    )}
                  </div>
                </div>

                <button
                  type="button"
                  onClick={() => handleToggleModel(m)}
                  title="Disable model"
                  className="p-1 text-[#707485] hover:text-rose-400 opacity-40 group-hover:opacity-100 transition-opacity rounded cursor-pointer"
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              </div>
            )
          })}

          {showAddModel ? (
            <form
              onSubmit={handleAddCustomModel}
              className="flex items-center gap-1.5 bg-[#17181c] border border-[var(--brand-primary)]/40 rounded-[7px] px-2.5 py-1.5"
            >
              <input
                type="text"
                autoFocus
                value={customModelName}
                onChange={(e) => setCustomModelName(e.target.value)}
                placeholder="e.g. gemini-2.0-flash"
                className="w-36 h-6 text-[11.5px] font-mono bg-transparent border-none text-[#e6e8f0] focus:outline-none"
              />
              <button
                type="submit"
                className="h-6 px-2 text-[10.5px] font-medium rounded bg-[var(--brand-primary)] text-white hover:opacity-90 cursor-pointer"
              >
                Add
              </button>
              <button
                type="button"
                onClick={() => setShowAddModel(false)}
                className="p-1 text-[#707485] hover:text-[#e6e8f0] cursor-pointer"
              >
                <X className="w-3 h-3" />
              </button>
            </form>
          ) : (
            <button
              type="button"
              onClick={() => setShowAddModel(true)}
              className="h-11 px-3.5 border border-dashed border-[#ea580c]/40 hover:border-[#ea580c] rounded-[7px] text-[12px] font-medium text-[#f97316] flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <Plus className="w-3.5 h-3.5" />
              Add Model
            </button>
          )}
        </div>

        {disabledModels.length > 0 && (
          <div className="pt-2 border-t border-[var(--border-subtle)]/60">
            <p className="text-[11.5px] font-medium text-[#7d8293] mb-2">
              Disabled models ({disabledModels.length}):
            </p>
            <div className="flex flex-wrap gap-1.5">
              {disabledModels.map((dm) => (
                <button
                  key={dm.external_name}
                  type="button"
                  onClick={() => handleToggleModel(dm)}
                  title={`Click to enable ${dm.display_name}`}
                  className="px-2 py-1 text-[11px] font-mono rounded-[5px] bg-[#1a1b20] hover:bg-[#23252d] border border-[#2d3039] hover:border-[#ea580c]/50 text-[#9da1b2] hover:text-[#ea580c] transition-all flex items-center gap-1 cursor-pointer"
                >
                  <Plus className="w-2.5 h-2.5" />
                  <span>{dm.external_name}</span>
                </button>
              ))}
            </div>
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
        isOpen={showApplyProxySheet}
        onClose={() => setShowApplyProxySheet(false)}
        title={`Apply Proxy (${accounts.length} connections)`}
        description="Distribute proxies across connections or assign a single relay"
        maxWidth="max-w-md"
      >
        <div className="space-y-3">
          <div className="space-y-2">
            <button
              type="button"
              disabled={isApplyingProxy}
              onClick={() => handleApplyProxy('rotate')}
              className="w-full px-3.5 py-2.5 rounded-[7px] bg-[var(--bg-panel)] hover:bg-[var(--bg-card)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] flex items-center gap-3 text-left transition-colors group cursor-pointer"
            >
              <div className="w-7 h-7 rounded-[6px] bg-[var(--bg-surface)] border border-[var(--border-subtle)] flex items-center justify-center shrink-0 text-[var(--text-secondary)] group-hover:text-[var(--text-primary)]">
                <ArrowLeftRight className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <span className="text-[13px] font-medium text-[var(--text-primary)] block">
                  One-to-one (rotate)
                </span>
                <span className="text-[11px] text-[var(--text-muted)] block truncate">
                  Distribute {proxies.length} proxies evenly across all {accounts.length} connections
                </span>
              </div>
            </button>

            <button
              type="button"
              disabled={isApplyingProxy}
              onClick={() => handleApplyProxy('none')}
              className="w-full px-3.5 py-2.5 rounded-[7px] bg-[var(--bg-panel)] hover:bg-[var(--bg-card)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] flex items-center gap-3 text-left transition-colors group cursor-pointer"
            >
              <div className="w-7 h-7 rounded-[6px] bg-[var(--bg-surface)] border border-[var(--border-subtle)] flex items-center justify-center shrink-0 text-[var(--text-secondary)] group-hover:text-[var(--text-primary)]">
                <Unlink className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <span className="text-[13px] font-medium text-[var(--text-primary)] block">
                  None (unbind all)
                </span>
                <span className="text-[11px] text-[var(--text-muted)] block truncate">
                  Remove proxy from all connections (direct connection)
                </span>
              </div>
            </button>
          </div>

          <div className="pt-2 pb-1 text-[11px] font-semibold uppercase text-[var(--text-muted)] tracking-wider">
            Available Proxies ({proxies.length})
          </div>

          <div className="max-h-80 overflow-y-auto space-y-1.5 pr-1">
            {proxies.map((p) => (
              <button
                key={p.id}
                type="button"
                disabled={isApplyingProxy}
                onClick={() => handleApplyProxy('single', p.id)}
                className="w-full px-3 py-2 rounded-[6px] bg-[var(--bg-panel)] hover:bg-[var(--bg-card)] border border-[var(--border-subtle)] hover:border-[var(--border-strong)] flex items-center gap-2.5 text-left transition-colors group cursor-pointer"
              >
                <div className="w-6 h-6 rounded-[5px] bg-[var(--bg-surface)] border border-[var(--border-subtle)] flex items-center justify-center shrink-0 text-[var(--text-secondary)]">
                  <Network className="w-3.5 h-3.5" />
                </div>
                <div className="flex-1 min-w-0">
                  <span className="text-[12.5px] font-medium text-[var(--text-primary)] block truncate">
                    {p.name}
                  </span>
                  <span className="text-[10.5px] font-mono text-[var(--text-muted)] block truncate">
                    {p.scheme}://{p.host}
                  </span>
                </div>
              </button>
            ))}
          </div>
        </div>
      </BottomSheet>

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
        isOpen={showRowProxySheet}
        onClose={() => setShowRowProxySheet(false)}
        title={`Set Proxy — ${proxyTargetAccount?.name || ''}`}
        description="Pilih relay proxy untuk koneksi ini atau gunakan direct (tanpa proxy)"
        maxWidth="max-w-md"
      >
        <div className="space-y-2.5">
          <button
            type="button"
            disabled={isSavingRowProxy}
            onClick={() => handleSetAccountProxy('')}
            className={`w-full px-3.5 py-2.5 rounded-[7px] border flex items-center justify-between transition-colors cursor-pointer text-left ${
              !proxyTargetAccount?.proxy_pool_id && !proxyTargetAccount?.proxy_url
                ? 'bg-[#2a1d17] border-[#ea580c]/40 text-[#f97316]'
                : 'bg-[var(--bg-panel)] hover:bg-[var(--bg-card)] border-[var(--border-subtle)] hover:border-[var(--border-strong)] text-[var(--text-primary)]'
            }`}
          >
            <div className="flex items-center gap-3">
              <div className="w-7 h-7 rounded-[6px] bg-[var(--bg-surface)] border border-[var(--border-subtle)] flex items-center justify-center shrink-0">
                <Unlink className="w-3.5 h-3.5 text-[#9fa3b4]" />
              </div>
              <div>
                <span className="text-[13px] font-medium block">Direct (Tanpa Proxy)</span>
                <span className="text-[11px] text-[var(--text-muted)] block">
                  Koneksi langsung ke server provider tanpa perantara proxy
                </span>
              </div>
            </div>
            {!proxyTargetAccount?.proxy_pool_id && !proxyTargetAccount?.proxy_url && (
              <Check className="w-4 h-4 text-[#ea580c] shrink-0" />
            )}
          </button>

          <div className="pt-2 pb-1 text-[11px] font-semibold uppercase text-[var(--text-muted)] tracking-wider">
            Daftar Proxy Tersedia ({proxies.length})
          </div>

          {proxies.length === 0 ? (
            <div className="py-6 text-center text-[12px] text-[var(--text-muted)] bg-[var(--bg-panel)] rounded-[7px] border border-[var(--border-subtle)]">
              Belum ada proxy di sistem. Tambahkan proxy di menu Proxies.
            </div>
          ) : (
            <div className="max-h-72 overflow-y-auto space-y-1.5 pr-1">
              {proxies.map((p) => {
                const isSelected = proxyTargetAccount?.proxy_pool_id === p.id
                return (
                  <button
                    key={p.id}
                    type="button"
                    disabled={isSavingRowProxy}
                    onClick={() => handleSetAccountProxy(p.id)}
                    className={`w-full px-3 py-2 rounded-[6px] border flex items-center justify-between transition-colors cursor-pointer text-left ${
                      isSelected
                        ? 'bg-[#2a1d17] border-[#ea580c]/40 text-[#f97316]'
                        : 'bg-[var(--bg-panel)] hover:bg-[var(--bg-card)] border-[var(--border-subtle)] hover:border-[var(--border-strong)] text-[var(--text-primary)]'
                    }`}
                  >
                    <div className="flex items-center gap-2.5 min-w-0">
                      <Globe className={`w-3.5 h-3.5 shrink-0 ${isSelected ? 'text-[#ea580c]' : 'text-[var(--text-muted)]'}`} />
                      <div className="min-w-0">
                        <span className="text-[12.5px] font-medium block truncate">{p.name}</span>
                        <span className="text-[11px] font-mono text-[var(--text-muted)] block truncate">
                          {p.scheme}://{p.host}{p.port && p.port !== 80 && p.port !== 443 ? `:${p.port}` : ''}
                        </span>
                      </div>
                    </div>
                    {isSelected && <Check className="w-4 h-4 text-[#ea580c] shrink-0 ml-2" />}
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </BottomSheet>

      <BottomSheet
        isOpen={showEditSheet}
        onClose={() => setShowEditSheet(false)}
        title={`Edit Koneksi — ${editingAccount?.name || ''}`}
        description="Perbarui nama koneksi, prioritas routing, proxy, atau ganti API key"
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
            <div className="flex items-center justify-between mb-1.5">
              <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)]">
                Ganti API Key Baru
              </label>
              <span className="text-[10.5px] text-[var(--text-muted)]">
                (Biarkan kosong jika tidak diubah)
              </span>
            </div>
            <input
              type="password"
              autoComplete="new-password"
              value={editForm.api_key}
              onChange={(e) => setEditForm({ ...editForm, api_key: e.target.value })}
              placeholder="Masukkan API key baru untuk mengganti yang lama…"
              className="w-full h-9 px-3 text-[13px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)] transition-colors"
            />
            <p className="text-[11px] text-[var(--text-muted)] mt-1">
              API key lama tetap tersimpan aman. Isi kolom ini hanya jika Anda ingin mengganti dengan API key baru.
            </p>
          </div>

          <div className="flex items-center gap-3 p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
            <label className="text-[12.5px] font-medium text-[var(--text-secondary)] flex-1">
              Connection Enabled
            </label>
            <button
              type="button"
              onClick={() => setEditForm({ ...editForm, enabled: !editForm.enabled })}
              className="transition-colors cursor-pointer"
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
