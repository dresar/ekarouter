import { useEffect, useState, FormEvent } from 'react'
import { Plus, Trash2, RefreshCw, AlertCircle, Save } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { ModelCatalogEntry, Provider } from '../../types/api.ts'
import { formatCompactNumber } from '../../utils/formatters.ts'

export function ModelsPage() {
  const [models, setModels] = useState<ModelCatalogEntry[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showAddForm, setShowAddForm] = useState(false)
  const [newModel, setNewModel] = useState({
    provider_id: '',
    external_name: '',
    display_name: '',
    context_limit: 128000,
    streaming: true,
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const loadData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [modelsData, provData] = await Promise.all([
        api.get<ModelCatalogEntry[]>('/api/models').catch(() => []),
        api.get<Provider[]>('/api/providers').catch(() => []),
      ])
      setModels(modelsData || [])
      setProviders(provData || [])
      if (provData && provData.length > 0 && !newModel.provider_id) {
        setNewModel((prev) => ({ ...prev, provider_id: provData[0].id }))
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load model catalog'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleCreateModel = async (e: FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    setFormError(null)

    try {
      await api.post('/api/models', {
        provider_id: newModel.provider_id,
        external_name: newModel.external_name.trim(),
        display_name: newModel.display_name.trim(),
        context_limit: Number(newModel.context_limit) || 128000,
        streaming: newModel.streaming,
      })
      setNewModel({
        provider_id: providers[0]?.id || '',
        external_name: '',
        display_name: '',
        context_limit: 128000,
        streaming: true,
      })
      setShowAddForm(false)
      await loadData()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to register model'
      setFormError(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/api/models/${id}`)
      setModels((prev) => prev.filter((m) => m.id !== id))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete model'
      setError(msg)
    }
  }

  const columns: Column<ModelCatalogEntry>[] = [
    {
      key: 'display_name',
      title: 'Model Alias',
      render: (item) => (
        <div className="flex items-center gap-2.5">
          <ProviderLogo providerId={item.provider_id} size="sm" />
          <div className="flex flex-col">
            <span className="font-semibold text-[13px] text-[var(--text-primary)]">{item.display_name}</span>
            <span className="text-[10.5px] font-mono text-[var(--text-muted)]">{item.provider_id}</span>
          </div>
        </div>
      ),
    },
    {
      key: 'external_name',
      title: 'Upstream Identifier',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--text-secondary)]">
          {item.external_name}
        </span>
      ),
    },
    {
      key: 'context_limit',
      title: 'Context Window',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--text-secondary)]">
          {item.context_limit ? `${formatCompactNumber(item.context_limit)} tokens` : '128K tokens'}
        </span>
      ),
    },
    {
      key: 'streaming',
      title: 'SSE Streaming',
      render: (item) => (
        <StatusBadge variant={item.streaming !== false ? 'healthy' : 'disabled'}>
          {item.streaming !== false ? 'Supported' : 'No'}
        </StatusBadge>
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
            >
              <Trash2 className="w-4 h-4" />
            </button>
          }
          confirmText="Delete?"
          onConfirm={() => handleDelete(item.id)}
        />
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Model Catalog"
        description="Catalog of upstream models, context limits, streaming flags, and routing aliases."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Model Catalog' },
        ]}
        metadata={
          <span>
            {models.length} model definitions
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={loadData}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Button
              variant="primary"
              size="compact"
              onClick={() => setShowAddForm(!showAddForm)}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
            >
              {showAddForm ? 'Close Form' : 'Register Model'}
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadData} />}

      {showAddForm && (
        <form
          onSubmit={handleCreateModel}
          className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3.5 animate-in fade-in duration-150"
        >
          <div className="flex items-center justify-between border-b border-[var(--border-subtle)] pb-2">
            <span className="text-[13px] font-semibold text-[var(--text-primary)]">
              Register Model Alias
            </span>
          </div>

          {formError && (
            <div className="p-2.5 rounded bg-[var(--status-danger)]/10 text-[var(--status-danger)] text-[12px] flex items-center gap-1.5">
              <AlertCircle className="w-3.5 h-3.5" />
              <span>{formError}</span>
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Provider Backend *
              </label>
              <select
                value={newModel.provider_id}
                onChange={(e) => setNewModel({ ...newModel, provider_id: e.target.value })}
                className="w-full px-2.5 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
              >
                {providers.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name} ({p.kind})
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Display Name / Alias *
              </label>
              <input
                type="text"
                required
                value={newModel.display_name}
                onChange={(e) => setNewModel({ ...newModel, display_name: e.target.value })}
                placeholder="e.g. GPT-4o Production"
                className="w-full px-2.5 py-1.5 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
              />
            </div>

            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Upstream Model Identifier *
              </label>
              <input
                type="text"
                required
                value={newModel.external_name}
                onChange={(e) => setNewModel({ ...newModel, external_name: e.target.value })}
                placeholder="e.g. gpt-4o-2024-08-06"
                className="w-full px-2.5 py-1.5 text-[12px] font-mono rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                Context Window (Tokens)
              </label>
              <input
                type="number"
                value={newModel.context_limit}
                onChange={(e) =>
                  setNewModel({ ...newModel, context_limit: parseInt(e.target.value) || 128000 })
                }
                className="w-full px-2.5 py-1.5 text-[12px] font-mono rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none"
              />
            </div>

            <div className="flex items-center pt-5">
              <label className="flex items-center gap-2 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={newModel.streaming}
                  onChange={(e) => setNewModel({ ...newModel, streaming: e.target.checked })}
                  className="rounded w-3.5 h-3.5 text-[var(--brand-primary)]"
                />
                <span className="text-[12px] text-[var(--text-primary)]">
                  Supports SSE streaming completions
                </span>
              </label>
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-[var(--border-subtle)]">
            <Button
              type="button"
              variant="ghost"
              size="compact"
              onClick={() => setShowAddForm(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isSubmitting}
              leftIcon={<Save className="w-3.5 h-3.5" />}
            >
              Save Model
            </Button>
          </div>
        </form>
      )}

      <DataTable
        columns={columns}
        data={models}
        isLoading={isLoading}
        keyExtractor={(m) => m.id}
        emptyMessage="No custom model aliases registered. Use the register button above to configure specific models."
      />
    </div>
  )
}
