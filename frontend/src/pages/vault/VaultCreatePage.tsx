import { useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Save, AlertCircle } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

export function VaultCreatePage() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState({
    name: '',
    provider_id: 'openai',
    credential_type: 'api_key',
    secret_value: '',
    environment: 'production',
    priority: 1,
    tags: 'ai,gateway',
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    setError(null)

    try {
      await api.post('/api/v1/credentials', {
        name: formData.name.trim(),
        provider_id: formData.provider_id.trim(),
        credential_type: formData.credential_type,
        secret_value: formData.secret_value.trim(),
        environment: formData.environment,
        priority: Number(formData.priority),
        tags: formData.tags.trim(),
      })
      navigate('/vault')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to store secret in vault'
      setError(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-4">
      <PageHeader
        title="Tambah Kredensial"
        description="Enkripsi dan simpan kredensial ke dalam vault AES-256-GCM."
        breadcrumbs={[
          { label: 'Vault', to: '/vault' },
          { label: 'New Credential' },
        ]}
        actions={
          <Link to="/vault">
            <Button variant="ghost" size="compact" leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}>
              Kembali
            </Button>
          </Link>
        }
      />

      {error && (
        <div className="p-3 rounded-[6px] bg-[var(--status-danger)]/10 border border-[var(--status-danger)]/30 flex items-center gap-2 text-[12.5px] text-[var(--status-danger)]">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      <form
        onSubmit={handleSubmit}
        className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-5 space-y-4"
      >
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Credential Name *
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Nama"
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Provider ID / Slug *
            </label>
            <input
              type="text"
              required
              value={formData.provider_id}
              onChange={(e) => setFormData({ ...formData, provider_id: e.target.value })}
              placeholder="Provider"
              className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
            />
          </div>
        </div>

        <div>
          <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
            Secret API Key / Token *
          </label>
          <input
            type="password"
            required
            autoComplete="new-password"
            value={formData.secret_value}
            onChange={(e) => setFormData({ ...formData, secret_value: e.target.value })}
            placeholder="Secret"
            className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
          />
          <span className="text-[11px] text-[var(--text-muted)] mt-1 block">
            Plaintext is never returned after creation.
          </span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Credential Type *
            </label>
            <select
              value={formData.credential_type}
              onChange={(e) => setFormData({ ...formData, credential_type: e.target.value })}
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            >
              <option value="api_key">API Key</option>
              <option value="bearer_token">Bearer Token</option>
              <option value="basic_auth">Basic Auth</option>
              <option value="oauth2">OAuth 2.0</option>
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Environment *
            </label>
            <select
              value={formData.environment}
              onChange={(e) => setFormData({ ...formData, environment: e.target.value })}
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            >
              <option value="production">Production</option>
              <option value="staging">Staging</option>
              <option value="development">Development</option>
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Priority (1 = Highest)
            </label>
            <input
              type="number"
              min="1"
              max="10"
              value={formData.priority}
              onChange={(e) =>
                setFormData({ ...formData, priority: parseInt(e.target.value) || 1 })
              }
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
          </div>
        </div>

        <div>
          <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
            Tags (Comma-separated)
          </label>
          <input
            type="text"
            value={formData.tags}
            onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
            placeholder="Tag"
            className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
          />
        </div>

        <div className="flex items-center justify-end gap-2 pt-3 border-t border-[var(--border-subtle)]">
          <Link to="/vault">
            <Button type="button" variant="ghost" size="compact">
              Batal
            </Button>
          </Link>
          <Button
            type="submit"
            variant="primary"
            size="compact"
            isLoading={isSubmitting}
            leftIcon={<Save className="w-3.5 h-3.5" />}
          >
            Simpan
          </Button>
        </div>
      </form>
    </div>
  )
}
