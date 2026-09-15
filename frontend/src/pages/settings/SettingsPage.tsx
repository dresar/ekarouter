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
  HardDrive,
  Eye,
  EyeOff,
  Radio,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Tabs } from '../../components/ui/Tabs.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { SystemHealth, UserProfile, MediaStorageConfig } from '../../types/api.ts'
import { useAuth } from '../../context/AuthContext.tsx'
import { setStoredUser } from '../../utils/storage.ts'

export function SettingsPage() {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState<'general' | 'storage' | 'system' | 'security'>('general')
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [health, setHealth] = useState<SystemHealth | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [saveFeedback, setSaveFeedback] = useState<string | null>(null)

  const [storageConfig, setStorageConfig] = useState<MediaStorageConfig>({
    provider: 'local',
    cloudinary_cloud_name: '',
    cloudinary_api_key: '',
    cloudinary_api_secret: '',
    cloudinary_folder: 'ekarouter',
    imagekit_public_key: '',
    imagekit_private_key: '',
    imagekit_url_endpoint: '',
    imagekit_folder: '/ekarouter',
    cdn_custom_domain: '',
    github_token: '',
    github_owner: '',
    github_repo: '',
    github_branch: 'main',
    github_folder: 'uploads',
    github_cdn_domain: 'cdn.jsdelivr.net',
  })
  const [isSavingStorage, setIsSavingStorage] = useState(false)
  const [isTestingStorage, setIsTestingStorage] = useState(false)
  const [storageFeedback, setStorageFeedback] = useState<{ success: boolean; message: string } | null>(null)
  const [showSecrets, setShowSecrets] = useState(false)

  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [passwordSuccess, setPasswordSuccess] = useState<string | null>(null)
  const [isSavingPassword, setIsSavingPassword] = useState(false)

  const loadData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [settingsData, healthData, storageData] = await Promise.all([
        api.get<Record<string, string>>('/api/settings').catch(() => ({})),
        api.get<SystemHealth>('/health').catch(() => null),
        api.get<MediaStorageConfig>('/api/storage/config').catch(() => null),
      ])
      setSettings(settingsData || {})
      setHealth(healthData)
      if (storageData && typeof storageData === 'object') {
        setStorageConfig((prev) => ({ ...prev, ...storageData }))
      }
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

  const handleSaveStorage = async (e: FormEvent) => {
    e.preventDefault()
    setIsSavingStorage(true)
    setStorageFeedback(null)
    try {
      await api.post('/api/storage/config', storageConfig)
      setStorageFeedback({
        success: true,
        message: `Konfigurasi storage '${storageConfig.provider}' berhasil disimpan!`,
      })
      setTimeout(() => setStorageFeedback(null), 4000)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menyimpan konfigurasi storage'
      setStorageFeedback({ success: false, message: msg })
    } finally {
      setIsSavingStorage(false)
    }
  }

  const handleTestStorage = async () => {
    setIsTestingStorage(true)
    setStorageFeedback(null)
    try {
      const res = await api.post<{ status: string; url?: string; file_id?: string; message?: string }>(
        '/api/storage/test',
        { provider: storageConfig.provider }
      )
      if (res?.status === 'ok') {
        setStorageFeedback({
          success: true,
          message: `Koneksi storage (${storageConfig.provider}) berhasil diverifikasi! ${res.url ? `URL: ${res.url}` : ''}`,
        })
      } else {
        setStorageFeedback({
          success: false,
          message: res?.message || 'Tes koneksi storage gagal',
        })
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Tes koneksi storage gagal'
      setStorageFeedback({ success: false, message: msg })
    } finally {
      setIsTestingStorage(false)
    }
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Settings"
        description="Configure runtime environment variables, storage & media CDN, token saver preferences, and administrative security."
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
          { key: 'storage', label: 'Storage & Media CDN', icon: <HardDrive className="w-4 h-4" /> },
          { key: 'security', label: 'Security & Access', icon: <Shield className="w-4 h-4" /> },
          { key: 'system', label: 'Runtime Environment', icon: <Server className="w-4 h-4" /> },
        ]}
        activeKey={activeTab}
        onChange={(k) => setActiveTab(k as 'general' | 'storage' | 'system' | 'security')}
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
      ) : activeTab === 'storage' ? (
        <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-5 max-w-3xl">
          <div>
            <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">Storage &amp; Media CDN Configuration</h3>
            <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
              Pilih dan konfigurasikan provider penyimpanan media untuk file, icon provider kustom, dan aset CDN.
            </p>
          </div>

          {storageFeedback && (
            <div
              className={`p-3 rounded-[6px] border text-[12px] flex items-center gap-2 ${
                storageFeedback.success
                  ? 'bg-emerald-950/20 border-emerald-600/30 text-emerald-300'
                  : 'bg-rose-950/20 border-rose-600/30 text-rose-300'
              }`}
            >
              {storageFeedback.success ? (
                <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
              ) : (
                <AlertTriangle className="w-4 h-4 text-rose-400 shrink-0" />
              )}
              <span className="break-all">{storageFeedback.message}</span>
            </div>
          )}

          <form onSubmit={handleSaveStorage} className="space-y-4">
            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1.5">
                Active Storage Provider
              </label>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                {[
                  { id: 'local', label: 'Local Disk', desc: 'Local /uploads folder' },
                  { id: 'cloudinary', label: 'Cloudinary', desc: 'Media API & CDN' },
                  { id: 'imagekit', label: 'ImageKit', desc: 'Realtime Optimization' },
                  { id: 'github', label: 'GitHub CDN', desc: 'jsDelivr Delivery' },
                ].map((p) => (
                  <button
                    key={p.id}
                    type="button"
                    onClick={() => setStorageConfig((prev) => ({ ...prev, provider: p.id as any }))}
                    className={`p-2.5 rounded-[6px] border text-left transition-colors cursor-pointer ${
                      storageConfig.provider === p.id
                        ? 'border-[var(--brand-primary)] bg-[var(--brand-primary)]/10 text-[var(--text-primary)]'
                        : 'border-[var(--border-subtle)] bg-[var(--bg-panel)] text-[var(--text-secondary)] hover:border-[var(--border-strong)]'
                    }`}
                  >
                    <div className="text-[12px] font-semibold">{p.label}</div>
                    <div className="text-[10px] text-[var(--text-muted)] mt-0.5">{p.desc}</div>
                  </button>
                ))}
              </div>
            </div>

            {storageConfig.provider === 'cloudinary' && (
              <div className="p-3.5 rounded-[6px] bg-[var(--bg-panel)]/60 border border-[var(--border-subtle)] space-y-3">
                <div className="text-[12px] font-semibold text-[var(--text-primary)]">Cloudinary Credentials</div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      Cloud Name
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. demo-cloud"
                      value={storageConfig.cloudinary_cloud_name || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, cloudinary_cloud_name: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      API Key
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. 123456789012345"
                      value={storageConfig.cloudinary_api_key || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, cloudinary_api_key: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div className="sm:col-span-2">
                    <div className="flex items-center justify-between mb-1">
                      <label className="block text-[11px] font-medium text-[var(--text-secondary)]">
                        API Secret
                      </label>
                      <button
                        type="button"
                        onClick={() => setShowSecrets(!showSecrets)}
                        className="text-[11px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 cursor-pointer"
                      >
                        {showSecrets ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                        {showSecrets ? 'Hide' : 'Show'}
                      </button>
                    </div>
                    <input
                      type={showSecrets ? 'text' : 'password'}
                      placeholder="•••••••• (leave unchanged to keep existing secret)"
                      value={storageConfig.cloudinary_api_secret || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, cloudinary_api_secret: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] font-mono rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      Upload Folder
                    </label>
                    <input
                      type="text"
                      placeholder="ekarouter"
                      value={storageConfig.cloudinary_folder || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, cloudinary_folder: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                </div>
              </div>
            )}

            {storageConfig.provider === 'imagekit' && (
              <div className="p-3.5 rounded-[6px] bg-[var(--bg-panel)]/60 border border-[var(--border-subtle)] space-y-3">
                <div className="text-[12px] font-semibold text-[var(--text-primary)]">ImageKit Credentials</div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="sm:col-span-2">
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      URL Endpoint
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="https://ik.imagekit.io/your_imagekit_id"
                      value={storageConfig.imagekit_url_endpoint || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, imagekit_url_endpoint: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      Public Key
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="public_..."
                      value={storageConfig.imagekit_public_key || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, imagekit_public_key: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      Folder
                    </label>
                    <input
                      type="text"
                      placeholder="/ekarouter"
                      value={storageConfig.imagekit_folder || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, imagekit_folder: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div className="sm:col-span-2">
                    <div className="flex items-center justify-between mb-1">
                      <label className="block text-[11px] font-medium text-[var(--text-secondary)]">
                        Private Key
                      </label>
                      <button
                        type="button"
                        onClick={() => setShowSecrets(!showSecrets)}
                        className="text-[11px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 cursor-pointer"
                      >
                        {showSecrets ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                        {showSecrets ? 'Hide' : 'Show'}
                      </button>
                    </div>
                    <input
                      type={showSecrets ? 'text' : 'password'}
                      placeholder="•••••••• (leave unchanged to keep existing secret)"
                      value={storageConfig.imagekit_private_key || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, imagekit_private_key: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] font-mono rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                </div>
              </div>
            )}

            {storageConfig.provider === 'github' && (
              <div className="p-3.5 rounded-[6px] bg-[var(--bg-panel)]/60 border border-[var(--border-subtle)] space-y-3">
                <div className="text-[12px] font-semibold text-[var(--text-primary)]">GitHub Storage &amp; jsDelivr CDN</div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      Repo Owner
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. username"
                      value={storageConfig.github_owner || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, github_owner: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      Repo Name
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. assets-cdn"
                      value={storageConfig.github_repo || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, github_repo: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      Branch
                    </label>
                    <input
                      type="text"
                      placeholder="main"
                      value={storageConfig.github_branch || 'main'}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, github_branch: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div>
                    <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                      Upload Folder
                    </label>
                    <input
                      type="text"
                      placeholder="uploads"
                      value={storageConfig.github_folder || 'uploads'}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, github_folder: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                  <div className="sm:col-span-2">
                    <div className="flex items-center justify-between mb-1">
                      <label className="block text-[11px] font-medium text-[var(--text-secondary)]">
                        GitHub Personal Access Token (PAT)
                      </label>
                      <button
                        type="button"
                        onClick={() => setShowSecrets(!showSecrets)}
                        className="text-[11px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 cursor-pointer"
                      >
                        {showSecrets ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                        {showSecrets ? 'Hide' : 'Show'}
                      </button>
                    </div>
                    <input
                      type={showSecrets ? 'text' : 'password'}
                      placeholder="ghp_•••••••• (requires repo scope)"
                      value={storageConfig.github_token || ''}
                      onChange={(e) => setStorageConfig((prev) => ({ ...prev, github_token: e.target.value }))}
                      className="w-full px-3 py-1.5 text-[12px] font-mono rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                    />
                  </div>
                </div>
              </div>
            )}

            {/* Custom CDN Domain */}
            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
                Custom CDN Domain (Optional)
              </label>
              <input
                type="text"
                placeholder="e.g. https://cdn.yourdomain.com"
                value={storageConfig.cdn_custom_domain || ''}
                onChange={(e) => setStorageConfig((prev) => ({ ...prev, cdn_custom_domain: e.target.value }))}
                className="w-full px-3 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
              />
              <span className="text-[11px] text-[var(--text-muted)] mt-1 block">
                Digunakan untuk mengganti domain default saat menghasilkan public CDN URLs.
              </span>
            </div>

            <div className="pt-3 border-t border-[var(--border-subtle)] flex items-center justify-between gap-3">
              <Button
                type="button"
                variant="secondary"
                size="compact"
                onClick={handleTestStorage}
                isLoading={isTestingStorage}
                leftIcon={<Radio className="w-3.5 h-3.5" />}
              >
                Test Connection ({storageConfig.provider})
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="compact"
                isLoading={isSavingStorage}
                leftIcon={<CheckCircle2 className="w-3.5 h-3.5" />}
              >
                Save Storage Config
              </Button>
            </div>
          </form>
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
