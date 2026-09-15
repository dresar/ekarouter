import { useEffect, useState, FormEvent } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  KeyRound,
  RotateCw,
  Activity,
  CheckCircle,
  AlertTriangle,
  AlertCircle,
  Save,
  Power,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { SecretViewer } from '../../components/ui/SecretViewer.tsx'
import { RightDrawer } from '../../components/ui/RightDrawer.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { VaultCredential } from '../../types/api.ts'
import { formatDate, formatNumber } from '../../utils/formatters.ts'

export function VaultDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [credential, setCredential] = useState<VaultCredential | null>(null)
  const [error, setError] = useState<string | null>(null)

  const [isTesting, setIsTesting] = useState(false)
  const [testResult, setTestResult] = useState<{ ok: boolean; message: string } | null>(null)

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [newSecret, setNewSecret] = useState('')
  const [isRotating, setIsRotating] = useState(false)
  const [rotateError, setRotateError] = useState<string | null>(null)

  const loadData = async () => {
    if (!id) return
    setError(null)
    try {
      const data = await api.get<VaultCredential>(`/api/v1/credentials/${id}`)
      setCredential(data)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load credential'
      setError(msg)
    }
  }

  useEffect(() => {
    loadData()
  }, [id])

  const handleTestConnection = async () => {
    if (!id) return
    setIsTesting(true)
    setTestResult(null)
    try {
      await api.post(`/api/v1/credentials/${id}/test`)
      setTestResult({ ok: true, message: 'Provider API connection verified successfully' })
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Connectivity test failed'
      setTestResult({ ok: false, message: msg })
    } finally {
      setIsTesting(false)
    }
  }

  const handleRotateSecret = async (e: FormEvent) => {
    e.preventDefault()
    if (!id || !newSecret.trim()) return
    setIsRotating(true)
    setRotateError(null)

    try {
      await api.post(`/api/v1/credentials/${id}/rotate`, { secret_value: newSecret.trim() })
      setNewSecret('')
      setDrawerOpen(false)
      await loadData()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Rotation failed'
      setRotateError(msg)
    } finally {
      setIsRotating(false)
    }
  }

  const handleToggleState = async () => {
    if (!credential) return
    const endpoint =
      credential.status === 'active'
        ? `/api/v1/credentials/${credential.id}/disable`
        : `/api/v1/credentials/${credential.id}/enable`
    try {
      await api.post(endpoint)
      await loadData()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to toggle status'
      setError(msg)
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-5">
      <PageHeader
        title={credential ? credential.name : 'Credential Inspector'}
        description={`Vault UUID: ${id || ''}`}
        breadcrumbs={[
          { label: 'Vault', to: '/vault' },
          { label: credential?.name || 'Inspect' },
        ]}
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="compact"
              onClick={() => navigate('/vault')}
              leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}
            >
              Back
            </Button>
            <Button
              variant="secondary"
              size="compact"
              isLoading={isTesting}
              onClick={handleTestConnection}
              leftIcon={<Activity className="w-3.5 h-3.5" />}
            >
              Test Connection
            </Button>
            <Button
              variant="primary"
              size="compact"
              onClick={() => setDrawerOpen(true)}
              leftIcon={<RotateCw className="w-3.5 h-3.5" />}
            >
              Rotate Secret
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadData} />}

      {testResult && (
        <div
          className={`p-3 rounded-[6px] text-[12.5px] flex items-center gap-2 ${
            testResult.ok
              ? 'bg-[var(--status-success)]/10 text-[var(--status-success)] border border-[var(--status-success)]/30'
              : 'bg-[var(--status-danger)]/10 text-[var(--status-danger)] border border-[var(--status-danger)]/30'
          }`}
        >
          {testResult.ok ? (
            <CheckCircle className="w-4 h-4 shrink-0" />
          ) : (
            <AlertTriangle className="w-4 h-4 shrink-0" />
          )}
          <span>{testResult.message}</span>
        </div>
      )}

      {credential && (
        <div className="space-y-4">
          <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-5">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-[var(--border-subtle)]">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded bg-[var(--bg-panel)] flex items-center justify-center text-[var(--brand-text)]">
                  <KeyRound className="w-5 h-5" />
                </div>
                <div>
                  <h2 className="text-[16px] font-semibold text-[var(--text-primary)]">
                    {credential.name}
                  </h2>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className="font-mono text-[11px] text-[var(--text-muted)]">
                      provider: {credential.provider_id}
                    </span>
                    <span className="text-[var(--border-strong)]">•</span>
                    <span className="font-mono text-[11px] text-[var(--text-muted)]">
                      type: {credential.credential_type}
                    </span>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <StatusBadge
                  variant={
                    credential.health_state === 'healthy'
                      ? 'healthy'
                      : credential.health_state === 'degraded'
                      ? 'degraded'
                      : credential.health_state === 'unhealthy'
                      ? 'unhealthy'
                      : 'neutral'
                  }
                >
                  {credential.health_state}
                </StatusBadge>
                <Button
                  variant={credential.status === 'active' ? 'ghost' : 'secondary'}
                  size="compact"
                  onClick={handleToggleState}
                  leftIcon={<Power className="w-3.5 h-3.5" />}
                >
                  {credential.status === 'active' ? 'Disable' : 'Enable'}
                </Button>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-4 text-[12.5px]">
              <div>
                <span className="text-[11px] font-semibold uppercase text-[var(--text-muted)] block mb-1">
                  Masked Secret
                </span>
                <SecretViewer value={credential.masked_value} />
              </div>

              <div>
                <span className="text-[11px] font-semibold uppercase text-[var(--text-muted)] block mb-1">
                  Environment & Priority
                </span>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-[12px] text-[var(--text-primary)] font-medium capitalize">
                    {credential.environment}
                  </span>
                  <span className="text-[var(--text-muted)]">•</span>
                  <span className="font-mono text-[12px] text-[var(--brand-text)]">
                    Tier {credential.priority}
                  </span>
                </div>
              </div>

              <div>
                <span className="text-[11px] font-semibold uppercase text-[var(--text-muted)] block mb-1">
                  Tags
                </span>
                <div className="flex flex-wrap gap-1">
                  {credential.tags
                    ? credential.tags.split(',').map((t, idx) => (
                        <span
                          key={idx}
                          className="px-1.5 py-0.5 bg-[var(--bg-panel)] rounded text-[11px] font-mono text-[var(--text-secondary)]"
                        >
                          {t.trim()}
                        </span>
                      ))
                    : '-'}
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-4">
              <span className="text-[11px] uppercase font-semibold text-[var(--text-muted)] block">
                Total Requests Processed
              </span>
              <span className="text-[22px] font-bold font-mono text-[var(--text-primary)] mt-1 block">
                {formatNumber(credential.request_count || 0)}
              </span>
            </div>

            <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-4">
              <span className="text-[11px] uppercase font-semibold text-[var(--text-muted)] block">
                Consecutive Errors
              </span>
              <span className="text-[22px] font-bold font-mono text-[var(--status-danger)] mt-1 block">
                {credential.error_count || 0}
              </span>
            </div>

            <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-4">
              <span className="text-[11px] uppercase font-semibold text-[var(--text-muted)] block">
                Last Activity
              </span>
              <span className="text-[13px] font-mono text-[var(--text-secondary)] mt-1 block">
                {formatDate(credential.last_used_at)}
              </span>
            </div>
          </div>
        </div>
      )}

      <RightDrawer
        isOpen={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        title="Rotate Secret"
        description="Perbarui nilai rahasia terenkripsi."
        footer={
          <>
            <Button variant="ghost" size="compact" onClick={() => setDrawerOpen(false)}>
              Batal
            </Button>
            <Button
              variant="primary"
              size="compact"
              isLoading={isRotating}
              onClick={handleRotateSecret}
              leftIcon={<Save className="w-3.5 h-3.5" />}
            >
              Rotasi
            </Button>
          </>
        }
      >
        <form onSubmit={handleRotateSecret} className="space-y-4">
          {rotateError && (
            <div className="p-2.5 rounded bg-[var(--status-danger)]/10 text-[var(--status-danger)] text-[12px] flex items-center gap-1.5">
              <AlertCircle className="w-3.5 h-3.5" />
              <span>{rotateError}</span>
            </div>
          )}

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              New Secret API Key / Token *
            </label>
            <input
              type="password"
              required
              autoComplete="new-password"
              value={newSecret}
              onChange={(e) => setNewSecret(e.target.value)}
              placeholder="Secret"
              className="w-full px-3 py-2 text-[13px] font-mono rounded-[5px] focus:outline-none"
            />
            <span className="text-[11px] text-[var(--text-muted)] mt-1 block">
              The new value will be encrypted with AES-256-GCM before replacing the existing secret.
            </span>
          </div>
        </form>
      </RightDrawer>
    </div>
  )
}
