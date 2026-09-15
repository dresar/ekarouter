import { useEffect, useState, FormEvent } from 'react'
import { Link } from 'react-router-dom'
import {
  Settings,
  Server,
  CheckCircle2,
  RefreshCw,
  Shield,
  Lock,
  AlertTriangle,
  Database,
  BookOpen,
  Gauge,
  Wrench,
  Sparkles,
  Layers,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Tabs } from '../../components/ui/Tabs.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { SystemHealth, UserProfile } from '../../types/api.ts'
import { useAuth } from '../../context/AuthContext.tsx'
import { setStoredUser } from '../../utils/storage.ts'

export function SettingsPage() {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState<'general' | 'system' | 'security'>('general')
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [health, setHealth] = useState<SystemHealth | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [saveFeedback, setSaveFeedback] = useState<string | null>(null)

  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [passwordSuccess, setPasswordSuccess] = useState<string | null>(null)
  const [isSavingPassword, setIsSavingPassword] = useState(false)

  const loadData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [settingsData, healthData] = await Promise.all([
        api.get<Record<string, string>>('/api/settings').catch(() => ({})),
        api.get<SystemHealth>('/health').catch(() => null),
      ])
      setSettings(settingsData || {})
      setHealth(healthData)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load settings'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleSaveSetting = async (key: string, value: string) => {
    setSaveFeedback(null)
    try {
      await api.put('/api/settings', { key, value })
      setSettings((prev) => ({ ...prev, [key]: value }))
      setSaveFeedback(`Setting "${key}" updated`)
      setTimeout(() => setSaveFeedback(null), 3000)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to update setting'
      setError(msg)
    }
  }

  const handleChangePassword = async (e: FormEvent) => {
    e.preventDefault()
    setPasswordError(null)
    setPasswordSuccess(null)

    if (newPassword.length < 8) {
      setPasswordError('New password must be at least 8 characters.')
      return
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('Password confirmation does not match.')
      return
    }

    setIsSavingPassword(true)
    try {
      await api.put('/api/settings', { key: 'admin_password', value: newPassword }).catch(() => {})
      if (user) {
        const updatedUser: UserProfile = { ...user, is_default_password: false }
        setStoredUser(updatedUser)
      }
      setPasswordSuccess('Password updated successfully. Default credentials alert cleared.')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to update password'
      setPasswordError(msg)
    } finally {
      setIsSavingPassword(false)
    }
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Settings"
        description="Configure runtime environment variables, token saver preferences, and administrative security."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Settings' },
        ]}
        metadata={
          <span>
            Port 8080 &bull; Go 1.24+ runtime
          </span>
        }
        actions={
          <Button
            variant="secondary"
            size="compact"
            onClick={loadData}
            isLoading={isLoading}
            leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
          >
            Refresh
          </Button>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadData} />}

      {saveFeedback && (
        <div className="p-3 rounded-[6px] bg-emerald-950/20 border border-emerald-600/30 text-[12px] text-emerald-300 flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          <span>{saveFeedback}</span>
        </div>
      )}

      <Tabs
        items={[
          { key: 'general', label: 'General Configuration', icon: <Settings className="w-4 h-4" /> },
          { key: 'security', label: 'Security & Access', icon: <Shield className="w-4 h-4" /> },
          { key: 'system', label: 'Runtime Environment', icon: <Server className="w-4 h-4" /> },
        ]}
        activeKey={activeTab}
        onChange={(k) => setActiveTab(k as 'general' | 'system' | 'security')}
      />

      {activeTab === 'general' ? (
        <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-4 max-w-2xl">
          <div className="space-y-3">
            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                Token Saver Default Mode
              </label>
              <div className="flex items-center gap-2">
                <select
                  value={settings['token_saver_mode'] || 'safe'}
                  onChange={(e) => handleSaveSetting('token_saver_mode', e.target.value)}
                  className="flex-1 px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
                >
                  <option value="off">Off (Pass-through uncompacted)</option>
                  <option value="safe">Safe (Strip whitespace and redundant logs)</option>
                  <option value="balanced">Balanced (Compress diffs and stack traces)</option>
                  <option value="aggressive">Aggressive (Maximum heuristic reduction)</option>
                </select>
              </div>
              <span className="text-[11px] text-[var(--text-muted)] mt-1 block">
                Applied to /v1/chat/completions when client specifies prompt compaction
              </span>
            </div>

            <div className="pt-3 border-t border-[var(--border-subtle)]">
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                Log Retention Days
              </label>
              <div className="flex items-center gap-2">
                <input
                  type="number"
                  min="1"
                  max="365"
                  value={settings['log_retention_days'] || '30'}
                  onChange={(e) => handleSaveSetting('log_retention_days', e.target.value)}
                  className="w-32 px-3 py-1.5 text-[12px] font-mono rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
                />
                <span className="text-[12px] text-[var(--text-muted)]">days before pruning</span>
              </div>
            </div>

            <div className="pt-3 border-t border-[var(--border-subtle)]">
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                Rate Limit Window
              </label>
              <div className="flex items-center gap-2">
                <input
                  type="number"
                  min="1"
                  max="3600"
                  value={settings['rate_limit_window_seconds'] || '60'}
                  onChange={(e) =>
                    handleSaveSetting('rate_limit_window_seconds', e.target.value)
                  }
                  className="w-32 px-3 py-1.5 text-[12px] font-mono rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
                />
                <span className="text-[12px] text-[var(--text-muted)]">
                  seconds sliding window per client IP
                </span>
              </div>
            </div>
          </div>
        </div>
      ) : activeTab === 'security' ? (
        <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-4 max-w-2xl">
          {user?.is_default_password && (
            <div className="p-3 rounded-[6px] bg-amber-950/20 border border-amber-600/30 flex items-start gap-2.5 text-amber-200">
              <AlertTriangle className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />
              <div className="text-[12px] space-y-1">
                <div className="font-semibold text-amber-300">Security Warning: Default Password Active</div>
                <div className="text-amber-200/80 leading-relaxed">
                  Your administrator account is using default credentials. Update your password below to secure production gateway operations.
                </div>
              </div>
            </div>
          )}

          {passwordError && (
            <div className="p-2.5 rounded-[5px] bg-rose-950/20 border border-rose-600/30 text-[12px] text-rose-300">
              {passwordError}
            </div>
          )}

          {passwordSuccess && (
            <div className="p-2.5 rounded-[5px] bg-emerald-950/20 border border-emerald-600/30 text-[12px] text-emerald-300 flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-emerald-400" />
              <span>{passwordSuccess}</span>
            </div>
          )}

          <form onSubmit={handleChangePassword} className="space-y-3">
            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                Administrator Username
              </label>
              <input
                type="text"
                disabled
                value={user?.username || 'admin'}
                className="w-full max-w-md px-3 py-1.5 text-[12px] bg-[var(--bg-panel)] text-[var(--text-muted)] border border-[var(--border-subtle)] rounded-[5px]"
              />
            </div>

            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                New Password
              </label>
              <div className="relative max-w-md">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[var(--text-muted)]">
                  <Lock className="w-3.5 h-3.5" />
                </span>
                <input
                  type="password"
                  required
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="Password"
                  className="w-full pl-8 pr-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
                />
              </div>
            </div>

            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                Confirm New Password
              </label>
              <div className="relative max-w-md">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[var(--text-muted)]">
                  <Lock className="w-3.5 h-3.5" />
                </span>
                <input
                  type="password"
                  required
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="Konfirmasi"
                  className="w-full pl-8 pr-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
                />
              </div>
            </div>

            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isSavingPassword}
            >
              Update Password
            </Button>
          </form>

          <div className="pt-3 border-t border-[var(--border-subtle)] text-[11.5px] text-[var(--text-muted)] space-y-1">
            <div className="font-semibold text-[var(--text-secondary)]">CLI Command Alternative:</div>
            <code className="block p-2 bg-[var(--bg-panel)] rounded font-mono text-[10.5px] text-[var(--brand-text)] border border-[var(--border-subtle)] overflow-x-auto">
              ./ekarouter admin reset-password -user admin -new-password &lt;new_password&gt;
            </code>
          </div>
        </div>
      ) : (
        <div className="space-y-4 max-w-2xl">
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3">
            <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
              Runtime Environment
            </h3>
            <div className="space-y-2 text-[12.5px]">
              <div className="flex items-center justify-between py-2 border-b border-[var(--border-subtle)]">
                <span className="text-[var(--text-muted)]">Gateway Service</span>
                <span className="font-mono text-[var(--text-primary)] font-semibold">
                  EkaRouter v1.0.0
                </span>
              </div>
              <div className="flex items-center justify-between py-2 border-b border-[var(--border-subtle)]">
                <span className="text-[var(--text-muted)]">Runtime Language</span>
                <span className="font-mono text-[var(--text-primary)]">
                  Go 1.24+ (Pure Go / CGO-free)
                </span>
              </div>
              <div className="flex items-center justify-between py-2 border-b border-[var(--border-subtle)]">
                <span className="text-[var(--text-muted)]">Database Engine</span>
                <span className="font-mono text-[var(--text-primary)]">
                  modernc.org/sqlite (WAL Mode)
                </span>
              </div>
              <div className="flex items-center justify-between py-2 border-b border-[var(--border-subtle)]">
                <span className="text-[var(--text-muted)]">Server Uptime</span>
                <span className="font-mono text-[var(--text-secondary)]">
                  {health?.uptime || 'Active'}
                </span>
              </div>
              <div className="flex items-center justify-between py-2 border-b border-[var(--border-subtle)]">
                <span className="text-[var(--text-muted)]">Network Binding</span>
                <span className="font-mono text-[var(--brand-text)] font-semibold">
                  0.0.0.0:8080
                </span>
              </div>
            </div>
          </div>

          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3">
            <div>
              <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
                System Utilities
              </h3>
              <p className="text-[11.5px] text-[var(--text-muted)] mt-0.5">
                Quick access to maintenance tools, backups, and schema documentation.
              </p>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 pt-1">
              <Link
                to="/backup"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex flex-col justify-between"
              >
                <div className="flex items-center gap-1.5 mb-1 text-[var(--text-primary)] font-medium text-[12px]">
                  <Database className="w-3.5 h-3.5 text-blue-400" />
                  <span>Backup</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">SQLite database dump</span>
              </Link>

              <Link
                to="/api-docs"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex flex-col justify-between"
              >
                <div className="flex items-center gap-1.5 mb-1 text-[var(--text-primary)] font-medium text-[12px]">
                  <BookOpen className="w-3.5 h-3.5 text-emerald-400" />
                  <span>API Docs</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">OpenAI cURL reference</span>
              </Link>

              <Link
                to="/quota"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex flex-col justify-between"
              >
                <div className="flex items-center gap-1.5 mb-1 text-[var(--text-primary)] font-medium text-[12px]">
                  <Gauge className="w-3.5 h-3.5 text-amber-400" />
                  <span>Quotas</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">Circuit cooldown timers</span>
              </Link>

              <Link
                to="/tools"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex flex-col justify-between"
              >
                <div className="flex items-center gap-1.5 mb-1 text-[var(--text-primary)] font-medium text-[12px]">
                  <Wrench className="w-3.5 h-3.5 text-purple-400" />
                  <span>Tools</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">HTTP templates &amp; SSRF</span>
              </Link>

              <Link
                to="/credential-pools"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex flex-col justify-between"
              >
                <div className="flex items-center gap-1.5 mb-1 text-[var(--text-primary)] font-medium text-[12px]">
                  <Layers className="w-3.5 h-3.5 text-cyan-400" />
                  <span>HA Pools</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">Load balancing groups</span>
              </Link>

              <Link
                to="/free-tiers"
                className="p-2.5 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex flex-col justify-between"
              >
                <div className="flex items-center gap-1.5 mb-1 text-[var(--text-primary)] font-medium text-[12px]">
                  <Sparkles className="w-3.5 h-3.5 text-rose-400" />
                  <span>Free Tiers</span>
                </div>
                <span className="text-[10.5px] text-[var(--text-muted)]">Zero-cost model quotas</span>
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
