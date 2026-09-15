import { useEffect, useState } from 'react'
import { RefreshCw, Eye, User, Search } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { RightDrawer } from '../../components/ui/RightDrawer.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { AuditRecord } from '../../types/api.ts'
import { formatDate } from '../../utils/formatters.ts'

export function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditRecord[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [selectedRecord, setSelectedRecord] = useState<AuditRecord | null>(null)
  const [search, setSearch] = useState('')

  const fetchLogs = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.get<AuditRecord[]>('/api/v1/audit-logs')
      setLogs(data || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load audit logs'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchLogs()
  }, [])

  useEffect(() => {
    if (!autoRefresh) return
    const timer = setInterval(fetchLogs, 10000)
    return () => clearInterval(timer)
  }, [autoRefresh])

  const filteredLogs = logs.filter(
    (l) =>
      l.action.toLowerCase().includes(search.toLowerCase()) ||
      l.actor_id.toLowerCase().includes(search.toLowerCase()) ||
      (l.resource_type && l.resource_type.toLowerCase().includes(search.toLowerCase()))
  )

  const columns: Column<AuditRecord>[] = [
    {
      key: 'timestamp',
      title: 'Timestamp',
      render: (item) => (
        <span className="font-mono text-[11px] text-[var(--text-muted)]">
          {formatDate(item.timestamp)}
        </span>
      ),
    },
    {
      key: 'actor_id',
      title: 'Actor',
      render: (item) => (
        <div className="flex items-center gap-1.5">
          <User className="w-3.5 h-3.5 text-[var(--text-muted)]" />
          <span className="font-mono text-[12px] text-[var(--text-primary)] font-medium">
            {item.actor_id}
          </span>
        </div>
      ),
    },
    {
      key: 'action',
      title: 'Action',
      render: (item) => (
        <span className="font-mono text-[12px] font-semibold text-[var(--brand-text)]">
          {item.action}
        </span>
      ),
    },
    {
      key: 'resource_type',
      title: 'Resource',
      render: (item) => (
        <div className="flex items-center gap-1.5 text-[11px] font-mono">
          <span className="text-[var(--text-secondary)]">{item.resource_type}</span>
          {item.resource_id && (
            <span className="text-[var(--text-muted)] truncate max-w-[120px]">
              ({item.resource_id})
            </span>
          )}
        </div>
      ),
    },
    {
      key: 'result',
      title: 'Result',
      render: (item) => (
        <StatusBadge
          variant={item.result === 'success' || item.result === 'ok' ? 'healthy' : 'error'}
        >
          {item.result}
        </StatusBadge>
      ),
    },
    {
      key: 'actions',
      title: 'Details',
      width: '70px',
      render: (item) => (
        <button
          type="button"
          onClick={() => setSelectedRecord(item)}
          title="Inspect log JSON metadata"
          className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
        >
          <Eye className="w-4 h-4" />
        </button>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Audit Logs"
        description="Immutable compliance audit records of administrative changes, credential updates, and gateway events."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Audit Logs' },
        ]}
        metadata={
          <span>
            {logs.length} audit records captured
          </span>
        }
        actions={
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 cursor-pointer select-none text-[12px] text-[var(--text-secondary)]">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="rounded w-3.5 h-3.5 text-[var(--brand-primary)] accent-blue-600"
              />
              <span>Auto-refresh (10s)</span>
            </label>

            <Button
              variant="secondary"
              size="compact"
              onClick={fetchLogs}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={fetchLogs} />}

      <div className="flex items-center gap-2.5 bg-[var(--bg-surface)] p-2.5 rounded-[8px] border border-[var(--border-subtle)]">
        <Search className="w-3.5 h-3.5 text-[var(--text-muted)] ml-1 pointer-events-none" />
        <input
          type="text"
          placeholder="Cari"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full px-2 py-1 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)]"
        />
      </div>

      <DataTable
        columns={columns}
        data={filteredLogs}
        isLoading={isLoading}
        keyExtractor={(r) => r.id}
        emptyMessage="No audit logs recorded yet."
      />

      <RightDrawer
        isOpen={!!selectedRecord}
        onClose={() => setSelectedRecord(null)}
        title="Audit Record Details"
        description={`Record ID: ${selectedRecord?.id || ''}`}
        footer={
          <Button variant="secondary" size="compact" onClick={() => setSelectedRecord(null)}>
            Close
          </Button>
        }
      >
        {selectedRecord && (
          <div className="space-y-4 text-[12.5px]">
            <div className="space-y-2 border-b border-[var(--border-subtle)] pb-3">
              <div>
                <span className="text-[10.5px] font-semibold uppercase text-[var(--text-muted)] block">
                  Action
                </span>
                <span className="font-mono text-[13px] font-bold text-[var(--brand-text)]">
                  {selectedRecord.action}
                </span>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <span className="text-[10.5px] font-semibold uppercase text-[var(--text-muted)] block">
                    Actor
                  </span>
                  <span className="font-mono text-[var(--text-primary)]">
                    {selectedRecord.actor_id}
                  </span>
                </div>
                <div>
                  <span className="text-[10.5px] font-semibold uppercase text-[var(--text-muted)] block">
                    Result
                  </span>
                  <span className="font-mono text-[var(--text-primary)]">
                    {selectedRecord.result}
                  </span>
                </div>
              </div>
            </div>

            <div>
              <span className="text-[10.5px] font-semibold uppercase text-[var(--text-muted)] block mb-1.5">
                Raw JSON Metadata
              </span>
              <pre className="bg-[#07090E] p-3 rounded-[6px] border border-[var(--border-subtle)] text-[11px] font-mono text-[var(--text-secondary)] overflow-x-auto leading-relaxed select-all">
                {JSON.stringify(selectedRecord.details || selectedRecord, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </RightDrawer>
    </div>
  )
}
