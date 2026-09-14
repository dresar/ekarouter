import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Server,
  Activity,
  AlertTriangle,
  Zap,
  ArrowUpRight,
  Terminal,
  Copy,
  Check,
  RefreshCw,
  Cpu,
  Layers,
  Send,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { MetricCard } from '../../components/ui/MetricCard.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { Provider, Account, UsageSummary, Route } from '../../types/api.ts'
import { formatNumber, formatCompactNumber } from '../../utils/formatters.ts'

export function OverviewPage() {
  const [providers, setProviders] = useState<Provider[]>([])
  const [accounts, setAccounts] = useState<Account[]>([])
  const [routes, setRoutes] = useState<Route[]>([])
  const [usage, setUsage] = useState<UsageSummary | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [copiedSnippet, setCopiedSnippet] = useState(false)

  const [testModel, setTestModel] = useState('gpt-4o')
  const [testPrompt, setTestPrompt] = useState('ping')
  const [testOutput, setTestOutput] = useState<string | null>(null)
  const [isRunningTest, setIsRunningTest] = useState(false)

  const fetchData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [provData, accData, routesData, usageData] = await Promise.all([
        api.get<Provider[]>('/api/providers').catch(() => []),
        api.get<Account[]>('/api/accounts').catch(() => []),
        api.get<Route[]>('/api/routes').catch(() => []),
        api.get<UsageSummary>('/api/usage').catch(() => null),
      ])
      setProviders(provData || [])
      setAccounts(accData || [])
      setRoutes(routesData || [])
      setUsage(usageData)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load telemetry'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const activeProvidersCount = providers.filter((p) => p.enabled).length
  const coolingDownAccounts = accounts.filter((a) => a.state === 'cooling_down')
  const totalRequests = usage?.total_requests || 0
  const promptTokens = usage?.prompt_tokens ?? 0
  const compTokens = usage?.completion_tokens ?? 0
  const totalTokens = usage?.total_tokens ?? (promptTokens + compTokens)

  const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL as string) || (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080')
  const curlSnippet = `curl -X POST ${apiBaseUrl}/v1/chat/completions \\
  -H "Authorization: Bearer eka_live_default" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${testModel}",
    "messages": [{"role": "user", "content": "${testPrompt}"}]
  }'`

  const handleCopySnippet = async () => {
    try {
      await navigator.clipboard.writeText(curlSnippet)
      setCopiedSnippet(true)
      setTimeout(() => setCopiedSnippet(false), 2000)
    } catch {}
  }

  const handleRunIngressTest = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsRunningTest(true)
    setTestOutput(null)
    const start = performance.now()
    try {
      const res = await api.post<Record<string, unknown>>('/v1/chat/completions', {
        model: testModel,
        messages: [{ role: 'user', content: testPrompt }],
        stream: false,
      })
      const elapsed = Math.round(performance.now() - start)
      setTestOutput(`HTTP/1.1 200 OK (${elapsed}ms)\n` + JSON.stringify(res, null, 2))
    } catch (err: unknown) {
      const elapsed = Math.round(performance.now() - start)
      setTestOutput(`Error (${elapsed}ms): ${err instanceof Error ? err.message : 'Gateway route error'}`)
    } finally {
      setIsRunningTest(false)
    }
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Command Center"
        description="Universal AI Gateway health, routing topology, and live telemetry control."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Command Center' },
        ]}
        metadata={
          <span>
            {activeProvidersCount} active backends &bull; {routes.length} routing combos
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Link to="/console">
              <Button variant="secondary" size="compact" leftIcon={<Terminal className="w-3.5 h-3.5" />}>
                Live Console
              </Button>
            </Link>
            <Button
              variant="secondary"
              size="compact"
              onClick={fetchData}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchData} />}

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <MetricCard
          label="Active Providers"
          value={`${activeProvidersCount} / ${providers.length}`}
          subtext="Configured upstream AI backends"
          icon={<Server className="w-4 h-4" />}
        />
        <MetricCard
          label="Total Requests"
          value={formatNumber(totalRequests)}
          subtext={`${usage?.success_requests || 0} served successfully`}
          icon={<Activity className="w-4 h-4" />}
        />
        <MetricCard
          label="Tokens Processed"
          value={formatCompactNumber(totalTokens)}
          subtext={`${formatCompactNumber(usage?.prompt_tokens || 0)} prompt tokens`}
          icon={<Zap className="w-4 h-4" />}
        />
        <MetricCard
          label="Active Cooldowns"
          value={coolingDownAccounts.length}
          subtext={coolingDownAccounts.length > 0 ? 'Accounts in rate-limit backoff' : 'All circuits healthy'}
          icon={<AlertTriangle className="w-4 h-4" />}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        <div className="lg:col-span-7 space-y-4">
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4">
            <div className="flex items-center justify-between mb-3">
              <div>
                <h3 className="text-[13.5px] font-semibold text-[var(--text-primary)]">
                  Provider Availability Matrix
                </h3>
                <p className="text-[11px] text-[var(--text-muted)] mt-0.5">
                  Registered upstream model engines
                </p>
              </div>
              <Link
                to="/providers"
                className="text-[11.5px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 font-medium"
              >
                <span>View All</span>
                <ArrowUpRight className="w-3.5 h-3.5" />
              </Link>
            </div>

            <div className="divide-y divide-[var(--border-subtle)]">
              {(providers.length > 0 ? providers.slice(0, 6) : [
                { id: 'openai', key: 'openai', name: 'OpenAI', kind: 'openai', base_url: 'https://api.openai.com/v1', enabled: true },
                { id: 'anthropic', key: 'anthropic', name: 'Anthropic Claude', kind: 'anthropic', base_url: 'https://api.anthropic.com', enabled: true },
                { id: 'google', key: 'gemini', name: 'Google Gemini', kind: 'gemini', base_url: 'https://generativelanguage.googleapis.com', enabled: true },
                { id: 'deepseek', key: 'deepseek', name: 'DeepSeek', kind: 'openai', base_url: 'https://api.deepseek.com', enabled: true },
                { id: 'qwen', key: 'qwen', name: 'Alibaba Qwen', kind: 'openai', base_url: 'https://dashscope.aliyuncs.com', enabled: true },
              ]).map((provider) => (
                <div
                  key={provider.id}
                  className="py-2.5 flex items-center justify-between text-[12.5px]"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <ProviderLogo providerId={provider.id} name={provider.name} size="sm" />
                    <div className="flex flex-col min-w-0">
                      <span className="font-semibold text-[var(--text-primary)] truncate">
                        {provider.name}
                      </span>
                      <span className="text-[10.5px] font-mono text-[var(--text-muted)] truncate max-w-[220px]">
                        {provider.base_url}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <span className="text-[10.5px] font-mono text-[var(--text-muted)] uppercase hidden sm:inline">
                      {provider.kind}
                    </span>
                    <StatusBadge variant={provider.enabled ? 'healthy' : 'disabled'}>
                      {provider.enabled ? 'Active' : 'Disabled'}
                    </StatusBadge>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {coolingDownAccounts.length > 0 && (
            <div className="bg-amber-950/20 border border-amber-600/30 rounded-[8px] p-4">
              <div className="flex items-center gap-2 mb-2 text-amber-300">
                <AlertTriangle className="w-4 h-4 shrink-0" />
                <h4 className="text-[13px] font-semibold">Active Failover Circuits</h4>
              </div>
              <p className="text-[12px] text-[var(--text-secondary)] mb-3">
                The following provider accounts hit rate limits (429) or authentication errors and are temporarily in cooldown:
              </p>
              <div className="space-y-1.5">
                {coolingDownAccounts.map((acc) => (
                  <div
                    key={acc.id}
                    className="flex items-center justify-between bg-[var(--bg-card)] px-3 py-2 rounded-[5px] text-[12px] border border-[var(--border-subtle)]"
                  >
                    <span className="font-medium text-[var(--text-primary)]">{acc.name}</span>
                    <span className="font-mono text-[11px] text-amber-400">
                      {acc.cooldown_until || 'In Backoff'}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        <div className="lg:col-span-5 space-y-4">
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <Terminal className="w-4 h-4 text-[var(--brand-text)]" />
                <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
                  Gateway Ingress Runner
                </h3>
              </div>
              <button
                type="button"
                onClick={handleCopySnippet}
                className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors inline-flex items-center gap-1 text-[11px]"
              >
                {copiedSnippet ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-[var(--status-success)]" />
                    <span className="text-[var(--status-success)]">Copied</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>Copy cURL</span>
                  </>
                )}
              </button>
            </div>

            <form onSubmit={handleRunIngressTest} className="space-y-2.5 mb-3">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="text-[10.5px] font-medium text-[var(--text-muted)] block mb-1">
                    Route Model
                  </label>
                  <input
                    type="text"
                    value={testModel}
                    onChange={(e) => setTestModel(e.target.value)}
                    className="w-full px-2.5 py-1 text-[11.5px] font-mono rounded-[5px]"
                  />
                </div>
                <div>
                  <label className="text-[10.5px] font-medium text-[var(--text-muted)] block mb-1">
                    Test Prompt
                  </label>
                  <input
                    type="text"
                    value={testPrompt}
                    onChange={(e) => setTestPrompt(e.target.value)}
                    className="w-full px-2.5 py-1 text-[11.5px] font-mono rounded-[5px]"
                  />
                </div>
              </div>

              <Button
                type="submit"
                variant="primary"
                size="compact"
                isLoading={isRunningTest}
                leftIcon={<Send className="w-3.5 h-3.5" />}
                className="w-full"
              >
                Execute Test Ingress
              </Button>
            </form>

            {testOutput && (
              <pre className="p-2.5 rounded-[6px] bg-[#07090E] border border-[var(--border-subtle)] font-mono text-[10.5px] text-emerald-400 overflow-x-auto max-h-40 leading-relaxed">
                {testOutput}
              </pre>
            )}

            {!testOutput && (
              <pre className="bg-[#07090E] p-2.5 rounded-[6px] border border-[var(--border-subtle)] text-[10.5px] font-mono text-[var(--text-muted)] overflow-x-auto select-all leading-relaxed">
                {curlSnippet}
              </pre>
            )}
          </div>

          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4">
            <h3 className="text-[13px] font-semibold text-[var(--text-primary)] mb-2.5">
              Infrastructure Shortcuts
            </h3>
            <div className="grid grid-cols-2 gap-2 text-[12px]">
              <Link
                to="/routing"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 hover:bg-[var(--border-strong)]/40 transition-colors flex flex-col border border-[var(--border-subtle)]"
              >
                <div className="flex items-center gap-1.5 mb-1">
                  <Cpu className="w-3.5 h-3.5 text-[var(--brand-text)]" />
                  <span className="font-semibold text-[var(--text-primary)]">Routing Combos</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">Priority failover chains</span>
              </Link>
              <Link
                to="/vault"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 hover:bg-[var(--border-strong)]/40 transition-colors flex flex-col border border-[var(--border-subtle)]"
              >
                <div className="flex items-center gap-1.5 mb-1">
                  <Layers className="w-3.5 h-3.5 text-[var(--brand-text)]" />
                  <span className="font-semibold text-[var(--text-primary)]">Credential Vault</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">AES-256 encrypted keys</span>
              </Link>
              <Link
                to="/token-saver"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 hover:bg-[var(--border-strong)]/40 transition-colors flex flex-col border border-[var(--border-subtle)]"
              >
                <div className="flex items-center gap-1.5 mb-1">
                  <Zap className="w-3.5 h-3.5 text-amber-400" />
                  <span className="font-semibold text-[var(--text-primary)]">Token Saver</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">Prompt compaction</span>
              </Link>
              <Link
                to="/api-keys"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 hover:bg-[var(--border-strong)]/40 transition-colors flex flex-col border border-[var(--border-subtle)]"
              >
                <div className="flex items-center gap-1.5 mb-1">
                  <Activity className="w-3.5 h-3.5 text-emerald-400" />
                  <span className="font-semibold text-[var(--text-primary)]">Client API Keys</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">Ingress tokens (eka_live_*)</span>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
