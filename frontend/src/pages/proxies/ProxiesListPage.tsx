import { useEffect, useState, useRef } from 'react'
import {
  Globe,
  Plus,
  Trash2,
  RefreshCw,
  CheckCircle,
  XCircle,
  FlaskConical,
  Pencil,
  Upload,
  Rocket,
  ChevronDown,
  Cloud,
  UploadCloud,
  Terminal,
  RotateCw,
  Shield,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { ProxyProfile } from '../../types/api.ts'
import { AddProxyPoolModal } from './AddProxyPoolModal.tsx'
import { DeployRelayModal, RelayPlatform } from './DeployRelayModal.tsx'
import { BatchImportModal } from './BatchImportModal.tsx'
import { SmartRotateModal } from './SmartRotateModal.tsx'

interface TestStatus {
  testing: boolean
  ok?: boolean
  latencyMs?: number
  error?: string
}

export function ProxiesListPage() {
  const [proxies, setProxies] = useState<ProxyProfile[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [testStatuses, setTestStatuses] = useState<Record<string, TestStatus>>({})

  const [isAddOpen, setIsAddOpen] = useState(false)
  const [editingProxy, setEditingProxy] = useState<ProxyProfile | null>(null)
  const [isDeployOpen, setIsDeployOpen] = useState(false)
  const [deployPlatform, setDeployPlatform] = useState<RelayPlatform>('cloudflare')
  const [isBatchOpen, setIsBatchOpen] = useState(false)
  const [isSmartRotateOpen, setIsSmartRotateOpen] = useState(false)
  const [isRelayDropdownOpen, setIsRelayDropdownOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  const fetchProxies = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.get<ProxyProfile[]>('/api/proxies')
      setProxies(data || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load proxies'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchProxies()
  }, [])

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsRelayDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleTestConnection = async (id: string) => {
    setTestStatuses((prev) => ({ ...prev, [id]: { testing: true } }))
    try {
      const res = await api.post<{ ok: boolean; latency_ms?: number; error?: string }>(
        `/api/proxies/${id}/test`
      )
      setTestStatuses((prev) => ({
        ...prev,
        [id]: { testing: false, ok: res.ok, latencyMs: res.latency_ms, error: res.error },
      }))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Connection test failed'
      setTestStatuses((prev) => ({
        ...prev,
        [id]: { testing: false, ok: false, error: msg },
      }))
    }
  }

  const handleToggleEnabled = async (p: ProxyProfile) => {
    const nextState = !p.enabled
    setProxies((prev) =>
      prev.map((item) => (item.id === p.id ? { ...item, enabled: nextState } : item))
    )
    try {
      if (nextState) {
        await api.post(`/api/proxies/${p.id}/enable`, {})
      } else {
        await api.post(`/api/proxies/${p.id}/disable`, {})
      }
    } catch {
      setProxies((prev) =>
        prev.map((item) => (item.id === p.id ? { ...item, enabled: !nextState } : item))
      )
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/api/proxies/${id}`)
      setProxies((prev) => prev.filter((p) => p.id !== id))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete proxy profile'
      setError(msg)
    }
  }

  const columns: Column<ProxyProfile>[] = [
    {
      key: 'name',
      title: 'Profile Name',
      render: (item) => (
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded-[5px] bg-[#232630] border border-[#333745] flex items-center justify-center shrink-0">
            {item.relay_type === 'cloudflare' ? (
              <Cloud className="w-3.5 h-3.5 text-[#f6821f]" />
            ) : item.relay_type === 'vercel' ? (
              <UploadCloud className="w-3.5 h-3.5 text-[#0070f3]" />
            ) : item.relay_type === 'deno' ? (
              <Terminal className="w-3.5 h-3.5 text-[#10b981]" />
            ) : (
              <Globe className="w-3.5 h-3.5 text-[var(--brand-text)]" />
            )}
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <span className="font-semibold text-[13px] text-[var(--text-primary)] truncate">
                {item.name}
              </span>
              {item.strict_proxy && (
                <span
                  title="Strict Proxy: Fails request if unreachable"
                  className="inline-flex items-center gap-0.5 text-[10px] px-1 py-0.2 rounded bg-amber-500/15 border border-amber-500/30 text-amber-400 font-mono"
                >
                  <Shield className="w-2.5 h-2.5" />
                  strict
                </span>
              )}
            </div>
            {item.no_proxy && (
              <div className="text-[10.5px] font-mono text-[var(--text-muted)] truncate max-w-[200px]">
                bypass: {item.no_proxy}
              </div>
            )}
          </div>
        </div>
      ),
    },
    {
      key: 'scheme',
      title: 'Protocol',
      width: '110px',
      render: (item) => (
        <span className="font-mono uppercase text-[10.5px] font-semibold text-[var(--brand-text)] bg-[var(--brand-subtle)] px-2 py-0.5 rounded border border-[var(--border-subtle)]">
          {item.relay_type && item.relay_type !== 'standard'
            ? item.relay_type
            : item.scheme}
        </span>
      ),
    },
    {
      key: 'host',
      title: 'Egress Host / Target URL',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--text-secondary)] select-all truncate block max-w-[280px]">
          {item.host}:{item.port}
        </span>
      ),
    },
    {
      key: 'latency',
      title: 'Live Connectivity',
      width: '160px',
      render: (item) => {
        const test = testStatuses[item.id]
        if (test?.testing) {
          return (
            <span className="text-[11px] font-mono text-[var(--text-muted)] inline-flex items-center gap-1.5 animate-pulse">
              <span className="h-2 w-2 rounded-full bg-[var(--brand-primary)]" />
              Testing...
            </span>
          )
        }
        if (test?.ok) {
          return (
            <span className="text-[11px] font-mono text-[var(--status-success)] inline-flex items-center gap-1">
              <CheckCircle className="w-3.5 h-3.5" />
              <span>Online{test.latencyMs !== undefined ? ` (${test.latencyMs}ms)` : ''}</span>
            </span>
          )
        }
        if (test?.ok === false) {
          return (
            <span
              title={test.error || 'Connection failed'}
              className="text-[11px] font-mono text-[var(--status-danger)] inline-flex items-center gap-1"
            >
              <XCircle className="w-3.5 h-3.5" />
              <span>Failed</span>
            </span>
          )
        }
        return (
          <span
            className={`text-[11px] font-mono px-2 py-0.5 rounded border ${
              item.enabled
                ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/25'
                : 'bg-zinc-500/10 text-zinc-400 border-zinc-500/25'
            }`}
          >
            {item.enabled ? 'Ready' : 'Inactive'}
          </span>
        )
      },
    },
    {
      key: 'actions',
      title: 'Actions',
      width: '180px',
      render: (item) => {
        const isTestingThis = testStatuses[item.id]?.testing
        return (
          <div className="flex items-center gap-2.5" onClick={(e) => e.stopPropagation()}>
            <button
              type="button"
              onClick={() => handleToggleEnabled(item)}
              title={item.enabled ? 'Disable proxy' : 'Enable proxy'}
              className={`relative inline-flex h-4 w-7 shrink-0 cursor-pointer rounded-full transition-colors duration-200 ease-in-out focus:outline-none ${
                item.enabled ? 'bg-[#ff6940]' : 'bg-[#374151]'
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out mt-0.25 ${
                  item.enabled ? 'translate-x-3.25' : 'translate-x-0.25'
                }`}
              />
            </button>

            <button
              type="button"
              disabled={isTestingThis}
              onClick={() => handleTestConnection(item.id)}
              title="Test connection"
              className="p-1 rounded text-[#9ca3af] hover:text-[#f3f4f6] hover:bg-[#2e323e] transition-colors cursor-pointer disabled:opacity-50"
            >
              <FlaskConical className={`w-4 h-4 ${isTestingThis ? 'animate-spin' : ''}`} />
            </button>

            <button
              type="button"
              onClick={() => {
                setEditingProxy(item)
                setIsAddOpen(true)
              }}
              title="Edit proxy profile"
              className="p-1 rounded text-[#9ca3af] hover:text-[#f3f4f6] hover:bg-[#2e323e] transition-colors cursor-pointer"
            >
              <Pencil className="w-3.5 h-3.5" />
            </button>

            <InlineConfirm
              trigger={
                <button
                  type="button"
                  title="Delete proxy"
                  className="p-1 rounded text-[#ef4444] hover:brightness-125 transition-all cursor-pointer"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              }
              confirmText="Delete?"
              onConfirm={() => handleDelete(item.id)}
            />
          </div>
        )
      },
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Outbound Proxies"
        description="Egress proxy pools, edge relays, and IP isolation tunnels for upstream AI providers."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Outbound Proxies' },
        ]}
        metadata={
          <span className="flex items-center gap-1.5 text-[11.5px] text-[var(--status-success)]">
            <span className="w-2 h-2 rounded-full bg-[var(--status-success)] animate-pulse" />
            {proxies.filter((p) => p.enabled).length} active of {proxies.length} pools
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={fetchProxies}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>

            <Button
              variant="secondary"
              size="compact"
              onClick={() => setIsSmartRotateOpen(true)}
              leftIcon={<RotateCw className="w-3.5 h-3.5 text-[#38bdf8]" />}
            >
              Smart Rotate
            </Button>

            <div className="relative" ref={dropdownRef}>
              <button
                type="button"
                onClick={() => setIsRelayDropdownOpen(!isRelayDropdownOpen)}
                className="inline-flex items-center gap-1.5 h-[32px] px-3 text-[12px] font-medium rounded-[6px] bg-[#232630] hover:bg-[#2d313e] text-[#f3f4f6] border border-[#393d4a] transition-colors cursor-pointer"
              >
                <Rocket className="w-3.5 h-3.5 text-[#ff6940]" />
                <span>Deploy Relay</span>
                <ChevronDown
                  className={`w-3.5 h-3.5 text-[#9ca3af] transition-transform duration-150 ${
                    isRelayDropdownOpen ? 'rotate-180' : ''
                  }`}
                />
              </button>

              {isRelayDropdownOpen && (
                <div className="absolute left-0 mt-1 w-44 rounded-[8px] bg-[#1a1c23] border border-[#2e323e] shadow-xl py-1 z-30 animate-in fade-in duration-100">
                  <button
                    type="button"
                    onClick={() => {
                      setDeployPlatform('cloudflare')
                      setIsDeployOpen(true)
                      setIsRelayDropdownOpen(false)
                    }}
                    className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-[#e5e7eb] hover:bg-[#282b37] transition-colors text-left cursor-pointer"
                  >
                    <Cloud className="w-4 h-4 text-[#f6821f] shrink-0" />
                    <span>Cloudflare Relay</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setDeployPlatform('vercel')
                      setIsDeployOpen(true)
                      setIsRelayDropdownOpen(false)
                    }}
                    className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-[#e5e7eb] hover:bg-[#282b37] transition-colors text-left cursor-pointer"
                  >
                    <UploadCloud className="w-4 h-4 text-[#0070f3] shrink-0" />
                    <span>Vercel Relay</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setDeployPlatform('deno')
                      setIsDeployOpen(true)
                      setIsRelayDropdownOpen(false)
                    }}
                    className="w-full flex items-center gap-2.5 px-3 py-2 text-[12px] text-[#e5e7eb] hover:bg-[#282b37] transition-colors text-left cursor-pointer"
                  >
                    <Terminal className="w-4 h-4 text-[#10b981] shrink-0" />
                    <span>Deno Relay</span>
                  </button>
                </div>
              )}
            </div>

            <Button
              variant="secondary"
              size="compact"
              onClick={() => setIsBatchOpen(true)}
              leftIcon={<Upload className="w-3.5 h-3.5" />}
            >
              Batch Import
            </Button>

            <button
              type="button"
              onClick={() => {
                setEditingProxy(null)
                setIsAddOpen(true)
              }}
              className="inline-flex items-center gap-1.5 h-[32px] px-3.5 text-[12.5px] font-semibold rounded-[6px] bg-[#ff6940] hover:bg-[#ff7b57] text-white shadow-xs transition-all cursor-pointer"
            >
              <Plus className="w-4 h-4" />
              <span>Add Proxy Pool</span>
            </button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchProxies} />}

      <DataTable
        columns={columns}
        data={proxies}
        isLoading={isLoading}
        keyExtractor={(p) => p.id}
        emptyMessage="No outbound proxies configured. Click '+ Add Proxy Pool' or 'Deploy Relay' to register egress network tunnels."
      />

      <AddProxyPoolModal
        isOpen={isAddOpen}
        onClose={() => {
          setIsAddOpen(false)
          setEditingProxy(null)
        }}
        onSuccess={fetchProxies}
        initialData={editingProxy}
      />

      <DeployRelayModal
        isOpen={isDeployOpen}
        onClose={() => setIsDeployOpen(false)}
        onSuccess={fetchProxies}
        defaultPlatform={deployPlatform}
      />

      <BatchImportModal
        isOpen={isBatchOpen}
        onClose={() => setIsBatchOpen(false)}
        onSuccess={fetchProxies}
      />

      <SmartRotateModal
        isOpen={isSmartRotateOpen}
        onClose={() => setIsSmartRotateOpen(false)}
        onSuccess={fetchProxies}
      />
    </div>
  )
}
