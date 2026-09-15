import { useState } from 'react'
import {
  Zap,
  Copy,
  Check,
  AlertCircle,
  ShieldCheck,
  FileCode,
  FileJson,
  MessageSquare,
  Trash2,
  Terminal,
  Sparkles,
  Layers,
  ChevronRight,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

interface CompactionResult {
  original_bytes: number
  compact_bytes: number
  reduction_pct: number
  original_tokens: number
  compact_tokens: number
  estimated_tokens_saved: number
  mode: string
  transformations: string[]
  output: string
}

type ModeType = 'safe' | 'balanced' | 'aggressive' | 'off'

interface PresetItem {
  id: string
  label: string
  description: string
  mode: ModeType
  icon: typeof FileCode
  content: string
}

const PRESETS: PresetItem[] = [
  {
    id: 'git-diff',
    label: 'Git Diff Patch',
    description: 'Large code diff with repeated hunks',
    mode: 'safe',
    icon: FileCode,
    content: `diff --git a/internal/gateway/gateway.go b/internal/gateway/gateway.go
index 8f3a1b2..9c4d2e1 100644
--- a/internal/gateway/gateway.go
+++ b/internal/gateway/gateway.go
@@ -45,120 +45,6 @@ func (g *Gateway) ChatCompletions(w http.ResponseWriter, r *http.Request) {
-	log.Println("Handling request start")
-	log.Println("Handling request validate")
-	log.Println("Handling request auth")
+	w.Header().Set("Content-Type", "application/json")
+	w.WriteHeader(http.StatusOK)
+	recordMetric("gateway_metric_0", time.Now().UnixNano())
+	recordMetric("gateway_metric_1", time.Now().UnixNano())
+	recordMetric("gateway_metric_2", time.Now().UnixNano())
+	recordMetric("gateway_metric_3", time.Now().UnixNano())
+	recordMetric("gateway_metric_4", time.Now().UnixNano())
+	recordMetric("gateway_metric_5", time.Now().UnixNano())
+	recordMetric("gateway_metric_6", time.Now().UnixNano())
+	recordMetric("gateway_metric_7", time.Now().UnixNano())
+	recordMetric("gateway_metric_8", time.Now().UnixNano())
+	recordMetric("gateway_metric_9", time.Now().UnixNano())
+	recordMetric("gateway_metric_10", time.Now().UnixNano())
+	recordMetric("gateway_metric_11", time.Now().UnixNano())
+	recordMetric("gateway_metric_12", time.Now().UnixNano())
+	recordMetric("gateway_metric_13", time.Now().UnixNano())
+	recordMetric("gateway_metric_14", time.Now().UnixNano())
+	recordMetric("gateway_metric_15", time.Now().UnixNano())
+	recordMetric("gateway_metric_16", time.Now().UnixNano())
+	recordMetric("gateway_metric_17", time.Now().UnixNano())
+	recordMetric("gateway_metric_18", time.Now().UnixNano())
+	recordMetric("gateway_metric_19", time.Now().UnixNano())
+	recordMetric("gateway_metric_20", time.Now().UnixNano())
+	recordMetric("gateway_metric_21", time.Now().UnixNano())
+	recordMetric("gateway_metric_22", time.Now().UnixNano())
+	recordMetric("gateway_metric_23", time.Now().UnixNano())
+	recordMetric("gateway_metric_24", time.Now().UnixNano())
+	recordMetric("gateway_metric_25", time.Now().UnixNano())
+	recordMetric("gateway_metric_26", time.Now().UnixNano())
+	recordMetric("gateway_metric_27", time.Now().UnixNano())
+	recordMetric("gateway_metric_28", time.Now().UnixNano())
+	recordMetric("gateway_metric_29", time.Now().UnixNano())
+	recordMetric("gateway_metric_30", time.Now().UnixNano())
+	recordMetric("gateway_metric_31", time.Now().UnixNano())
+	recordMetric("gateway_metric_32", time.Now().UnixNano())
+	recordMetric("gateway_metric_33", time.Now().UnixNano())
+	recordMetric("gateway_metric_34", time.Now().UnixNano())
+	recordMetric("gateway_metric_35", time.Now().UnixNano())
+	recordMetric("gateway_metric_36", time.Now().UnixNano())
+	recordMetric("gateway_metric_37", time.Now().UnixNano())
+	recordMetric("gateway_metric_38", time.Now().UnixNano())
+	recordMetric("gateway_metric_39", time.Now().UnixNano())
+	recordMetric("gateway_metric_40", time.Now().UnixNano())
+	recordMetric("gateway_metric_41", time.Now().UnixNano())
+	recordMetric("gateway_metric_42", time.Now().UnixNano())
+	recordMetric("gateway_metric_43", time.Now().UnixNano())
+	recordMetric("gateway_metric_44", time.Now().UnixNano())
+	recordMetric("gateway_metric_45", time.Now().UnixNano())
\\ No newline at end of file
diff --git a/internal/routing/router.go b/internal/routing/router.go
index 112233..445566 100644
--- a/internal/routing/router.go
+++ b/internal/routing/router.go
@@ -10,15 +10,4 @@
-	traceA()
-	traceB()
-	traceC()
-	traceD()
-	traceE()
-	traceF()
-	traceG()
-	traceH()
-	traceI()
-	traceJ()
-	traceK()
+	router.Sync()`,
  },
  {
    id: 'server-logs',
    label: 'Logs + Protected Trace',
    description: 'Repeated lines with fail-safe panic protection',
    mode: 'balanced',
    icon: Terminal,
    content: `2026-09-15 15:30:00 [INFO] Cluster worker initialized across 16 CPU cores
2026-09-15 15:30:01 [WARN] Connection timeout connecting to upstream endpoint replica-03; retry scheduled in 500ms
2026-09-15 15:30:02 [WARN] Connection timeout connecting to upstream endpoint replica-03; retry scheduled in 500ms
2026-09-15 15:30:03 [WARN] Connection timeout connecting to upstream endpoint replica-03; retry scheduled in 500ms
2026-09-15 15:30:04 [WARN] Connection timeout connecting to upstream endpoint replica-03; retry scheduled in 500ms
2026-09-15 15:30:05 [WARN] Connection timeout connecting to upstream endpoint replica-03; retry scheduled in 500ms
2026-09-15 15:30:06 [WARN] Connection timeout connecting to upstream endpoint replica-03; retry scheduled in 500ms
2026-09-15 15:30:07 [WARN] Connection timeout connecting to upstream endpoint replica-03; retry scheduled in 500ms
2026-09-15 15:30:08 [WARN] Connection timeout connecting to upstream endpoint replica-03; retry scheduled in 500ms
2026-09-15 15:30:09 [FATAL] panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x7f9a1b]

goroutine 42 [running]:
main.processBatch(0xc000124000, 0x10, 0x20)
	/app/worker/processor.go:88 +0x145
main.workerLoop()
	/app/worker/loop.go:34 +0x7a
created by main.main in goroutine 1
	/app/main.go:52 +0x180`,
  },
  {
    id: 'json-payload',
    label: 'Formatted JSON',
    description: 'Multi-level indented payload minification',
    mode: 'aggressive',
    icon: FileJson,
    content: `{
  "gateway": "EkaRouter Enterprise Edge",
  "version": "2.4.0",
  "environment": "production",
  "routing": {
    "default_strategy": "priority_failover",
    "timeout_ms": 30000,
    "max_retries": 3,
    "circuit_breaker": {
      "enabled": true,
      "failure_threshold": 5,
      "cooldown_seconds": 60
    }
  },
  "providers": [
    {
      "name": "claude-9router",
      "priority": 1,
      "weight": 100,
      "supported_models": [
        "claude-3-7-sonnet-20250219",
        "claude-3-5-sonnet-20241022"
      ]
    },
    {
      "name": "gemini-flash",
      "priority": 2,
      "weight": 80,
      "supported_models": [
        "gemini-2.5-flash",
        "gemini-2.5-pro"
      ]
    }
  ]
}`,
  },
  {
    id: 'chat-prompt',
    label: 'Chat Conversation',
    description: 'Excessive whitespace & blank line cleanup',
    mode: 'balanced',
    icon: MessageSquare,
    content: `System: You are an expert cloud infrastructure and security engineer.

Please inspect the following configuration thoroughly.   



User: Can you check if my kubernetes service configuration is production-ready?  


User: Can you check if my kubernetes service configuration is production-ready?  


User: Can you check if my kubernetes service configuration is production-ready?  


User: Can you check if my kubernetes service configuration is production-ready?  



Assistant: Reviewing your service definition...`,
  },
]

export function TokenSaverPage() {
  const [inputText, setInputText] = useState(PRESETS[0].content)
  const [mode, setMode] = useState<ModeType>('balanced')
  const [result, setResult] = useState<CompactionResult | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [activeSnippetTab, setActiveSnippetTab] = useState<'curl' | 'python' | 'node'>('curl')

  const executeCompaction = async (text: string, targetMode: ModeType) => {
    if (!text.trim()) return
    setIsLoading(true)
    setError(null)

    try {
      const data = await api.post<CompactionResult>('/api/tokensaver/preview', {
        input: text,
        mode: targetMode,
      })
      setResult(data)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Compaction preview failed'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCompact = () => {
    executeCompaction(inputText, mode)
  }

  const handleSelectPreset = (preset: PresetItem) => {
    setInputText(preset.content)
    setMode(preset.mode)
    executeCompaction(preset.content, preset.mode)
  }

  const handleCopy = async () => {
    if (!result?.output) return
    try {
      await navigator.clipboard.writeText(result.output)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {}
  }

  const handleClear = () => {
    setInputText('')
    setResult(null)
  }

  const estInputTokens = Math.ceil(inputText.length / 4)
  const estOutputTokens = result ? result.compact_tokens : 0

  return (
    <div className="space-y-4">
      <PageHeader
        title="Token Saver"
        description="High-performance prompt and diff compaction engine to slash LLM context window costs."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Token Saver' },
        ]}
        metadata={
          <span className="flex items-center gap-1.5 text-[11.5px] text-[var(--status-success)]">
            <span className="w-2 h-2 rounded-full bg-[var(--status-success)] animate-pulse" />
            Deterministic fail-safe engine active
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

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-3">
        {(
          [
            {
              id: 'safe' as const,
              title: 'Safe Mode',
              badge: 'Non-Destructive',
              desc: 'Compresses git diff hunks, deduplicates repeated logs (≥3x), and guarantees zero semantic mutation.',
            },
            {
              id: 'balanced' as const,
              title: 'Balanced Mode',
              badge: 'Recommended',
              desc: 'Safe + collapses consecutive blank lines, trims trailing whitespace, and truncates boundary over 15k.',
            },
            {
              id: 'aggressive' as const,
              title: 'Aggressive Mode',
              badge: 'Maximum Savings',
              desc: 'Balanced + minifies formatted JSON blocks, removes all empty line redundancy, and truncates over 8k.',
            },
            {
              id: 'off' as const,
              title: 'Pass-Through (Off)',
              badge: 'Disabled',
              desc: 'Bypasses TokenSaver entirely. Request content is delivered to upstream LLMs untouched.',
            },
          ] as const
        ).map((item) => {
          const isSelected = mode === item.id
          return (
            <button
              key={item.id}
              type="button"
              onClick={() => {
                setMode(item.id)
                if (inputText.trim()) {
                  executeCompaction(inputText, item.id)
                }
              }}
              className={`text-left p-3 rounded-[8px] border transition-all cursor-pointer ${
                isSelected
                  ? 'bg-[var(--bg-card)] border-[var(--brand-primary)] ring-1 ring-[var(--brand-primary)]/40 shadow-sm'
                  : 'bg-[var(--bg-card)] border-[var(--border-subtle)] hover:border-[var(--border-strong)]'
              }`}
            >
              <div className="flex items-center justify-between gap-1 mb-1">
                <span className="text-[13px] font-semibold text-[var(--text-primary)]">
                  {item.title}
                </span>
                <span
                  className={`text-[9.5px] px-1.5 py-0.5 rounded-[4px] font-semibold uppercase tracking-wider ${
                    item.id === 'balanced'
                      ? 'bg-blue-500/15 text-blue-400 border border-blue-500/30'
                      : item.id === 'aggressive'
                      ? 'bg-purple-500/15 text-purple-400 border border-purple-500/30'
                      : isSelected
                      ? 'bg-[var(--bg-surface)] text-[var(--text-primary)] border border-[var(--border-strong)]'
                      : 'bg-[var(--bg-surface)] text-[var(--text-muted)]'
                  }`}
                >
                  {item.badge}
                </span>
              </div>
              <p className="text-[11.5px] leading-relaxed text-[var(--text-muted)] line-clamp-2">
                {item.desc}
              </p>
            </button>
          )
        })}
      </div>

      <div className="bg-[var(--bg-card)] p-3 rounded-[8px] border border-[var(--border-subtle)]">
        <div className="flex items-center justify-between gap-2 mb-2">
          <div className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">
            <Sparkles className="w-3.5 h-3.5 text-[var(--brand-primary)]" />
            <span>1-Click Test Scenarios</span>
          </div>
          <span className="text-[11px] text-[var(--text-muted)]">Click preset to test compaction instantly</span>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          {PRESETS.map((preset) => {
            const Icon = preset.icon
            return (
              <button
                key={preset.id}
                type="button"
                onClick={() => handleSelectPreset(preset)}
                className="flex items-center gap-2.5 p-2 rounded-[6px] border border-[var(--border-subtle)] bg-[var(--bg-surface)] hover:bg-[var(--border-subtle)]/40 hover:border-[var(--border-strong)] text-left transition-colors cursor-pointer group"
              >
                <div className="w-7 h-7 rounded-[5px] bg-[var(--bg-card)] border border-[var(--border-subtle)] flex items-center justify-center shrink-0 text-[var(--text-secondary)] group-hover:text-[var(--brand-primary)]">
                  <Icon className="w-3.5 h-3.5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-[12px] font-medium text-[var(--text-primary)] truncate">
                    {preset.label}
                  </div>
                  <div className="text-[10.5px] text-[var(--text-muted)] truncate">
                    {preset.description}
                  </div>
                </div>
                <ChevronRight className="w-3.5 h-3.5 text-[var(--text-muted)] shrink-0 opacity-40 group-hover:opacity-100" />
              </button>
            )
          })}
        </div>
      </div>

      {result && (
        <div className="space-y-3">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] block">
                Original Payload
              </span>
              <div className="flex items-baseline gap-2 mt-1">
                <span className="text-[17px] font-mono font-bold text-[var(--text-primary)]">
                  {result.original_bytes.toLocaleString()} B
                </span>
                <span className="text-[11px] font-mono text-[var(--text-muted)]">
                  ≈ {result.original_tokens.toLocaleString()} tok
                </span>
              </div>
            </div>

            <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] block">
                Compacted Output
              </span>
              <div className="flex items-baseline gap-2 mt-1">
                <span className="text-[17px] font-mono font-bold text-[var(--brand-text)]">
                  {result.compact_bytes.toLocaleString()} B
                </span>
                <span className="text-[11px] font-mono text-[var(--text-muted)]">
                  ≈ {result.compact_tokens.toLocaleString()} tok
                </span>
              </div>
            </div>

            <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] block">
                Estimated Tokens Saved
              </span>
              <div className="flex items-baseline gap-2 mt-1">
                <span className="text-[17px] font-mono font-bold text-[var(--status-success)]">
                  +{result.estimated_tokens_saved.toLocaleString()} tok
                </span>
                <span className="text-[11px] font-mono text-[var(--status-success)]/80">
                  (-{Math.max(0, result.original_bytes - result.compact_bytes).toLocaleString()} B)
                </span>
              </div>
            </div>

            <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] block">
                Reduction Ratio
              </span>
              <div className="flex items-baseline gap-2 mt-1">
                <span className="text-[17px] font-mono font-bold text-[var(--status-success)]">
                  {result.reduction_pct.toFixed(1)}%
                </span>
                <span className="text-[10.5px] px-1.5 py-0.5 rounded-[4px] bg-[var(--status-success)]/10 text-[var(--status-success)] font-medium">
                  {result.reduction_pct > 0 ? 'Optimal' : 'Identical'}
                </span>
              </div>
            </div>
          </div>

          <div className="p-3 rounded-[8px] bg-[var(--bg-card)] border border-[var(--border-subtle)] space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Layers className="w-3.5 h-3.5 text-[var(--brand-primary)]" />
                <span className="text-[11.5px] font-semibold uppercase tracking-wider text-[var(--text-secondary)]">
                  Transformations Applied ({result.transformations?.length || 0})
                </span>
              </div>
              <div className="flex items-center gap-1.5 text-[11px] text-[var(--status-success)]">
                <ShieldCheck className="w-3.5 h-3.5" />
                <span>Fail-Open Guaranteed</span>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              {result.transformations && result.transformations.length > 0 ? (
                result.transformations.map((tf, idx) => (
                  <span
                    key={idx}
                    className="inline-flex items-center gap-1 px-2.5 py-1 rounded-[5px] text-[11.5px] font-mono bg-[var(--bg-surface)] border border-[var(--border-subtle)] text-[var(--text-primary)]"
                  >
                    <span className="w-1.5 h-1.5 rounded-full bg-[var(--brand-primary)]" />
                    {tf}
                  </span>
                ))
              ) : (
                <span className="text-[12px] text-[var(--text-muted)] italic">
                  No transformations were required.
                </span>
              )}
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <label className="text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">
                Raw Input Prompt / Payload
              </label>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-[11px] font-mono text-[var(--text-muted)]">
                {inputText.length.toLocaleString()} chars · ≈ {estInputTokens.toLocaleString()} tok
              </span>
              <button
                type="button"
                onClick={handleClear}
                className="text-[11px] text-[var(--text-muted)] hover:text-[var(--status-danger)] inline-flex items-center gap-1 font-medium transition-colors cursor-pointer"
              >
                <Trash2 className="w-3 h-3" />
                <span>Clear</span>
              </button>
            </div>
          </div>
          <textarea
            rows={14}
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            placeholder="Paste raw prompt, git diff, logs, or payload here..."
            className="w-full p-3 font-mono text-[12px] rounded-[6px] bg-[var(--bg-card)] border border-[var(--border-subtle)] focus:border-[var(--brand-primary)] focus:outline-none resize-y leading-relaxed text-[var(--text-primary)]"
          />
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <label className="text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">
                Compacted Output
              </label>
              {result && (
                <span className="text-[10px] px-1.5 py-0.2 rounded-[4px] bg-[var(--bg-surface)] text-[var(--brand-text)] font-mono font-medium border border-[var(--border-subtle)]">
                  Mode: {result.mode}
                </span>
              )}
            </div>
            <div className="flex items-center gap-3">
              {result?.output && (
                <span className="text-[11px] font-mono text-[var(--text-muted)]">
                  {result.output.length.toLocaleString()} chars · ≈ {estOutputTokens.toLocaleString()} tok
                </span>
              )}
              {result?.output && (
                <button
                  type="button"
                  onClick={handleCopy}
                  className="text-[11px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 font-medium cursor-pointer"
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
          </div>
          <textarea
            readOnly
            rows={14}
            value={result?.output || ''}
            placeholder="Compacted output will appear here after clicking 'Run Compaction' or selecting a scenario..."
            className="w-full p-3 font-mono text-[12px] rounded-[6px] border border-[var(--border-subtle)] focus:outline-none resize-y bg-[#07090E] text-[var(--text-secondary)] leading-relaxed"
          />
        </div>
      </div>

      <div className="bg-[var(--bg-card)] rounded-[8px] border border-[var(--border-subtle)] p-4 space-y-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-[var(--border-subtle)] pb-3">
          <div>
            <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">
              Gateway Integration Guide
            </h3>
            <p className="text-[11.5px] text-[var(--text-muted)]">
              Enable TokenSaver per-request with the standard <code className="text-[var(--brand-text)]">X-Token-Saver</code> HTTP header.
            </p>
          </div>
          <div className="flex items-center gap-1 bg-[var(--bg-surface)] p-1 rounded-[6px] border border-[var(--border-subtle)]">
            {(['curl', 'python', 'node'] as const).map((tab) => (
              <button
                key={tab}
                type="button"
                onClick={() => setActiveSnippetTab(tab)}
                className={`px-2.5 py-1 rounded-[4px] text-[11.5px] font-medium transition-colors cursor-pointer ${
                  activeSnippetTab === tab
                    ? 'bg-[var(--bg-card)] text-[var(--text-primary)] shadow-xs'
                    : 'text-[var(--text-muted)] hover:text-[var(--text-secondary)]'
                }`}
              >
                {tab === 'curl' ? 'cURL' : tab === 'python' ? 'Python (OpenAI SDK)' : 'Node.js (TypeScript)'}
              </button>
            ))}
          </div>
        </div>

        <div className="relative">
          <pre className="p-3.5 rounded-[6px] bg-[#07090E] border border-[var(--border-subtle)] text-[12px] font-mono text-[var(--text-secondary)] overflow-x-auto leading-relaxed">
            {activeSnippetTab === 'curl' &&
`curl -X POST http://localhost:8080/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_EKAROUTER_KEY" \\
  -H "X-Token-Saver: ${mode}" \\
  -d '{
    "model": "claude-3-7-sonnet",
    "messages": [
      {"role": "user", "content": "Analyze this git diff and logs..."}
    ]
  }'`}

            {activeSnippetTab === 'python' &&
`from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="YOUR_EKAROUTER_KEY",
    default_headers={"X-Token-Saver": "${mode}"}
)

response = client.chat.completions.create(
    model="claude-3-7-sonnet",
    messages=[{"role": "user", "content": "Analyze this git diff and logs..."}]
)
print(response.choices[0].message.content)`}

            {activeSnippetTab === 'node' &&
`import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: 'YOUR_EKAROUTER_KEY',
  defaultHeaders: {
    'X-Token-Saver': '${mode}',
  },
});

const response = await client.chat.completions.create({
  model: 'claude-3-7-sonnet',
  messages: [{ role: 'user', content: 'Analyze this git diff and logs...' }],
});
console.log(response.choices[0].message.content);`}
          </pre>
        </div>
      </div>
    </div>
  )
}
