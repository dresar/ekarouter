import { useState } from 'react'
import { Zap, Copy, Check, AlertCircle } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

interface CompactionResult {
  original_bytes: number
  compact_bytes: number
  reduction_pct: number
  output: string
}

export function TokenSaverPage() {
  const [inputText, setInputText] = useState(
    `diff --git a/internal/gateway/gateway.go b/internal/gateway/gateway.go\nindex 8f3a1b2..9c4d2e1 100644\n--- a/internal/gateway/gateway.go\n+++ b/internal/gateway/gateway.go\n@@ -45,7 +45,7 @@ func (g *Gateway) ChatCompletions(w http.ResponseWriter, r *http.Request) {\n-\tlog.Println("Handling request")\n+\t// Optimized without redundant log line\n \tw.Header().Set("Content-Type", "application/json")\n \tw.WriteHeader(http.StatusOK)\n }\n`
  )
  const [mode, setMode] = useState<'safe' | 'balanced' | 'aggressive'>('safe')
  const [result, setResult] = useState<CompactionResult | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const handleCompact = async () => {
    if (!inputText.trim()) return
    setIsLoading(true)
    setError(null)

    try {
      const data = await api.post<CompactionResult>('/api/tokensaver/preview', {
        input: inputText,
      })
      setResult(data)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Compaction preview failed'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCopy = async () => {
    if (!result?.output) return
    try {
      await navigator.clipboard.writeText(result.output)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {}
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Token Saver"
        description="Heuristic prompt and code-diff compaction engine to reduce token consumption."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Token Saver' },
        ]}
        metadata={
          <span>
            Compaction algorithm active
          </span>
        }
        actions={
          <Button
            variant="primary"
            size="compact"
            onClick={handleCompact}
            isLoading={isLoading}
            leftIcon={<Zap className="w-3.5 h-3.5" />}
          >
            Run Compaction
          </Button>
        }
      />

      {error && (
        <div className="p-3 rounded-[6px] bg-[var(--status-danger)]/10 border border-[var(--status-danger)]/30 flex items-center gap-2 text-[12.5px] text-[var(--status-danger)]">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      <div className="flex items-center gap-3 bg-[var(--bg-surface)] p-2.5 rounded-[8px] border border-[var(--border-subtle)] text-[12px]">
        <span className="font-semibold uppercase text-[11px] text-[var(--text-muted)]">
          Compaction Mode:
        </span>
        <div className="flex items-center gap-3">
          {(['safe', 'balanced', 'aggressive'] as const).map((m) => (
            <label key={m} className="inline-flex items-center gap-1.5 cursor-pointer select-none">
              <input
                type="radio"
                name="mode"
                value={m}
                checked={mode === m}
                onChange={() => setMode(m)}
                className="text-[var(--brand-primary)] accent-blue-600"
              />
              <span className="capitalize font-medium text-[var(--text-secondary)]">{m}</span>
            </label>
          ))}
        </div>
      </div>

      {result && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[7px] p-3">
            <span className="text-[10px] font-semibold uppercase text-[var(--text-muted)] block">
              Original Size
            </span>
            <span className="text-[15px] font-mono font-bold text-[var(--text-primary)]">
              {result.original_bytes} B
            </span>
          </div>
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[7px] p-3">
            <span className="text-[10px] font-semibold uppercase text-[var(--text-muted)] block">
              Compacted Size
            </span>
            <span className="text-[15px] font-mono font-bold text-[var(--brand-text)]">
              {result.compact_bytes} B
            </span>
          </div>
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[7px] p-3">
            <span className="text-[10px] font-semibold uppercase text-[var(--text-muted)] block">
              Bytes Saved
            </span>
            <span className="text-[15px] font-mono font-bold text-[var(--status-success)]">
              {Math.max(0, result.original_bytes - result.compact_bytes)} B
            </span>
          </div>
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[7px] p-3">
            <span className="text-[10px] font-semibold uppercase text-[var(--text-muted)] block">
              Reduction Ratio
            </span>
            <span className="text-[15px] font-mono font-bold text-[var(--status-success)]">
              {typeof result.reduction_pct === 'number' ? result.reduction_pct.toFixed(1) : '0.0'}%
            </span>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <label className="text-[11px] font-semibold uppercase text-[var(--text-muted)]">
              Raw Input Prompt / Diff
            </label>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">
              {inputText.length} characters
            </span>
          </div>
          <textarea
            rows={12}
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            placeholder="Paste code diff, JSON log, or prompt text here..."
            className="w-full p-3 font-mono text-[12px] rounded-[6px] bg-[var(--bg-card)] border border-[var(--border-subtle)] focus:outline-none resize-y leading-relaxed"
          />
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <label className="text-[11px] font-semibold uppercase text-[var(--text-muted)]">
              Compacted Output Preview
            </label>
            {result?.output && (
              <button
                type="button"
                onClick={handleCopy}
                className="text-[11px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 font-medium"
              >
                {copied ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-[var(--status-success)]" />
                    <span>Copied</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>Copy Compacted</span>
                  </>
                )}
              </button>
            )}
          </div>
          <textarea
            readOnly
            rows={12}
            value={result?.output || ''}
            placeholder="Compacted output will appear here after execution..."
            className="w-full p-3 font-mono text-[12px] rounded-[6px] border border-[var(--border-subtle)] focus:outline-none resize-y bg-[#07090E] text-[var(--text-secondary)] leading-relaxed"
          />
        </div>
      </div>
    </div>
  )
}
