import { useEffect, useState } from 'react'
import { Database, HardDrive, RefreshCw, CheckCircle2, Play } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { BackupItem } from '../../types/api.ts'
import { formatDate, formatBytes } from '../../utils/formatters.ts'

export function BackupPage() {
  const [backups, setBackups] = useState<BackupItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isCreating, setIsCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [feedback, setFeedback] = useState<string | null>(null)

  const loadBackups = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.get<BackupItem[]>('/api/backup')
      setBackups(data || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load backup archives'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadBackups()
  }, [])

  const handleCreateBackup = async () => {
    setIsCreating(true)
    setFeedback(null)
    try {
      await api.post('/api/backup')
      setFeedback('Online SQLite VACUUM snapshot created successfully')
      setTimeout(() => setFeedback(null), 4000)
      await loadBackups()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Backup execution failed'
      setError(msg)
    } finally {
      setIsCreating(false)
    }
  }

  const columns: Column<BackupItem>[] = [
    {
      key: 'filename',
      title: 'Archive File Name',
      render: (item) => (
        <div className="flex items-center gap-2">
          <Database className="w-3.5 h-3.5 text-[var(--brand-text)]" />
          <span className="font-mono text-[12px] font-medium text-[var(--text-primary)]">
            {item.filename}
          </span>
        </div>
      ),
    },
    {
      key: 'size_bytes',
      title: 'File Size',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--text-secondary)]">
          {formatBytes(item.size_bytes || 0)}
        </span>
      ),
    },
    {
      key: 'created_at',
      title: 'Created Timestamp',
      render: (item) => (
        <span className="font-mono text-[11px] text-[var(--text-muted)]">
          {formatDate(item.created_at)}
        </span>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Database Backup"
        description="Manage online SQLite VACUUM INTO database archives without taking the gateway offline."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Database Backup' },
        ]}
        metadata={
          <span>
            {backups.length} snapshots available
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={loadBackups}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Button
              variant="primary"
              size="compact"
              onClick={handleCreateBackup}
              isLoading={isCreating}
              leftIcon={<Play className="w-3.5 h-3.5" />}
            >
              Create Snapshot
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadBackups} />}

      {feedback && (
        <div className="p-3 rounded-[6px] bg-emerald-950/20 border border-emerald-600/30 text-[12px] text-emerald-300 flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          <span>{feedback}</span>
        </div>
      )}

      <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3.5 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-[6px] bg-[var(--bg-panel)] flex items-center justify-center text-[var(--brand-text)]">
            <HardDrive className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
              Embedded SQLite WAL Engine
            </h3>
            <p className="text-[11px] text-[var(--text-muted)]">
              Hot backups run concurrently with live gateway traffic without lock contention
            </p>
          </div>
        </div>
        <span className="font-mono text-[11px] text-[var(--text-muted)] bg-[var(--bg-panel)] px-2 py-0.5 rounded border border-[var(--border-subtle)]">
          data/ekarouter.db
        </span>
      </div>

      <DataTable
        columns={columns}
        data={backups}
        isLoading={isLoading}
        keyExtractor={(b) => b.filename}
        emptyMessage="No backup archives recorded. Click 'Create Snapshot' to execute a snapshot."
      />
    </div>
  )
}
