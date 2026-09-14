import { useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Save, AlertCircle } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

export function ProviderCreatePage() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState({
    id: '',
    key: '',
    name: '',
    kind: 'openai',
    base_url: 'https://api.openai.com/v1',
    enabled: true,
  })
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const kinds = [
    { value: 'openai', label: 'OpenAI Compatible', defaultUrl: 'https://api.openai.com/v1' },
    { value: 'anthropic', label: 'Anthropic Claude', defaultUrl: 'https://api.anthropic.com/v1' },
    { value: 'gemini', label: 'Google Gemini', defaultUrl: 'https://generativelanguage.googleapis.com/v1beta' },
    { value: 'groq', label: 'Groq Cloud', defaultUrl: 'https://api.groq.com/openai/v1' },
    { value: 'cerebras', label: 'Cerebras AI', defaultUrl: 'https://api.cerebras.ai/v1' },
    { value: 'openrouter', label: 'OpenRouter', defaultUrl: 'https://openrouter.ai/api/v1' },
    { value: 'cloudflare', label: 'Cloudflare AI', defaultUrl: 'https://api.cloudflare.com/client/v4/accounts' },
    { value: 'ollama', label: 'Ollama Local', defaultUrl: 'http://localhost:11434' },
    { value: 'custom', label: 'Custom Upstream API', defaultUrl: 'https://api.custom.com/v1' },
  ]

  const handleKindChange = (newKind: string) => {
    const selected = kinds.find((k) => k.value === newKind)
    setFormData((prev) => ({
      ...prev,
      kind: newKind,
      base_url: selected ? selected.defaultUrl : prev.base_url,
    }))
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)

    const payload = {
      ...formData,
      id: formData.id.trim() || `prov_${formData.key.toLowerCase().replace(/[^a-z0-9]/g, '_')}`,
    }

    try {
      await api.post('/api/providers', payload)
      navigate('/providers')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to register provider'
      setError(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-4">
      <PageHeader
        title="Register AI Provider"
        description="Configure a new upstream AI model backend and connection parameters."
        breadcrumbs={[
          { label: 'Providers', to: '/providers' },
          { label: 'New Provider' },
        ]}
        actions={
          <Link to="/providers">
            <Button variant="ghost" size="compact" leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}>
              Back to List
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
              Provider Key (Slug) *
            </label>
            <input
              type="text"
              required
              value={formData.key}
              onChange={(e) => setFormData({ ...formData, key: e.target.value })}
              placeholder="e.g. openai-prod"
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
            <span className="text-[11px] text-[var(--text-muted)] mt-1 block">
              Unique identifier used in routes
            </span>
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Display Name *
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g. OpenAI Commercial"
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Backend Adapter Kind *
            </label>
            <select
              value={formData.kind}
              onChange={(e) => handleKindChange(e.target.value)}
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            >
              {kinds.map((k) => (
                <option key={k.value} value={k.value}>
                  {k.label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Custom Provider ID (Optional)
            </label>
            <input
              type="text"
              value={formData.id}
              onChange={(e) => setFormData({ ...formData, id: e.target.value })}
              placeholder="Auto-generated if empty"
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
          </div>
        </div>

        <div>
          <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
            Upstream Base URL *
          </label>
          <input
            type="url"
            required
            value={formData.base_url}
            onChange={(e) => setFormData({ ...formData, base_url: e.target.value })}
            placeholder="https://api.openai.com/v1"
            className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
          />
        </div>

        <div className="pt-2 border-t border-[var(--border-subtle)] flex items-center justify-between">
          <label className="flex items-center gap-2 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={formData.enabled}
              onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
              className="rounded w-4 h-4 text-[var(--brand-primary)]"
            />
            <span className="text-[13px] font-medium text-[var(--text-primary)]">
              Enable provider immediately
            </span>
          </label>

          <div className="flex items-center gap-2">
            <Link to="/providers">
              <Button type="button" variant="ghost" size="compact">
                Cancel
              </Button>
            </Link>
            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isSubmitting}
              leftIcon={<Save className="w-3.5 h-3.5" />}
            >
              Save Provider
            </Button>
          </div>
        </div>
      </form>
    </div>
  )
}
