import { useEffect, useState, useRef } from 'react'
import {
  Play,
  Pause,
  Trash2,
  Download,
  Filter,
  Search,
  Send,
  Maximize2,
  Minimize2,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'
import { AuditRecord } from '../../types/api.ts'

interface LogEntry {
  id: string
  timestamp: string
  level: 'INFO' | 'WARN' | 'ERROR' | 'DEBUG' | 'HTTP'
  subsystem: string
  action: string
  message: string
  latencyMs?: number
  statusCode?: number
  details?: Record<string, unknown>
}

export function LiveConsolePage() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [filterLevel, setFilterLevel] = useState<string>('ALL')
  const [filterSubsystem, setFilterSubsystem] = useState<string>('ALL')
  const [searchQuery, setSearchQuery] = useState('')
  const [isStreaming, setIsStreaming] = useState(true)
  const [autoScroll, setAutoScroll] = useState(true)
  const [selectedEntry, setSelectedEntry] = useState<LogEntry | null>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [isPinging, setIsPinging] = useState(false)

  const consoleEndRef = useRef<HTMLDivElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const formatAuditLog = (record: AuditRecord): LogEntry => {
    const isErr = record.result === 'failure' || record.result === 'error'
    const timeStr = record.timestamp
      ? new Date(record.timestamp).toLocaleTimeString('en-US', {
          hour12: false,
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          fractionalSecondDigits: 3,
        })
      : new Date().toLocaleTimeString('en-US', {
          hour12: false,
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          fractionalSecondDigits: 3,
        })

    return {
      id: record.id,
      timestamp: timeStr,
      level: isErr ? 'ERROR' : 'INFO',
      subsystem: record.resource_type || 'platform',
      action: record.action,
      message: `Actor ${record.actor_id || 'system'} performed ${record.action} on ${
        record.resource_id || record.resource_type || 'resource'
      }${isErr ? ` (${record.result})` : ''}`,
      statusCode: isErr ? 400 : 200,
      details: record.details,
    }
  }

  // Load audit logs from backend
  const loadInitialLogs = async () => {
    try {
      const data = await api.get<AuditRecord[]>('/api/v1/audit-logs', { params: { limit: 50 } })
      if (Array.isArray(data) && data.length > 0) {
        setLogs(data.map(formatAuditLog))
      }
    } catch {
      // Backend offline or audit logs empty
    }
  }

  useEffect(() => {
    loadInitialLogs()
  }, [])

  // Live polling for new backend audit records
  useEffect(() => {
    if (!isStreaming) return

    const interval = setInterval(async () => {
      try {
        const data = await api.get<AuditRecord[]>('/api/v1/audit-logs', { params: { limit: 30 } })
        if (Array.isArray(data) && data.length > 0) {
          setLogs((prev) => {
            const existingIds = new Set(prev.map((l) => l.id))
            const newEntries = data.filter((r) => !existingIds.has(r.id)).map(formatAuditLog)

            if (newEntries.length === 0) return prev
            const combined = [...prev, ...newEntries]
            return combined.length > 500 ? combined.slice(combined.length - 500) : combined
          })
        }
      } catch {
        // Polling silently ignores transient network errors
      }
    }, 4000)

    return () => clearInterval(interval)
  }, [isStreaming])

  // Auto-scroll to bottom
  useEffect(() => {
    if (autoScroll && consoleEndRef.current) {
      consoleEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [logs, autoScroll])

  const handleSendPing = async () => {
    setIsPinging(true)
    const start = performance.now()
    const now = new Date().toLocaleTimeString('en-US', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3,
    })
    try {
      await api.get('/health', { timeoutMs: 4000 })
      const elapsed = Math.round(performance.now() - start)
      const entry: LogEntry = {
        id: `probe-${Date.now()}`,
        timestamp: now,
        level: 'INFO',
        subsystem: 'health',
        action: 'GET /health',
        message: `Live gateway health probe resolved in ${elapsed}ms (status: healthy)`,
        latencyMs: elapsed,
        statusCode: 200,
      }
      setLogs((prev) => [...prev, entry])
    } catch (err: unknown) {
      const elapsed = Math.round(performance.now() - start)
      const entry: LogEntry = {
        id: `probe-${Date.now()}`,
        timestamp: now,
        level: 'ERROR',
        subsystem: 'health',
        action: 'GET /health',
        message: `Live gateway health probe failed: ${err instanceof Error ? err.message : 'Timeout'}`,
        latencyMs: elapsed,
        statusCode: 503,
      }
      setLogs((prev) => [...prev, entry])
    } finally {
      setIsPinging(false)
    }
  }

  const handleClear = () => {
    setLogs([])
    setSelectedEntry(null)
  }

  const handleExport = () => {
    const raw = logs
      .map(
        (l) =>
          `[${l.timestamp}] [${l.level}] [${l.subsystem}] ${l.action} - ${l.message} ${
            l.statusCode ? `(${l.statusCode})` : ''
          } ${l.latencyMs ? `[${l.latencyMs}ms]` : ''}`
      )
      .join('\n')
    const blob = new Blob([raw], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `ekarouter-console-${new Date().toISOString().slice(0, 10)}.log`
    a.click()
    URL.revokeObjectURL(url)
  }

  const filteredLogs = logs.filter((log) => {
    const matchesLevel = filterLevel === 'ALL' || log.level === filterLevel
    const matchesSubsystem = filterSubsystem === 'ALL' || log.subsystem === filterSubsystem
    const matchesSearch =
      !searchQuery ||
      log.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.action.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.subsystem.toLowerCase().includes(searchQuery.toLowerCase())
    return matchesLevel && matchesSubsystem && matchesSearch
  })

  const getLevelBadgeClass = (lvl: LogEntry['level']) => {
    switch (lvl) {
      case 'ERROR':
        return 'text-rose-400 bg-rose-950/40 border-rose-800/40'
      case 'WARN':
        return 'text-amber-400 bg-amber-950/40 border-amber-800/40'
      case 'HTTP':
        return 'text-blue-400 bg-blue-950/40 border-blue-800/40'
      case 'DEBUG':
        return 'text-slate-400 bg-slate-800/40 border-slate-700/40'
      default:
        return 'text-emerald-400 bg-emerald-950/40 border-emerald-800/40'
    }
  }

  return (
    <div className={`space-y-3 ${isFullscreen ? 'fixed inset-0 z-50 p-4 bg-[var(--bg-app)]' : ''}`}>
      <PageHeader
        title="Live Server Console"
        description="Real-time streaming gateway logs, telemetry events, and system audit traces."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Live Console' },
        ]}
        metadata={
          <span className="flex items-center gap-2">
            <span
              className={`w-2 h-2 rounded-full ${
                isStreaming ? 'bg-[var(--status-success)] animate-pulse' : 'bg-amber-500'
              }`}
            />
            <span>{isStreaming ? 'Streaming Live' : 'Stream Paused'}</span>
            <span>&bull;</span>
            <span>{logs.length} events buffered</span>
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={() => setIsStreaming(!isStreaming)}
              leftIcon={isStreaming ? <Pause className="w-3.5 h-3.5" /> : <Play className="w-3.5 h-3.5" />}
            >
              {isStreaming ? 'Pause' : 'Resume'}
            </Button>
            <Button
              variant="secondary"
              size="compact"
              onClick={handleSendPing}
              isLoading={isPinging}
              leftIcon={<Send className="w-3.5 h-3.5" />}
            >
              Send Probe
            </Button>
            <Button
              variant="secondary"
              size="compact"
              onClick={handleExport}
              leftIcon={<Download className="w-3.5 h-3.5" />}
            >
              Export
            </Button>
            <Button
              variant="secondary"
              size="compact"
              onClick={handleClear}
              leftIcon={<Trash2 className="w-3.5 h-3.5" />}
            >
              Clear
            </Button>
            <button
              type="button"
              onClick={() => setIsFullscreen(!isFullscreen)}
              title={isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'}
              className="p-2 rounded-[6px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors border border-[var(--border-subtle)]"
            >
              {isFullscreen ? <Minimize2 className="w-3.5 h-3.5" /> : <Maximize2 className="w-3.5 h-3.5" />}
            </button>
          </div>
        }
      />

      {/* Filter and control bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-2.5 bg-[var(--bg-surface)] p-2.5 rounded-[8px] border border-[var(--border-subtle)] text-[12px]">
        <div className="flex items-center gap-1.5 overflow-x-auto w-full sm:w-auto pb-1 sm:pb-0 scrollbar-none">
          <span className="text-[11px] font-medium text-[var(--text-muted)] flex items-center gap-1 mr-1">
            <Filter className="w-3 h-3" />
            <span>Level:</span>
          </span>
          {['ALL', 'INFO', 'HTTP', 'WARN', 'ERROR', 'DEBUG'].map((lvl) => (
            <button
              key={lvl}
              type="button"
              onClick={() => setFilterLevel(lvl)}
              className={`px-2 py-0.5 rounded-[4px] font-mono text-[11px] transition-colors ${
                filterLevel === lvl
                  ? 'bg-[var(--brand-primary)] text-white font-semibold'
                  : 'text-[var(--text-secondary)] hover:bg-[var(--bg-panel)]'
              }`}
            >
              {lvl}
            </button>
          ))}
        </div>

        <div className="flex items-center gap-2 w-full sm:w-auto">
          <select
            value={filterSubsystem}
            onChange={(e) => setFilterSubsystem(e.target.value)}
            className="px-2 py-1 text-[11.5px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-secondary)]"
          >
            <option value="ALL">All Subsystems</option>
            {Array.from(new Set(logs.map((l) => l.subsystem).filter(Boolean))).map((sub) => (
              <option key={sub} value={sub}>
                {sub}
              </option>
            ))}
          </select>

          <div className="relative flex-1 sm:w-48">
            <Search className="w-3 h-3 absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--text-muted)] pointer-events-none" />
            <input
              type="text"
              placeholder="Grep logs..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-7 pr-2 py-1 text-[11.5px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)]"
            />
          </div>

          <button
            type="button"
            onClick={() => setAutoScroll(!autoScroll)}
            className={`px-2 py-1 rounded-[5px] text-[11px] font-medium border transition-colors whitespace-nowrap ${
              autoScroll
                ? 'bg-blue-950/40 border-blue-600/40 text-blue-300'
                : 'bg-[var(--bg-panel)] border-[var(--border-subtle)] text-[var(--text-muted)]'
            }`}
          >
            {autoScroll ? 'Autoscroll ON' : 'Autoscroll OFF'}
          </button>
        </div>
      </div>

      {/* Main Terminal Window */}
      <div
        ref={containerRef}
        className="rounded-[8px] bg-[#07090E] border border-[var(--border-subtle)] font-mono text-[12px] h-[580px] overflow-y-auto flex flex-col relative"
      >
        <div className="sticky top-0 bg-[#0B0E17]/95 backdrop-blur-xs border-b border-[var(--border-subtle)] px-3 py-1.5 flex items-center justify-between text-[11px] text-[var(--text-muted)] z-10 shrink-0 select-none">
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-red-500/80 inline-block" />
            <span className="w-2.5 h-2.5 rounded-full bg-yellow-500/80 inline-block" />
            <span className="w-2.5 h-2.5 rounded-full bg-green-500/80 inline-block" />
            <span className="ml-2 font-semibold text-[var(--text-secondary)]">
              stdout / server.log &bull; session active
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span>Encoding: UTF-8</span>
            <span>Protocol: SSE / REST</span>
          </div>
        </div>

        <div className="p-3 space-y-1 flex-1">
          {filteredLogs.map((log) => (
            <div
              key={log.id}
              onClick={() => setSelectedEntry(log.details ? log : null)}
              className={`group flex items-start gap-2.5 py-1 px-1.5 rounded-[4px] transition-colors leading-relaxed ${
                log.details ? 'cursor-pointer hover:bg-slate-800/40' : 'hover:bg-slate-900/40'
              }`}
            >
              <span className="text-slate-500 text-[11px] shrink-0 select-none">
                {log.timestamp}
              </span>

              <span
                className={`text-[10px] font-bold px-1.5 py-0.2 rounded border shrink-0 uppercase select-none ${getLevelBadgeClass(
                  log.level
                )}`}
              >
                {log.level}
              </span>

              <span className="text-blue-400/90 text-[11px] shrink-0">
                [{log.subsystem}]
              </span>

              <span className="text-slate-300 font-semibold text-[11.5px] shrink-0">
                {log.action}
              </span>

              <span className="text-slate-400 flex-1 truncate group-hover:text-slate-200">
                {log.message}
              </span>

              {log.latencyMs !== undefined && (
                <span className="text-slate-500 text-[10.5px] shrink-0">
                  {log.latencyMs}ms
                </span>
              )}

              {log.statusCode !== undefined && (
                <span
                  className={`text-[10px] px-1 py-0.2 rounded font-bold shrink-0 ${
                    log.statusCode >= 400
                      ? 'text-rose-400 bg-rose-950/40'
                      : 'text-emerald-400 bg-emerald-950/40'
                  }`}
                >
                  {log.statusCode}
                </span>
              )}
            </div>
          ))}

          {logs.length === 0 ? (
            <div className="py-24 text-center flex flex-col items-center justify-center gap-3">
              <div className="w-10 h-10 rounded-full bg-slate-900 border border-slate-800 flex items-center justify-center text-slate-400">
                <Send className="w-4 h-4" />
              </div>
              <div className="space-y-1">
                <p className="text-[13px] font-semibold text-slate-200">Console buffer empty</p>
                <p className="text-[11.5px] text-slate-400 max-w-sm mx-auto">
                  No gateway events recorded yet. Perform actions in EkaRouter or dispatch a probe to test live responses.
                </p>
              </div>
              <Button
                variant="secondary"
                size="compact"
                onClick={handleSendPing}
                isLoading={isPinging}
                leftIcon={<Send className="w-3.5 h-3.5" />}
              >
                Dispatch Live Probe
              </Button>
            </div>
          ) : filteredLogs.length === 0 ? (
            <div className="py-20 text-center text-slate-500 text-[12px]">
              No log records match the current filter criteria.
            </div>
          ) : null}

          <div ref={consoleEndRef} />
        </div>
      </div>

      {/* Expandable JSON Detail Drawer / Modal if clicked */}
      {selectedEntry && selectedEntry.details && (
        <div className="p-3 rounded-[8px] bg-[var(--bg-surface)] border border-[var(--border-strong)] text-[12px] space-y-2 animate-in fade-in-50">
          <div className="flex items-center justify-between">
            <span className="font-semibold text-[var(--text-primary)]">
              Event Details &bull; {selectedEntry.action}
            </span>
            <button
              type="button"
              onClick={() => setSelectedEntry(null)}
              className="text-[var(--text-muted)] hover:text-[var(--text-primary)]"
            >
              Close
            </button>
          </div>
          <pre className="p-2.5 rounded-[5px] bg-[var(--bg-input)] border border-[var(--border-subtle)] font-mono text-[11px] text-sky-300 overflow-x-auto">
            {JSON.stringify(selectedEntry.details, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}
