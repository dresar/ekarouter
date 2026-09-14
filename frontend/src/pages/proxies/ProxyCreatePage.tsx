import { useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Save, AlertCircle } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

export function ProxyCreatePage() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState({
    name: '',
    scheme: 'http',
    host: '',
    port: 8080,
    username: '',
    password: '',
    enabled: true,
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    setError(null)

    try {
      await api.post('/api/proxies', {
        name: formData.name.trim(),
        scheme: formData.scheme,
        host: formData.host.trim(),
        port: Number(formData.port),
        username: formData.username.trim() || undefined,
        password: formData.password || undefined,
        enabled: formData.enabled,
      })
      navigate('/proxies')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to register proxy profile'
      setError(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-4">
      <PageHeader
        title="Add Outbound Proxy Profile"
        description="Configure an egress proxy for geographic routing or network IP isolation."
        breadcrumbs={[
          { label: 'Proxies', to: '/proxies' },
          { label: 'New Proxy' },
        ]}
        actions={
          <Link to="/proxies">
            <Button variant="ghost" size="compact" leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}>
              Back to Proxies
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
              Profile Name *
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g. US Residential Egress"
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Proxy Protocol *
            </label>
            <select
              value={formData.scheme}
              onChange={(e) => setFormData({ ...formData, scheme: e.target.value })}
              className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
            >
              <option value="http">HTTP</option>
              <option value="https">HTTPS</option>
              <option value="socks5">SOCKS5</option>
              <option value="relay">Relay</option>
            </select>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="sm:col-span-2">
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Host / IP Address *
            </label>
            <input
              type="text"
              required
              value={formData.host}
              onChange={(e) => setFormData({ ...formData, host: e.target.value })}
              placeholder="e.g. proxy.gateway.net"
              className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Port *
            </label>
            <input
              type="number"
              min="1"
              max="65535"
              required
              value={formData.port}
              onChange={(e) =>
                setFormData({ ...formData, port: parseInt(e.target.value) || 8080 })
              }
              className="w-full px-3 py-1.5 text-[13px] font-mono rounded-[5px] focus:outline-none"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Username (Optional)
            </label>
            <input
              type="text"
              value={formData.username}
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              placeholder="Auth username"
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1">
              Password (Optional)
            </label>
            <input
              type="password"
              autoComplete="new-password"
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              placeholder="Auth password"
              className="w-full px-3 py-1.5 text-[13px] rounded-[5px] focus:outline-none"
            />
          </div>
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
              Enable proxy profile immediately
            </span>
          </label>

          <div className="flex items-center gap-2">
            <Link to="/proxies">
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
              Save Proxy
            </Button>
          </div>
        </div>
      </form>
    </div>
  )
}
