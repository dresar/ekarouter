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
  KeyRound,
  Loader2,
  X,
} from 'lucide-react'
import { Button } from '../../components/ui/Button.tsx'
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
  const [selectedAccountIds, setSelectedAccountIds] = useState<Set<string>>(new Set())
  const [isBulkDeleting, setIsBulkDeleting] = useState(false)
  const [addModalTab, setAddModalTab] = useState<'single' | 'bulk'>('single')
  const [bulkPriority, setBulkPriority] = useState(1)
  const [checkingKey, setCheckingKey] = useState(false)
  const [keyCheckResult, setKeyCheckResult] = useState<{
    status: 'valid' | 'invalid'
    latency?: number
    message?: string
  } | null>(null)
  const [bulkText, setBulkText] = useState('')
  const [bulkPrefix, setBulkPrefix] = useState('Key')
  const [bulkProxyId, setBulkProxyId] = useState('')
  const [isBulkAdding, setIsBulkAdding] = useState(false)
  const [bulkError, setBulkError] = useState<string | null>(null)

  const [showOAuthModal, setShowOAuthModal] = useState(false)
  const [isStartingOAuth, setIsStartingOAuth] = useState(false)
  const [oauthData, setOauthData] = useState<{ auth_url: string; state: string } | null>(null)
  const [manualCode, setManualCode] = useState('')
  const [isCompletingOAuth, setIsCompletingOAuth] = useState(false)
  const [oauthError, setOauthError] = useState<string | null>(null)

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

  useEffect(() => {
    const handleMsg = (e: MessageEvent) => {
      if (e.data && e.data.type === 'oauth_success') {
        setShowOAuthModal(false)
        setActionSuccess('Account connected successfully via OAuth!')
        setTimeout(() => setActionSuccess(null), 3500)
        loadData()
      }
    }
    window.addEventListener('message', handleMsg)
    return () => window.removeEventListener('message', handleMsg)
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

  const handleCheckApiKey = async () => {
    if (!addForm.api_key.trim() || !id) return
    setCheckingKey(true)
    setKeyCheckResult(null)
    try {
      const res = await api.post<{ status: string; latency_ms?: number; message?: string }>('/api/keys/health', {
        api_key: addForm.api_key.trim(),
        provider: id,
        proxy_url: addForm.proxy_pool_id ? proxies.find((p) => p.id === addForm.proxy_pool_id)?.host : undefined,
      })
      setKeyCheckResult({
        status: res.status === 'valid' || res.status === 'healthy' ? 'valid' : 'invalid',
        latency: res.latency_ms,
        message: res.message,
      })
    } catch (err: unknown) {
      setKeyCheckResult({
        status: 'invalid',
        message: err instanceof Error ? err.message : 'Invalid API key',
      })
    } finally {
      setCheckingKey(false)
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
      await api.patch(`/api/accounts/${proxyTargetAccount.id}`, {
        proxy_pool_id: proxyPoolId || null,
      })
      setShowRowProxySheet(false)
      setProxyTargetAccount(null)
      await loadData()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to update proxy')
    } finally {
      setIsSavingRowProxy(false)
    }
  }

  const handleBulkAddSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!provider) return

    const lines = bulkText
      .split('\n')
      .map((l) => l.trim())
      .filter((l) => l.length > 0)

    if (lines.length === 0) {
      setBulkError('Masukkan minimal 1 API key')
      return
    }

    setIsBulkAdding(true)
    setBulkError(null)

    try {
      const payloads = lines.map((line, idx) => {
        let name = ''
        let key = line

        if (line.includes(':') && !line.startsWith('http')) {
          const parts = line.split(':')
          name = parts[0].trim()
          key = parts.slice(1).join(':').trim()
        }

        if (!name) {
          name = `${bulkPrefix.trim() || 'Key'} ${accounts.length + idx + 1}`
        }

        return {
          provider_id: provider.id,
          name,
          auth_type: 'apikey',
          priority: Number(bulkPriority) || (accounts.length + idx + 1),
          api_key: key,
          proxy_pool_id: bulkProxyId || undefined,
        }
      })

      const results = await Promise.allSettled(
        payloads.map((p) => api.post('/api/accounts', p))
      )

      const successCount = results.filter((r) => r.status === 'fulfilled').length
      setShowAddSheet(false)
      setBulkText('')
      setActionSuccess(`Berhasil menambahkan ${successCount} key secara massal!`)
      setTimeout(() => setActionSuccess(null), 4000)
      await loadData()
    } catch (err: unknown) {
      setBulkError(err instanceof Error ? err.message : 'Gagal menambahkan key massal')
    } finally {
      setIsBulkAdding(false)
    }
  }

  const handleStartOAuth = async () => {
    if (!id) return
    setIsStartingOAuth(true)
    setOauthError(null)
    setOauthData(null)
    setManualCode('')
    try {
      const redirectUri = window.location.origin + '/api/accounts/oauth/callback'
      const res = await api.post<{ auth_url: string; state: string }>('/api/accounts/oauth/start', {
        provider_id: id,
        redirect_uri: redirectUri,
      })
      if (res && res.auth_url) {
        setOauthData(res)
        setShowOAuthModal(true)
        window.open(res.auth_url, '_blank', 'width=600,height=700')
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to initiate OAuth')
    } finally {
      setIsStartingOAuth(false)
    }
  }

  const handleCompleteOAuth = async (codeToUse?: string) => {
    const code = codeToUse || manualCode.trim()
    if (!code || !oauthData) return
    setIsCompletingOAuth(true)
    setOauthError(null)
    try {
      let cleanCode = code
      if (cleanCode.includes('code=')) {
        const u = new URL(cleanCode.startsWith('http') ? cleanCode : 'http://dummy/' + cleanCode)
        cleanCode = u.searchParams.get('code') || cleanCode
      }
      const redirectUri = window.location.origin + '/api/accounts/oauth/callback'
      await api.post('/api/accounts/oauth/callback', {
        state: oauthData.state,
        code: cleanCode,
        redirect_uri: redirectUri,
      })
      setShowOAuthModal(false)
      setActionSuccess('Account successfully connected via OAuth!')
      setTimeout(() => setActionSuccess(null), 3500)
      await loadData()
    } catch (err: unknown) {
      setOauthError(err instanceof Error ? err.message : 'Failed to complete OAuth authentication')
    } finally {
      setIsCompletingOAuth(false)
    }
  }

  const handleResetAllActive = async () => {
    const inactiveAccounts = accounts.filter(
      (a) => !a.enabled || a.state === 'cooling_down' || !!a.last_error
    )
    if (inactiveAccounts.length === 0) return

    try {
      await Promise.allSettled(
        inactiveAccounts.map((a) =>
          api.put(`/api/accounts/${a.id}`, { state: 'active', enabled: true })
        )
      )
      setActionSuccess(`Semua ${inactiveAccounts.length} koneksi berhasil diaktifkan kembali ke status active!`)
      setTimeout(() => setActionSuccess(null), 3500)
      await loadData()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal me-reset status active')
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

  const handleMovePriority = async (index: number, direction: 'up' | 'down') => {
    const targetIndex = direction === 'up' ? index - 1 : index + 1
    if (targetIndex < 0 || targetIndex >= accounts.length) return

    const currentAcc = accounts[index]
    const targetAcc = accounts[targetIndex]

    let currP = currentAcc.priority
    let targetP = targetAcc.priority

    if (currP === targetP) {
      if (direction === 'up') {
        currP = targetP + 1
      } else {
        currP = Math.max(1, targetP - 1)
      }
    } else {
      const temp = currP
      currP = targetP
      targetP = temp
    }

    try {
      await Promise.all([
        api.put(`/api/accounts/${currentAcc.id}`, { priority: currP }),
        api.put(`/api/accounts/${targetAcc.id}`, { priority: targetP }),
      ])
      await loadData()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal memindahkan prioritas')
    }
  }

  const handleToggleSelectAll = () => {
    if (selectedAccountIds.size === accounts.length && accounts.length > 0) {
      setSelectedAccountIds(new Set())
    } else {
      setSelectedAccountIds(new Set(accounts.map((a) => a.id)))
    }
  }

  const handleToggleSelectRow = (id: string) => {
    setSelectedAccountIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const handleBulkDeleteSelected = async () => {
    if (selectedAccountIds.size === 0) return
    if (!window.confirm(`Hapus ${selectedAccountIds.size} koneksi yang dipilih?`)) return
    setIsBulkDeleting(true)
    try {
      const ids = Array.from(selectedAccountIds)
      await Promise.allSettled(ids.map((id) => api.delete(`/api/accounts/${id}`)))
      setSelectedAccountIds(new Set())
      setActionSuccess(`Berhasil menghapus ${ids.length} koneksi.`)
      setTimeout(() => setActionSuccess(null), 3500)
      await loadData()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal menghapus koneksi terpilih')
    } finally {
      setIsBulkDeleting(false)
    }
  }

  const handleBulkDeleteUnavailable = async () => {
    const unavailable = accounts.filter(
      (a) => !a.enabled || a.state === 'cooling_down' || !!a.last_error
    )
    if (unavailable.length === 0) return
    if (!window.confirm(`Hapus ${unavailable.length} koneksi yang bermasalah / unavailable?`)) return
    setIsBulkDeleting(true)
    try {
      await Promise.allSettled(unavailable.map((a) => api.delete(`/api/accounts/${a.id}`)))
      setSelectedAccountIds((prev) => {
        const next = new Set(prev)
        unavailable.forEach((a) => next.delete(a.id))
        return next
      })
      setActionSuccess(`Berhasil menghapus ${unavailable.length} koneksi yang unavailable.`)
      setTimeout(() => setActionSuccess(null), 3500)
      await loadData()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Gagal menghapus koneksi unavailable')
    } finally {
      setIsBulkDeleting(false)
    }
  }

  const toggleExpand = (accountId: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(accountId)) next.delete(accountId)
      else next.add(accountId)
      return next
    })
  }


  const getStateLabel = (acc: Account) => {
    if (!acc.enabled) return 'disabled'
    return acc.state || 'unknown'
  }

  const proxyById = (proxyId?: string) => proxies.find((p) => p.id === proxyId)

  const activeModels = models.filter((m) => m.enabled)
  const disabledModels = models.filter((m) => !m.enabled)

  const isOAuth =
    id === 'antigravity' ||
    id === 'gemini-agy' ||
    id === 'gemini-cli' ||
    id === 'claude' ||
    id === 'codex' ||
    id === 'github-copilot' ||
    id === 'github' ||
    id === 'qoder' ||
    id === 'cursor' ||
    id === 'kilocode' ||
    id === 'cline' ||
    id === 'clinepass' ||
    id === 'codebuddy-intl' ||
    id === 'codebuddy-cn' ||
    id === 'kimi' ||
    id === 'grok-cli' ||
    id === 'xai' ||
    id === 'xiaomi-mimo' ||
    provider?.kind === 'gemini-agy' ||
    platProvider?.category === 'oauth' ||
    platProvider?.auth_type === 'oauth2'

  return (
    <div className="space-y-4 pb-28">
      <div className="flex items-center justify-between gap-3">
        <button
          type="button"
          onClick={() => navigate('/providers')}
          className="inline-flex items-center gap-1.5 text-[12px] font-medium text-[#9ca3af] hover:text-white transition-colors cursor-pointer"
        >
          <ArrowLeft className="w-3.5 h-3.5" />
          <span>Kembali</span>
        </button>

        {provider && (
          <span className="text-[11px] font-mono text-[#717686] bg-[#1a1b20] border border-[#2b2f3a] px-2 py-0.5 rounded-[4px]">
            Provider ID: {id || provider.id}
          </span>
        )}
      </div>

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
                <h2 className="text-[18px] font-bold text-[var(--text-primary)]">{provider.name}</h2>
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
        <div className="flex flex-wrap items-center justify-between gap-3 px-4 py-3 border-b border-[var(--border-subtle)] bg-[#14151b]">
          <div className="flex items-center gap-2">
            <h3 className="text-[16px] font-bold text-[var(--text-primary)]">
              Connections
            </h3>
            <span className="text-[11px] font-mono text-[var(--text-muted)] bg-[var(--bg-panel)] px-1.5 py-0.5 rounded">
              {accounts.length}
            </span>
          </div>

          <div className="flex items-center gap-2 flex-wrap">
            {isOAuth ? (
              <button
                type="button"
                onClick={handleStartOAuth}
                disabled={isStartingOAuth}
                className="h-8 px-3 text-[12px] font-semibold rounded-[6px] bg-indigo-600 hover:bg-indigo-500 text-white flex items-center gap-1.5 transition-colors cursor-pointer shadow-sm"
              >
                <KeyRound className="w-3.5 h-3.5" />
                <span>{isStartingOAuth ? 'Starting...' : 'Connect OAuth'}</span>
              </button>
            ) : (
              <button
                type="button"
                onClick={() => {
                  setAddModalTab('single')
                  setAddForm({
                    name: '',
                    auth_type: 'apikey',
                    priority: accounts.length + 1,
                    api_key: '',
                    proxy_pool_id: '',
                  })
                  setBulkText('')
                  setBulkPrefix('Key')
                  setBulkPriority(accounts.length + 1)
                  setBulkProxyId('')
                  setFormError(null)
                  setBulkError(null)
                  setKeyCheckResult(null)
                  setShowAddSheet(true)
                }}
                className="h-8 px-3 text-[12px] font-semibold rounded-[6px] bg-[#ea580c] hover:bg-[#f97316] text-white flex items-center gap-1.5 transition-colors cursor-pointer shadow-sm"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add Key</span>
              </button>
            )}

            <button
              type="button"
              onClick={() => setShowApplyProxySheet(true)}
              className="h-8 px-3 text-[12px] font-medium rounded-[6px] bg-[#202227] hover:bg-[#2a2d35] border border-[#363a45] text-[#e0e2eb] flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <Network className="w-3.5 h-3.5 text-[#9fa3b4]" />
              <span>Apply Proxy</span>
            </button>

            <button
              type="button"
              onClick={handleTestOneByOne}
              className={`h-8 px-3 text-[12px] font-medium rounded-[6px] border text-[#e0e2eb] flex items-center gap-1.5 transition-colors cursor-pointer ${
                isTestingOneByOne
                  ? 'bg-amber-950/30 border-amber-600/50 text-amber-300'
                  : 'bg-[#202227] hover:bg-[#2a2d35] border-[#363a45]'
              }`}
            >
              <RefreshCw className={`w-3.5 h-3.5 text-[#9fa3b4] ${isTestingOneByOne ? 'animate-spin' : ''}`} />
              <span>{isTestingOneByOne ? 'Stop Testing' : 'Test One-by-One'}</span>
            </button>

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

        <div className="flex items-center justify-between px-4 py-2 bg-[#121318] border-b border-[#232630] text-[12px] flex-wrap gap-2">
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 cursor-pointer select-none text-[#9ca3af] hover:text-[#e0e2eb]">
              <input
                type="checkbox"
                checked={accounts.length > 0 && selectedAccountIds.size === accounts.length}
                onChange={handleToggleSelectAll}
                className="w-4 h-4 rounded border-[#383d4c] bg-[#1a1b20] text-[#ea580c] focus:ring-0 focus:ring-offset-0 cursor-pointer"
              />
              <span className="font-medium">Select All</span>
            </label>

            {selectedAccountIds.size > 0 && (
              <span className="text-[#6b7280]">
                ({selectedAccountIds.size} of {accounts.length} selected)
              </span>
            )}
          </div>

          <div className="flex items-center gap-2 flex-wrap">
            {selectedAccountIds.size > 0 && (
              <button
                type="button"
                disabled={isBulkDeleting}
                onClick={handleBulkDeleteSelected}
                className="h-7 px-2.5 text-[11px] font-medium rounded-[5px] bg-rose-950/40 border border-rose-600/40 text-rose-300 hover:bg-rose-900/50 flex items-center gap-1.5 transition-colors cursor-pointer"
              >
                <Trash2 className="w-3 h-3" />
                <span>Hapus ({selectedAccountIds.size})</span>
              </button>
            )}

            {accounts.some((a) => !a.enabled || a.state === 'cooling_down' || !!a.last_error) && (
              <>
                <button
                  type="button"
                  onClick={handleResetAllActive}
                  className="h-7 px-2.5 text-[11px] font-medium rounded-[5px] bg-emerald-950/40 border border-emerald-600/40 text-emerald-300 hover:bg-emerald-900/50 flex items-center gap-1.5 transition-colors cursor-pointer"
                >
                  <Check className="w-3 h-3" />
                  <span>Aktifkan</span>
                </button>

                <button
                  type="button"
                  disabled={isBulkDeleting}
                  onClick={handleBulkDeleteUnavailable}
                  className="h-7 px-2.5 text-[11px] font-medium rounded-[5px] bg-[#221c21] border border-rose-500/30 text-rose-300 hover:bg-rose-950/50 flex items-center gap-1.5 transition-colors cursor-pointer"
                >
                  <Trash2 className="w-3 h-3" />
                  <span>Bersihkan</span>
                </button>
              </>
            )}
          </div>
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
          <div className="max-h-[480px] overflow-y-auto divide-y divide-[#232630] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-[#363a45] [&::-webkit-scrollbar-thumb]:rounded-full">
            {accounts.map((acc, idx) => {
              const isExpanded = expandedIds.has(acc.id)
              const proxy = proxyById(acc.proxy_pool_id)
              const isCurrentlyTesting = isTestingOneByOne && testingIndex === idx
              const hasProxy = !!(acc.proxy_pool_id || acc.proxy_url)

              return (
                <div
                  key={acc.id}
                  className={`group transition-colors ${
                    isCurrentlyTesting ? 'bg-amber-950/10 border-l-2 border-l-amber-400' : 'hover:bg-[#15161c]'
                  }`}
                >
                  <div className="flex items-center gap-3 px-4 py-2.5">
                    <input
                      type="checkbox"
                      checked={selectedAccountIds.has(acc.id)}
                      onChange={() => handleToggleSelectRow(acc.id)}
                      className="w-4 h-4 rounded border-[#383d4c] bg-[#1a1b20] text-[#ea580c] focus:ring-0 focus:ring-offset-0 cursor-pointer shrink-0"
                    />

                    <div className="flex flex-col items-center justify-center -space-y-1 text-[#6b7280] shrink-0">
                      <button
                        type="button"
                        disabled={idx === 0}
                        title="Move priority up"
                        onClick={() => handleMovePriority(idx, 'up')}
                        className="p-0.5 hover:text-white disabled:opacity-20 disabled:hover:text-[#6b7280] cursor-pointer"
                      >
                        <ChevronUp className="w-3.5 h-3.5" />
                      </button>
                      <button
                        type="button"
                        disabled={idx === accounts.length - 1}
                        title="Move priority down"
                        onClick={() => handleMovePriority(idx, 'down')}
                        className="p-0.5 hover:text-white disabled:opacity-20 disabled:hover:text-[#6b7280] cursor-pointer"
                      >
                        <ChevronDown className="w-3.5 h-3.5" />
                      </button>
                    </div>

                    <KeyRound className="w-4 h-4 text-[#8e93a6] shrink-0" />

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <button
                          type="button"
                          onClick={() => toggleExpand(acc.id)}
                          className="text-[13px] font-bold text-[#f3f4f6] hover:text-[#ea580c] transition-colors truncate text-left cursor-pointer"
                        >
                          {acc.name}
                        </button>

                        <span
                          className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10.5px] font-medium ${
                            !acc.enabled
                              ? 'bg-[#202228] text-[#8e93a6] border border-[#333744]'
                              : acc.state === 'active'
                              ? 'bg-emerald-950/40 text-emerald-400 border border-emerald-600/30'
                              : 'bg-rose-950/40 text-rose-400 border border-rose-600/30'
                          }`}
                        >
                          <span
                            className={`w-1.5 h-1.5 rounded-full ${
                              !acc.enabled
                                ? 'bg-[#8e93a6]'
                                : acc.state === 'active'
                                ? 'bg-emerald-400'
                                : 'bg-rose-400'
                            }`}
                          />
                          {getStateLabel(acc)}
                        </span>

                        {hasProxy && (
                          <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[4px] bg-emerald-950/40 border border-emerald-600/30 text-emerald-300">
                            Proxy
                          </span>
                        )}

                        <span className="text-[10.5px] font-mono text-[#6b7280]">
                          #{idx + 1}
                        </span>
                      </div>

                      <div className="flex items-center gap-2 flex-wrap text-[11px] text-[#9ca3af] mt-0.5">
                        <span>Pool: {acc.proxy_name || (hasProxy ? 'Proxy Relay' : 'Direct (no proxy)')}</span>
                        {acc.proxy_url && (
                          <span className="px-1.5 py-0.2 rounded-[4px] bg-[#1a1c22] border border-[#2b2f3d] text-[#8e93a6] font-mono text-[10.5px] truncate max-w-[280px]">
                            {acc.proxy_url}
                          </span>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center gap-1 shrink-0">
                      <button
                        type="button"
                        title={hasProxy ? `Proxy aktif: ${acc.proxy_name || acc.proxy_url}` : 'Pasang proxy untuk koneksi ini'}
                        onClick={() => openAccountProxy(acc)}
                        className={`min-w-[42px] py-1 px-1.5 rounded-[6px] flex flex-col items-center justify-center gap-0.5 transition-colors cursor-pointer ${
                          hasProxy
                            ? 'text-[#f97316] hover:bg-[#2a1d17]'
                            : 'text-[#8e93a6] hover:text-[#e0e2eb] hover:bg-[#20222a]'
                        }`}
                      >
                        <Network className="w-4 h-4" />
                        <span className="text-[10px] font-medium">Proxy</span>
                      </button>

                      <button
                        type="button"
                        title="Edit connection"
                        onClick={() => openEdit(acc)}
                        className="min-w-[42px] py-1 px-1.5 rounded-[6px] flex flex-col items-center justify-center gap-0.5 text-[#8e93a6] hover:text-white hover:bg-[#20222a] transition-colors cursor-pointer"
                      >
                        <Pencil className="w-4 h-4" />
                        <span className="text-[10px] font-medium">Edit</span>
                      </button>

                      <InlineConfirm
                        trigger={
                          <button
                            type="button"
                            title="Delete connection"
                            className="min-w-[42px] py-1 px-1.5 rounded-[6px] flex flex-col items-center justify-center gap-0.5 text-rose-400 hover:text-rose-300 hover:bg-rose-950/30 transition-colors cursor-pointer"
                          >
                            <Trash2 className="w-4 h-4" />
                            <span className="text-[10px] font-medium">Delete</span>
                          </button>
                        }
                        confirmText="Delete?"
                        onConfirm={() => handleDelete(acc.id)}
                      />

                      <button
                        type="button"
                        title={acc.enabled ? 'Click to disable' : 'Click to enable'}
                        onClick={() => handleToggle(acc)}
                        className={`w-9 h-5 rounded-full transition-colors relative flex items-center px-0.5 cursor-pointer shrink-0 ml-1 ${
                          acc.enabled ? 'bg-[#ea580c]' : 'bg-[#2d3139]'
                        }`}
                      >
                        <div
                          className={`w-4 h-4 rounded-full bg-white transition-transform ${
                            acc.enabled ? 'translate-x-4' : 'translate-x-0'
                          }`}
                        />
                      </button>
                    </div>
                  </div>

                  {isExpanded && (
                    <div className="px-12 py-3 bg-[#111216] border-t border-[#232630] grid grid-cols-1 sm:grid-cols-3 gap-3 text-[11.5px]">
                      <div>
                        <span className="text-[10px] font-semibold uppercase text-[#717686] block mb-0.5">
                          Masked Key
                        </span>
                        <span className="font-mono text-[#d1d5db]">
                          {acc.masked_secret || '••••••••••••'}
                        </span>
                      </div>
                      <div>
                        <span className="text-[10px] font-semibold uppercase text-[#717686] block mb-0.5">
                          Priority
                        </span>
                        <span className="font-mono text-[#ea580c] font-semibold">
                          #{acc.priority}
                        </span>
                      </div>
                      <div>
                        <span className="text-[10px] font-semibold uppercase text-[#717686] block mb-0.5">
                          Proxy Details
                        </span>
                        {proxy ? (
                          <a
                            href={`${proxy.scheme}://${proxy.host}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="font-mono text-[#ea580c] hover:underline truncate block max-w-[220px]"
                          >
                            {proxy.scheme}://{proxy.host} ({proxy.name})
                          </a>
                        ) : acc.proxy_url ? (
                          <span className="font-mono text-[#d1d5db] truncate block max-w-[220px]">
                            {acc.proxy_url}
                          </span>
                        ) : (
                          <span className="text-[#717686]">Direct (no proxy)</span>
                        )}
                      </div>
                      {acc.last_error && (
                        <div className="col-span-full p-2.5 rounded-[6px] bg-rose-950/30 border border-rose-600/30 text-rose-300 font-mono text-[11px] break-all">
                          <span className="font-bold block mb-1 text-rose-200">Error Detail:</span>
                          {acc.last_error}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}

        <div className="flex items-center justify-between px-4 py-2.5 border-t border-[#232630] bg-[#14151b] text-[11.5px] text-[#717686]">
          <span>{accounts.length} total connections</span>
          <span>Max 6 visible before internal scroll</span>
        </div>
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
                placeholder="Model"
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
        onClose={() => {
          setShowAddSheet(false)
          setKeyCheckResult(null)
          setFormError(null)
          setBulkError(null)
        }}
        title={`Add ${provider?.name || ''} API Key`}
        maxWidth="max-w-lg"
      >
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => {
                setAddModalTab('single')
                setFormError(null)
              }}
              className={`px-3 py-1 text-[12px] font-semibold rounded-[6px] transition-colors cursor-pointer ${
                addModalTab === 'single'
                  ? 'bg-[#ea580c] text-white shadow-xs'
                  : 'text-[#9fa3b4] hover:text-white hover:bg-[#202227]'
              }`}
            >
              Single
            </button>
            <button
              type="button"
              onClick={() => {
                setAddModalTab('bulk')
                setBulkError(null)
              }}
              className={`px-3 py-1 text-[12px] font-semibold rounded-[6px] transition-colors cursor-pointer ${
                addModalTab === 'bulk'
                  ? 'bg-[#ea580c] text-white shadow-xs'
                  : 'text-[#9fa3b4] hover:text-white hover:bg-[#202227]'
              }`}
            >
              Bulk Add
            </button>
          </div>

          {addModalTab === 'single' ? (
            <form onSubmit={handleAddSubmit} className="space-y-4">
              {formError && (
                <div className="flex items-center gap-2 p-2.5 rounded-[6px] bg-rose-950/20 border border-rose-600/30 text-rose-300 text-[12px]">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  {formError}
                </div>
              )}

              <div>
                <label className="block text-[12px] font-semibold text-[#e0e2eb] mb-1.5">
                  Name
                </label>
                <input
                  type="text"
                  required
                  value={addForm.name}
                  onChange={(e) => setAddForm({ ...addForm, name: e.target.value })}
                  placeholder="Production Key"
                  className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] placeholder:text-[#63687b] focus:outline-none focus:border-[#ea580c] transition-colors"
                />
              </div>

              <div>
                <label className="block text-[12px] font-semibold text-[#e0e2eb] mb-1.5">
                  API Key
                </label>
                <div className="relative flex items-center">
                  <input
                    type="password"
                    required
                    autoComplete="new-password"
                    value={addForm.api_key}
                    onChange={(e) => {
                      setAddForm({ ...addForm, api_key: e.target.value })
                      setKeyCheckResult(null)
                    }}
                    placeholder="Enter API Key"
                    className="w-full h-9 pl-3 pr-20 text-[13px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] placeholder:text-[#63687b] focus:outline-none focus:border-[#ea580c] transition-colors"
                  />
                  <button
                    type="button"
                    onClick={handleCheckApiKey}
                    disabled={checkingKey || !addForm.api_key.trim()}
                    className="absolute right-1.5 h-6 px-2.5 text-[11.5px] font-medium rounded-[5px] bg-[#2a2d37] hover:bg-[#363a45] text-[#d0d3de] border border-[#3e4250] flex items-center gap-1 cursor-pointer transition-colors disabled:opacity-40"
                  >
                    {checkingKey && <Loader2 className="w-3 h-3 animate-spin" />}
                    <span>Check</span>
                  </button>
                </div>
                {keyCheckResult && (
                  <div
                    className={`mt-1.5 text-[11.5px] flex items-center gap-1.5 ${
                      keyCheckResult.status === 'valid'
                        ? 'text-emerald-400'
                        : 'text-rose-400'
                    }`}
                  >
                    {keyCheckResult.status === 'valid' ? (
                      <CheckCircle2 className="w-3.5 h-3.5 shrink-0" />
                    ) : (
                      <AlertCircle className="w-3.5 h-3.5 shrink-0" />
                    )}
                    <span>
                      {keyCheckResult.message || (keyCheckResult.status === 'valid' ? 'Key valid' : 'Key invalid')}
                      {keyCheckResult.latency ? ` (${keyCheckResult.latency}ms)` : ''}
                    </span>
                  </div>
                )}
              </div>

              <div>
                <label className="block text-[12px] font-semibold text-[#e0e2eb] mb-1.5">
                  Priority
                </label>
                <input
                  type="number"
                  min="1"
                  max="100"
                  value={addForm.priority}
                  onChange={(e) => setAddForm({ ...addForm, priority: parseInt(e.target.value) || 1 })}
                  className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                />
              </div>

              <div>
                <label className="block text-[12px] font-semibold text-[#e0e2eb] mb-1.5">
                  Proxy Pool
                </label>
                <select
                  value={addForm.proxy_pool_id}
                  onChange={(e) => setAddForm({ ...addForm, proxy_pool_id: e.target.value })}
                  className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                >
                  <option value="">None</option>
                  {proxies.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name} — {p.scheme}://{p.host}
                    </option>
                  ))}
                </select>
                <p className="text-[11px] text-[#787c8d] mt-2">
                  Legacy manual proxy fields are still accepted by API for backward compatibility.
                </p>
              </div>

              <div className="flex items-center justify-between pt-3 border-t border-[var(--border-subtle)]">
                <Button
                  type="submit"
                  variant="primary"
                  size="compact"
                  isLoading={isSubmitting}
                  className="h-9 px-6 text-[12.5px] font-medium rounded-[6px] bg-[#2a2d37] hover:bg-[#363a45] text-[#e0e2eb] border border-[#3e4250]"
                >
                  Save
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="compact"
                  onClick={() => setShowAddSheet(false)}
                  className="h-9 px-4 text-[12.5px] font-medium text-[#8e92a4] hover:text-white"
                >
                  Cancel
                </Button>
              </div>
            </form>
          ) : (
            <form onSubmit={handleBulkAddSubmit} className="space-y-4">
              {bulkError && (
                <div className="flex items-center gap-2 p-2.5 rounded-[6px] bg-rose-950/20 border border-rose-600/30 text-rose-300 text-[12px]">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  {bulkError}
                </div>
              )}

              <div>
                <label className="block text-[12px] font-semibold text-[#e0e2eb] mb-1.5">
                  Name
                </label>
                <input
                  type="text"
                  value={bulkPrefix}
                  onChange={(e) => setBulkPrefix(e.target.value)}
                  placeholder="Key"
                  className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="block text-[12px] font-semibold text-[#e0e2eb]">
                    API Key
                  </label>
                  <span className="text-[11px] font-mono text-[#ea580c]">
                    {
                      bulkText
                        .split('\n')
                        .map((l) => l.trim())
                        .filter((l) => l.length > 0).length
                    }{' '}
                    keys detected
                  </span>
                </div>
                <textarea
                  required
                  rows={6}
                  value={bulkText}
                  onChange={(e) => setBulkText(e.target.value)}
                  placeholder="Paste multiple API keys here (one per line)..."
                  className="w-full p-3 text-[12.5px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] placeholder:text-[#63687b] focus:outline-none focus:border-[#ea580c] transition-colors leading-relaxed"
                />
              </div>

              <div>
                <label className="block text-[12px] font-semibold text-[#e0e2eb] mb-1.5">
                  Priority
                </label>
                <input
                  type="number"
                  min="1"
                  max="100"
                  value={bulkPriority}
                  onChange={(e) => setBulkPriority(parseInt(e.target.value) || 1)}
                  className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                />
              </div>

              <div>
                <label className="block text-[12px] font-semibold text-[#e0e2eb] mb-1.5">
                  Proxy Pool
                </label>
                <select
                  value={bulkProxyId}
                  onChange={(e) => setBulkProxyId(e.target.value)}
                  className="w-full h-9 px-3 text-[13px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                >
                  <option value="">None</option>
                  {proxies.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name} — {p.scheme}://{p.host}
                    </option>
                  ))}
                </select>
                <p className="text-[11px] text-[#787c8d] mt-2">
                  Legacy manual proxy fields are still accepted by API for backward compatibility.
                </p>
              </div>

              <div className="flex items-center justify-between pt-3 border-t border-[var(--border-subtle)]">
                <Button
                  type="submit"
                  variant="primary"
                  size="compact"
                  isLoading={isBulkAdding}
                  className="h-9 px-6 text-[12.5px] font-medium rounded-[6px] bg-[#2a2d37] hover:bg-[#363a45] text-[#e0e2eb] border border-[#3e4250]"
                >
                  Save
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="compact"
                  onClick={() => setShowAddSheet(false)}
                  className="h-9 px-4 text-[12.5px] font-medium text-[#8e92a4] hover:text-white"
                >
                  Cancel
                </Button>
              </div>
            </form>
          )}
        </div>
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
              placeholder="Kunci"
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

      <BottomSheet
        isOpen={showOAuthModal}
        onClose={() => setShowOAuthModal(false)}
        title={`Connect ${provider?.name || id} via OAuth`}
        description="Masuk dengan akun provider Anda secara aman menggunakan protokol OAuth 2.0 PKCE."
        maxWidth="max-w-md"
      >
        <div className="space-y-4">
          <div className="p-3.5 rounded-[8px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-2 text-[12px]">
            <div className="flex items-center gap-2 text-[var(--status-success)] font-medium">
              <Shield className="w-4 h-4" />
              <span>Proteksi Kredensial Maksimal (PKCE + AES-256-GCM)</span>
            </div>
            <p className="text-[var(--text-secondary)] leading-relaxed">
              Jendela otorisasi resmi Google / Provider telah dibuka di tab baru. Silakan izinkan akses untuk menghubungkan akun.
            </p>
          </div>

          {oauthError && (
            <div className="p-3 rounded-[6px] bg-rose-950/40 border border-rose-600/40 text-rose-300 text-[12px]">
              {oauthError}
            </div>
          )}

          <div className="flex flex-col gap-2">
            <Button
              type="button"
              variant="primary"
              onClick={() => {
                if (oauthData?.auth_url) {
                  window.open(oauthData.auth_url, '_blank', 'width=600,height=700')
                }
              }}
              className="w-full justify-center"
              leftIcon={<ExternalLink className="w-3.5 h-3.5" />}
            >
              Buka Ulang Jendela Otorisasi
            </Button>
          </div>

          <div className="pt-3 border-t border-[var(--border-subtle)] space-y-2">
            <label className="text-[12px] font-medium text-[var(--text-secondary)] block">
              Atau Tempel Kode Otorisasi / Callback URL Manual:
            </label>
            <input
              type="text"
              placeholder="Contoh: 4/0AQl... atau http://localhost:8080/api/accounts/oauth/callback?code=..."
              value={manualCode}
              onChange={(e) => setManualCode(e.target.value)}
              className="w-full h-8.5 px-3 text-[12px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)]"
            />
            <Button
              type="button"
              variant="secondary"
              size="compact"
              disabled={!manualCode.trim()}
              isLoading={isCompletingOAuth}
              onClick={() => handleCompleteOAuth()}
              className="w-full justify-center mt-2"
            >
              Selesaikan Koneksi
            </Button>
          </div>
        </div>
      </BottomSheet>
    </div>
  )
}
