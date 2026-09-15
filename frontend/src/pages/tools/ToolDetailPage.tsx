import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Play, Wrench, AlertCircle, Clock } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { ToolDefinition } from '../../types/api.ts'

interface ExecutionResponse {
  status_code: number
  latency_ms: number
  body: unknown
  error?: string
}

export function ToolDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [tool, setTool] = useState<ToolDefinition | null>(null)
  const [error, setError] = useState<string | null>(null)

  const [paramsJson, setParamsJson] = useState('{\n  "to": "user@example.com",\n  "subject": "Hello from EkaRouter"\n}')
  const [isExecuting, setIsExecuting] = useState(false)
  const [executionResult, setExecutionResult] = useState<ExecutionResponse | null>(null)
  const [execError, setExecError] = useState<string | null>(null)

  useEffect(() => {
    async function loadData() {
      if (!id) return
      setError(null)
      try {
        const toolData = await api.get<ToolDefinition>(`/api/v1/tools/${id}`)
        setTool(toolData)
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : 'Failed to load tool details'
        setError(msg)
      }
    }

    loadData()
  }, [id])

  const handleExecute = async () => {
    if (!id) return
    setIsExecuting(true)
    setExecError(null)
    setExecutionResult(null)

    let parsed = {}
    try {
      if (paramsJson.trim()) {
        parsed = JSON.parse(paramsJson)
      }
    } catch {
      setExecError('Invalid JSON format in parameters')
      setIsExecuting(false)
      return
    }

    try {
      const data = await api.post<ExecutionResponse>(`/api/v1/tools/${id}/execute`, {
        parameters: parsed,
      })
      setExecutionResult(data)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Tool execution failed'
      setExecError(msg)
    } finally {
      setIsExecuting(false)
    }
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={tool ? tool.name : 'Tool Inspector & Execution'}
        description={`Tool UUID: ${id || ''}`}
        breadcrumbs={[
          { label: 'Tools', to: '/tools' },
          { label: tool?.name || 'Detail' },
        ]}
        actions={
          <Button
            variant="ghost"
            size="compact"
            onClick={() => navigate('/tools')}
            leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}
          >
            Kembali
          </Button>
        }
      />

      {error && <ErrorBanner message={error} />}

      {tool && (
        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-5">
          <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)]">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded bg-[var(--bg-panel)] flex items-center justify-center text-[var(--brand-text)]">
                <Wrench className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-[16px] font-semibold text-[var(--text-primary)]">
                  {tool.name}
                </h2>
                <div className="flex items-center gap-2 mt-0.5 text-[11px] font-mono text-[var(--text-muted)]">
                  <span>provider: {tool.provider_id}</span>
                  <span>•</span>
                  <span>category: {tool.category}</span>
                </div>
              </div>
            </div>

            <span
              className={`font-mono text-[11px] font-bold px-2 py-0.5 rounded border ${
                tool.method === 'GET'
                  ? 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
                  : tool.method === 'POST'
                  ? 'text-blue-400 bg-blue-500/10 border-blue-500/20'
                  : 'text-amber-400 bg-amber-500/10 border-amber-500/20'
              }`}
            >
              {tool.method}
            </span>
          </div>

          <div className="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-3 text-[12px]">
            <div>
              <span className="text-[10.5px] uppercase font-semibold text-[var(--text-muted)] block">
                URL Template
              </span>
              <span className="font-mono text-[var(--text-secondary)] break-all">
                {tool.url_template}
              </span>
            </div>
            <div>
              <span className="text-[10.5px] uppercase font-semibold text-[var(--text-muted)] block">
                Timeout & Retries
              </span>
              <span className="font-mono text-[var(--text-primary)]">
                {tool.timeout_ms}ms • {tool.retry_count} retries
              </span>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
              Execution Test Parameters
            </h3>
            <span className="text-[11px] text-[var(--text-muted)] font-mono">JSON format</span>
          </div>

          {execError && (
            <div className="p-2.5 rounded bg-[var(--status-danger)]/10 text-[var(--status-danger)] text-[12px] flex items-center gap-1.5">
              <AlertCircle className="w-3.5 h-3.5" />
              <span>{execError}</span>
            </div>
          )}

          <textarea
            rows={10}
            value={paramsJson}
            onChange={(e) => setParamsJson(e.target.value)}
            className="w-full p-3 font-mono text-[12px] rounded-[6px] focus:outline-none resize-y leading-relaxed"
          />

          <Button
            variant="primary"
            size="compact"
            onClick={handleExecute}
            isLoading={isExecuting}
            leftIcon={<Play className="w-3.5 h-3.5" />}
            className="w-full"
          >
            Jalankan
          </Button>
        </div>

        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
              Execution Output
            </h3>
            {executionResult && (
              <div className="flex items-center gap-2 text-[11px] font-mono">
                <span className="flex items-center gap-1 text-[var(--status-success)]">
                  <Clock className="w-3 h-3" />
                  <span>{executionResult.latency_ms}ms</span>
                </span>
                <span
                  className={`px-1.5 py-0.2 rounded font-bold ${
                    executionResult.status_code >= 200 && executionResult.status_code < 300
                      ? 'bg-emerald-500/10 text-emerald-400'
                      : 'bg-rose-500/10 text-rose-400'
                  }`}
                >
                  HTTP {executionResult.status_code}
                </span>
              </div>
            )}
          </div>

          <div className="h-[250px] overflow-y-auto bg-[var(--bg-input)] p-3 rounded-[6px] border border-[var(--border-strong)] text-[11.5px] font-mono text-[var(--text-secondary)]">
            {executionResult ? (
              <pre className="whitespace-pre-wrap leading-relaxed select-all">
                {JSON.stringify(executionResult.body || executionResult, null, 2)}
              </pre>
            ) : (
              <span className="text-[var(--text-muted)]">
                Run an execution to inspect response payload and status codes.
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
