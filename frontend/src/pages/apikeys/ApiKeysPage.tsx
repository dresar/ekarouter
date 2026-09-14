import { useEffect, useState, FormEvent } from 'react'
import { Key, Plus, Trash2, Copy, Check, AlertTriangle, ShieldCheck, RefreshCw } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { ApiKey } from '../../types/api.ts'
import { formatDate } from '../../utils/formatters.ts'

interface CreatedKeyResult {
  key: string
  name: string
}

export function ApiKeysPage() {
  const [keys, setKeys] = useState<ApiKey[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showGenerateForm, setShowGenerateForm] = useState(false)
  const [keyName, setKeyName] = useState('')
  const [keyScopes, setKeyScopes] = useState('*')
  const [isGenerating, setIsGenerating] = useState(false)
  const [generateError, setGenerateError] = useState<string | null>(null)

  const [oneTimeSecret, setOneTimeSecret] = useState<CreatedKeyResult | null>(null)
  const [copied, setCopied] = useState(false)

  const fetchKeys = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.get<ApiKey[]>('/api/keys')
      setKeys(data || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load API keys'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchKeys()
  }, [])

  const handleGenerateKey = async (e: FormEvent) => {
    e.preventDefault()
    if (!keyName.trim()) return
    setIsGenerating(true)
    setGenerateError(null)

    try {
      const res = await api.post<{ key: string; name: string }>('/api/keys', {
        name: keyName.trim(),
        scopes: keyScopes.trim() || '*',
      })
      setOneTimeSecret({
        key: res.key,
        name: keyName.trim(),
      })
      setKeyName('')
      setKeyScopes('*')
      setShowGenerateForm(false)
      await fetchKeys()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to generate key'
      setGenerateError(msg)
    } finally {
      setIsGenerating(false)
    }
  }

  const handleCopyOneTime = async () => {
    if (!oneTimeSecret?.key) return
    try {
      await navigator.clipboard.writeText(oneTimeSecret.key)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {}
  }

  const handleRevoke = async (id: string) => {
    try {
      await api.delete(`/api/keys/${id}`)
      setKeys((prev) => prev.filter((k) => k.id !== id))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to revoke key'
      setError(msg)
    }
  }

  const columns: Column<ApiKey>[] = [
    {
      key: 'name',
      title: 'Key Name',
      render: (item) => (
        <div className="flex items-center gap-2">
          <Key className="w-3.5 h-3.5 text-[var(--brand-text)]" />
          <span className="font-semibold text-[13px] text-[var(--text-primary)]">{item.name}</span>
        </div>
      ),
    },
    {
      key: 'prefix',
      title: 'Token Prefix',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--text-secondary)] bg-[var(--bg-panel)] px-2 py-0.5 rounded border border-[var(--border-subtle)]">
          {item.prefix || 'eka_live_'}...
        </span>
      ),
    },
    {
      key: 'scopes',
      title: 'Scopes',
      render: (item) => (
        <span className="font-mono text-[10.5px] text-[var(--text-muted)] bg-[var(--bg-panel)] px-2 py-0.5 rounded">
          {item.scopes || '*'}
        </span>
      ),
    },
    {
      key: 'created_at',
      title: 'Created At',
      render: (item) => (
        <span className="text-[11px] font-mono text-[var(--text-muted)]">
          {formatDate(item.created_at)}
        </span>
      ),
    },
    {
      key: 'last_used_at',
      title: 'Last Used',
      render: (item) => (
        <span className="text-[11px] font-mono text-[var(--text-muted)]">
          {formatDate(item.last_used_at)}
        </span>
      ),
    },
    {
      key: 'actions',
      title: 'Actions',
      width: '80px',
      render: (item) => (
        <InlineConfirm
          trigger={
            <button
              type="button"
              className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] hover:bg-[var(--status-danger)]/10 transition-colors"
              title="Revoke API Key"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          }
          confirmText="Revoke?"
          onConfirm={() => handleRevoke(item.id)}
        />
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Ingress API Keys"
        description="Manage developer client bearer tokens used to authenticate external AI tools against /v1/* endpoints."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Ingress API Keys' },
        ]}
        metadata={
          <span>
            {keys.length} active developer keys
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={fetchKeys}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Button
              variant="primary"
              size="compact"
              onClick={() => setShowGenerateForm(!showGenerateForm)}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
            >
              {showGenerateForm ? 'Close' : 'Generate Key'}
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchKeys} />}

      {oneTimeSecret && (
        <div className="bg-[var(--brand-primary)]/10 border-2 border-[var(--brand-primary)] rounded-[8px] p-4 space-y-2.5 animate-in fade-in duration-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-[var(--brand-text)] font-semibold text-[13px]">
              <ShieldCheck className="w-4 h-4" />
              <span>Save Your New API Key</span>
            </div>
            <span className="text-[11px] text-amber-400 font-mono font-medium">
              Will not be shown again
            </span>
          </div>

          <p className="text-[12px] text-[var(--text-secondary)]">
            Key <span className="font-semibold text-[var(--text-primary)]">{oneTimeSecret.name}</span> has been generated. Copy it now to configure Cursor, Claude Code, or Aider:
          </p>

          <div className="flex items-center gap-2">
            <input
              type="text"
              readOnly
              value={oneTimeSecret.key}
              className="flex-1 px-3 py-2 bg-[var(--bg-panel)] text-[12px] font-mono text-[var(--brand-text)] rounded-[5px] border border-[var(--border-subtle)] select-all focus:outline-none"
            />
            <Button
              variant="primary"
              size="compact"
              onClick={handleCopyOneTime}
              leftIcon={copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
            >
              {copied ? 'Copied' : 'Copy'}
            </Button>
            <Button
              variant="secondary"
              size="compact"
              onClick={() => setOneTimeSecret(null)}
            >
              Saved
            </Button>
          </div>
        </div>
      )}

      {showGenerateForm && (
        <form
          onSubmit={handleGenerateKey}
          className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3 animate-in fade-in duration-150"
        >
          <div className="flex items-center justify-between border-b border-[var(--border-subtle)] pb-2">
            <span className="text-[13px] font-semibold text-[var(--text-primary)]">
              Generate Ingress API Key
            </span>
          </div>

          {generateError && (
            <div className="p-2.5 rounded bg-[var(--status-danger)]/10 text-[var(--status-danger)] text-[12px] flex items-center gap-1.5">
              <AlertTriangle className="w-3.5 h-3.5" />
              <span>{generateError}</span>
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Client / Key Name *
              </label>
              <input
                type="text"
                required
                value={keyName}
                onChange={(e) => setKeyName(e.target.value)}
                placeholder="e.g. Cursor IDE Workstation"
                className="w-full px-2.5 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
              />
            </div>

            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Scopes
              </label>
              <input
                type="text"
                value={keyScopes}
                onChange={(e) => setKeyScopes(e.target.value)}
                placeholder="*"
                className="w-full px-2.5 py-1.5 text-[12px] font-mono rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-[var(--border-subtle)]">
            <Button
              type="button"
              variant="ghost"
              size="compact"
              onClick={() => setShowGenerateForm(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isGenerating}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
            >
              Generate Key
            </Button>
          </div>
        </form>
      )}

      <DataTable
        columns={columns}
        data={keys}
        isLoading={isLoading}
        keyExtractor={(k) => k.id}
        emptyMessage="No client API keys generated yet. Click 'Generate Key' above."
      />
    </div>
  )
}
