import { useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Save, AlertCircle } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

export function ToolCreatePage() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState({
    name: '',
    provider_id: 'resend',
    category: 'communication',
    method: 'POST',
    url_template: 'https://api.resend.com/emails',
    timeout_ms: 10000,
    retry_count: 2,
    description: 'Dispatch transactional emails via Resend API',
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    setError(null)

    try {
      await api.post('/api/v1/tools', {
        name: formData.name.trim(),
        provider_id: formData.provider_id.trim(),
        category: formData.category,
        method: formData.method,
        url_template: formData.url_template.trim(),
        timeout_ms: Number(formData.timeout_ms) || 10000,
        retry_count: Number(formData.retry_count) || 2,
        description: formData.description.trim(),
      })
      navigate('/tools')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create tool'
      setError(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-4">
      <PageHeader
        title="Define Generic HTTP Tool"
        description="Configure parameterized REST API actions with SSRF protection."
        breadcrumbs={[
          { label: 'Tools', to: '/tools' },
          { label: 'New Tool' },
        ]}
        actions={
          <Link to="/tools">
            <Button variant="ghost" size="compact" leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}>
              Back to Tools
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
              Tool Name *
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g. send_email"
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Provider ID / Vault Scope *
            </label>
            <input
              type="text"
              required
              value={formData.provider_id}
              onChange={(e) => setFormData({ ...formData, provider_id: e.target.value })}
              placeholder="e.g. resend or stripe"
              className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
            />
          </div>
        </div>

        <div>
          <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
            Description
          </label>
          <input
            type="text"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="Functional summary for tool calling"
            className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
          />
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Category *
            </label>
            <select
              value={formData.category}
              onChange={(e) => setFormData({ ...formData, category: e.target.value })}
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            >
              <option value="developer">Developer</option>
              <option value="communication">Communication</option>
              <option value="monitoring">Monitoring</option>
              <option value="storage">Storage</option>
              <option value="payments">Payments</option>
              <option value="custom">Custom</option>
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              HTTP Method *
            </label>
            <select
              value={formData.method}
              onChange={(e) => setFormData({ ...formData, method: e.target.value })}
              className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
            >
              <option value="POST">POST</option>
              <option value="GET">GET</option>
              <option value="PUT">PUT</option>
              <option value="PATCH">PATCH</option>
              <option value="DELETE">DELETE</option>
            </select>
          </div>
        </div>

        <div>
          <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
            Target URL Template *
          </label>
          <input
            type="url"
            required
            value={formData.url_template}
            onChange={(e) => setFormData({ ...formData, url_template: e.target.value })}
            placeholder="https://api.service.com/v1/resource/{{id}}"
            className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
          />
          <span className="text-[11px] text-[var(--text-muted)] mt-1 block">
            Supports mustache variables like {'{{param}}'} with automated SSRF loopback blocking
          </span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Timeout (ms)
            </label>
            <input
              type="number"
              value={formData.timeout_ms}
              onChange={(e) =>
                setFormData({ ...formData, timeout_ms: parseInt(e.target.value) || 10000 })
              }
              className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Retry Attempts
            </label>
            <input
              type="number"
              min="0"
              max="5"
              value={formData.retry_count}
              onChange={(e) =>
                setFormData({ ...formData, retry_count: parseInt(e.target.value) || 0 })
              }
              className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
            />
          </div>
        </div>

        <div className="flex items-center justify-end gap-2 pt-3 border-t border-[var(--border-subtle)]">
          <Link to="/tools">
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
            Save Tool Definition
          </Button>
        </div>
      </form>
    </div>
  )
}
