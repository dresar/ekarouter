import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  BarChart3,
  Zap,
  RefreshCw,
  Clock,
  Terminal,
  ShieldCheck,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { MetricCard } from '../../components/ui/MetricCard.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { EmptyState } from '../../components/ui/EmptyState.tsx'
import { api } from '../../api/client.ts'
import {
  UsageAdminSummary,
  UsageSummary,
  ProviderUsageItem,
  CredentialUsageItem,
  ProjectUsageItem,
} from '../../types/api.ts'
import { formatCompactNumber, formatNumber, formatDate } from '../../utils/formatters.ts'

export function UsagePage() {
  const [adminUsage, setAdminUsage] = useState<UsageAdminSummary | null>(null)
  const [platformUsage, setPlatformUsage] = useState<UsageSummary | null>(null)
  const [providerUsage, setProviderUsage] = useState<ProviderUsageItem[]>([])
  const [credentialUsage, setCredentialUsage] = useState<CredentialUsageItem[]>([])
  const [projectUsage, setProjectUsage] = useState<ProjectUsageItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadUsage = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [adminRes, platformRes, provRes, credRes, projRes] = await Promise.allSettled([
        api.get<UsageAdminSummary>('/api/usage'),
        api.get<UsageSummary>('/api/v1/usage/summary'),
        api.get<ProviderUsageItem[]>('/api/v1/usage/providers'),
        api.get<CredentialUsageItem[]>('/api/v1/usage/credentials'),
        api.get<ProjectUsageItem[]>('/api/v1/usage/projects'),
      ])

      if (adminRes.status === 'fulfilled') setAdminUsage(adminRes.value)
      if (platformRes.status === 'fulfilled') setPlatformUsage(platformRes.value)
      if (provRes.status === 'fulfilled' && Array.isArray(provRes.value)) setProviderUsage(provRes.value)
      if (credRes.status === 'fulfilled' && Array.isArray(credRes.value)) setCredentialUsage(credRes.value)
      if (projRes.status === 'fulfilled' && Array.isArray(projRes.value)) setProjectUsage(projRes.value)

      if (adminRes.status === 'rejected' && platformRes.status === 'rejected') {
        throw new Error('Failed to load usage telemetry from backend')
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load usage telemetry'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadUsage()
  }, [])

  const totalRequests = adminUsage?.total_requests ?? platformUsage?.total_requests ?? 0
  const totalTokens = adminUsage?.total_tokens ?? 0
  const promptTokens = adminUsage?.prompt_tokens ?? 0
  const outputTokens = adminUsage?.output_tokens ?? 0
  const avgLatency = adminUsage?.avg_latency_ms ?? 0
  const totalErrors = platformUsage?.total_errors ?? 0
  const activeCredentials = platformUsage?.active_credentials ?? credentialUsage.length

  const hasAnyTelemetry =
    totalRequests > 0 ||
    providerUsage.length > 0 ||
    credentialUsage.some((c) => c.request_count > 0)

  return (
    <div className="space-y-4">
      <PageHeader
        title="Usage & Analytics"
        description="Historical telemetry, token throughput, and provider workload distribution."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Usage & Analytics' },
        ]}
        metadata={
          <span>
            {formatCompactNumber(totalRequests)} requests &bull; {formatCompactNumber(totalTokens)} tokens
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
              onClick={loadUsage}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadUsage} />}

      {/* Top 4 Infrastructure Metrics */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <MetricCard
          label="Total Ingress Requests"
          value={formatNumber(totalRequests)}
          subtext={totalErrors > 0 ? `${totalErrors} failures recorded` : 'All requests successful'}
          icon={<BarChart3 className="w-4 h-4" />}
        />
        <MetricCard
          label="Total Tokens Processed"
          value={formatCompactNumber(totalTokens)}
          subtext={`${formatCompactNumber(promptTokens)} in • ${formatCompactNumber(outputTokens)} out`}
          icon={<Zap className="w-4 h-4" />}
        />
        <MetricCard
          label="Mean Gateway Latency"
          value={avgLatency > 0 ? `${avgLatency} ms` : '—'}
          subtext="Average upstream response time"
          icon={<Clock className="w-4 h-4" />}
        />
        <MetricCard
          label="Active Credentials"
          value={String(activeCredentials)}
          subtext={`${providerUsage.length} providers with credentials`}
          icon={<ShieldCheck className="w-4 h-4" />}
        />
      </div>

      {!hasAnyTelemetry && !isLoading ? (
        <EmptyState
          icon={<BarChart3 className="w-5 h-5" />}
          title="No telemetry recorded yet"
          description="Inference requests dispatched through /v1/chat/completions will record throughput metrics."
          action={
            <div className="flex items-center gap-2">
              <Link to="/console">
                <Button variant="primary" size="compact" leftIcon={<Terminal className="w-3.5 h-3.5" />}>
                  Open Live Console
                </Button>
              </Link>
              <Button variant="secondary" size="compact" onClick={loadUsage}>
                Check Again
              </Button>
            </div>
          }
        />
      ) : (
        <div className="space-y-4">
          {/* Provider Workload Share */}
          <div className="bg-[var(--bg-surface)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
                  Provider Workload Share
                </h3>
                <p className="text-[11px] text-[var(--text-muted)]">
                  Aggregate requests, credential count, and error totals grouped by provider.
                </p>
              </div>
              <span className="text-[11px] font-mono text-[var(--text-muted)]">
                {providerUsage.length} active providers
              </span>
            </div>

            {providerUsage.length === 0 ? (
              <div className="py-6 text-center text-[12px] text-[var(--text-muted)] border border-dashed border-[var(--border-subtle)] rounded-[6px]">
                No provider-level telemetry recorded.
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                {providerUsage.map((p) => {
                  const errorRate =
                    p.total_requests > 0
                      ? ((p.total_errors / p.total_requests) * 100).toFixed(1)
                      : '0.0'
                  return (
                    <div
                      key={p.provider_id}
                      className="p-3 rounded-[6px] bg-[var(--bg-card)] border border-[var(--border-subtle)] flex items-center justify-between"
                    >
                      <div className="flex items-center gap-2.5 min-w-0">
                        <ProviderLogo providerId={p.provider_id} size="md" />
                        <div className="flex flex-col min-w-0">
                          <span className="text-[12.5px] font-semibold text-[var(--text-primary)] truncate">
                            {p.provider_id}
                          </span>
                          <span className="text-[10.5px] font-mono text-[var(--text-muted)]">
                            {p.active_credentials} credentials configured
                          </span>
                        </div>
                      </div>

                      <div className="text-right shrink-0 ml-2">
                        <span className="text-[13px] font-mono font-bold text-[var(--text-primary)] block">
                          {formatNumber(p.total_requests)}
                        </span>
                        <span
                          className={`text-[10.5px] font-mono ${
                            p.total_errors > 0 ? 'text-[var(--status-danger)]' : 'text-[var(--status-success)]'
                          }`}
                        >
                          {p.total_errors > 0 ? `${errorRate}% err` : '0 errors'}
                        </span>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>

          {/* Credential Volume Table */}
          <div className="bg-[var(--bg-surface)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
                  Credential Usage Volume
                </h3>
                <p className="text-[11px] text-[var(--text-muted)]">
                  Traffic dispatched per individual vault key and environment.
                </p>
              </div>
              <span className="text-[11px] font-mono text-[var(--text-muted)]">
                {credentialUsage.length} credentials
              </span>
            </div>

            {credentialUsage.length === 0 ? (
              <div className="py-6 text-center text-[12px] text-[var(--text-muted)] border border-dashed border-[var(--border-subtle)] rounded-[6px]">
                No credentials registered in vault.
              </div>
            ) : (
              <div className="overflow-x-auto border border-[var(--border-subtle)] rounded-[6px]">
                <table className="w-full text-left text-[12px] border-collapse">
                  <thead>
                    <tr className="bg-[var(--bg-panel)] border-b border-[var(--border-subtle)] text-[11px] font-semibold text-[var(--text-muted)] uppercase tracking-wider">
                      <th className="py-2.5 px-3">Credential Name</th>
                      <th className="py-2.5 px-3">Environment</th>
                      <th className="py-2.5 px-3 text-right">Requests</th>
                      <th className="py-2.5 px-3 text-right">Errors</th>
                      <th className="py-2.5 px-3 text-right">Last Used</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[var(--border-subtle)] bg-[var(--bg-card)]">
                    {credentialUsage.map((c) => (
                      <tr key={c.id} className="hover:bg-[var(--bg-panel)]/40 transition-colors">
                        <td className="py-2 px-3">
                          <div className="flex items-center gap-2">
                            <ProviderLogo providerId={c.provider_id} size="sm" />
                            <div className="flex flex-col">
                              <span className="font-semibold text-[12.5px] text-[var(--text-primary)]">
                                {c.name}
                              </span>
                              <span className="font-mono text-[10px] text-[var(--text-muted)]">
                                {c.provider_id}
                              </span>
                            </div>
                          </div>
                        </td>
                        <td className="py-2 px-3">
                          <span
                            className={`inline-flex px-2 py-0.5 rounded-[4px] text-[10.5px] font-mono font-medium border ${
                              c.environment === 'production'
                                ? 'bg-rose-500/10 text-rose-400 border-rose-500/20'
                                : c.environment === 'staging'
                                ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                                : 'bg-blue-500/10 text-blue-400 border-blue-500/20'
                            }`}
                          >
                            {c.environment}
                          </span>
                        </td>
                        <td className="py-2 px-3 text-right font-mono font-medium text-[var(--text-primary)]">
                          {formatNumber(c.request_count)}
                        </td>
                        <td className="py-2 px-3 text-right font-mono">
                          <span
                            className={
                              c.error_count > 0 ? 'text-[var(--status-danger)]' : 'text-[var(--text-muted)]'
                            }
                          >
                            {c.error_count}
                          </span>
                        </td>
                        <td className="py-2 px-3 text-right font-mono text-[11px] text-[var(--text-muted)]">
                          {c.last_used_at ? formatDate(c.last_used_at) : 'Never'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* Project Workload */}
          {projectUsage.length > 0 && (
            <div className="bg-[var(--bg-surface)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3">
              <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
                Project & Workspace Distribution
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {projectUsage.map((p) => (
                  <div
                    key={p.project_id}
                    className="p-3 rounded-[6px] bg-[var(--bg-card)] border border-[var(--border-subtle)] flex items-center justify-between"
                  >
                    <div>
                      <span className="text-[12.5px] font-semibold text-[var(--text-primary)] font-mono block">
                        {p.project_id}
                      </span>
                      <span className="text-[10.5px] text-[var(--text-muted)]">
                        {p.active_credentials} active credentials
                      </span>
                    </div>
                    <div className="text-right">
                      <span className="text-[13px] font-mono font-bold text-[var(--text-primary)] block">
                        {formatNumber(p.total_requests)}
                      </span>
                      <span className="text-[10px] font-mono text-[var(--text-muted)]">
                        {p.total_errors} errors
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
