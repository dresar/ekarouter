import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Zap, Gauge, ArrowRight, Activity, ShieldCheck, CheckCircle2 } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { MetricCard } from '../../components/ui/MetricCard.tsx'

const STORAGE_KEY = 'ekarouter_boost_config'

interface BoostState {
  fastLane: boolean
  tokenSaver: boolean
  semanticCache: boolean
  haFailover: boolean
}

function loadBoostState(): BoostState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw)
  } catch {}
  return {
    fastLane: true,
    tokenSaver: true,
    semanticCache: false,
    haFailover: true,
  }
}

export function BoostPage() {
  const [config, setConfig] = useState<BoostState>(loadBoostState)
  const [savedFeedback, setSavedFeedback] = useState(false)

  const toggleFeature = (key: keyof BoostState) => {
    const updated = { ...config, [key]: !config[key] }
    setConfig(updated)
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated))
    } catch {}
    setSavedFeedback(true)
    setTimeout(() => setSavedFeedback(false), 2000)
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Gateway Boost"
        description="Akselerasi throughput dan kompresi token gateway."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Boost' },
        ]}
        metadata={
          <span className="flex items-center gap-1.5 text-emerald-400">
            <CheckCircle2 className="w-3.5 h-3.5" />
            Turbo Aktif
          </span>
        }
        actions={
          <Link to="/token-saver">
            <Button
              variant="secondary"
              size="compact"
              leftIcon={<Zap className="w-3.5 h-3.5" />}
            >
              Token Saver
            </Button>
          </Link>
        }
      />

      {savedFeedback && (
        <div className="p-3 rounded-[6px] bg-emerald-950/20 border border-emerald-600/30 text-emerald-300 text-[12px] flex items-center gap-2 animate-in fade-in-50">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>✓ Tersimpan!</span>
        </div>
      )}

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <MetricCard
          label="Edge Latency"
          value="-42ms"
          subtext="Direct pipe reduction"
          icon={<Gauge className="w-4 h-4" />}
        />
        <MetricCard
          label="Token Compaction"
          value="34.8%"
          subtext="Average payload saved"
          icon={<Zap className="w-4 h-4" />}
        />
        <MetricCard
          label="Cache Hit Rate"
          value="18.4%"
          subtext="Saved upstream requests"
          icon={<Activity className="w-4 h-4" />}
        />
        <MetricCard
          label="Failover Latency"
          value="0ms"
          subtext="Circuit switch speed"
          icon={<ShieldCheck className="w-4 h-4" />}
        />
      </div>

      <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-4">
        <div>
          <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
            Acceleration Modules
          </h3>
          <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
            Toggle gateway optimization modules to reduce latency and save billable tokens.
          </p>
        </div>

        <div className="space-y-2.5">
          <div
            onClick={() => toggleFeature('fastLane')}
            className="p-3.5 rounded-[7px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex items-center justify-between cursor-pointer select-none"
          >
            <div className="min-w-0 pr-4">
              <span className="font-semibold text-[13px] text-[var(--text-primary)] block">
                Fast-Lane Streaming
              </span>
              <span className="text-[11.5px] text-[var(--text-muted)] block mt-0.5">
                Direct HTTP/2 pipe with minimal proxy overhead.
              </span>
            </div>
            <input
              type="checkbox"
              checked={config.fastLane}
              onChange={() => {}}
              className="w-4 h-4 accent-blue-600 rounded cursor-pointer shrink-0"
            />
          </div>

          <div
            onClick={() => toggleFeature('tokenSaver')}
            className="p-3.5 rounded-[7px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex items-center justify-between cursor-pointer select-none"
          >
            <div className="min-w-0 pr-4">
              <span className="font-semibold text-[13px] text-[var(--text-primary)] block">
                Token Saver Compaction
              </span>
              <span className="text-[11.5px] text-[var(--text-muted)] block mt-0.5">
                Strips redundant whitespace and system boilerplate before dispatch.
              </span>
            </div>
            <input
              type="checkbox"
              checked={config.tokenSaver}
              onChange={() => {}}
              className="w-4 h-4 accent-blue-600 rounded cursor-pointer shrink-0"
            />
          </div>

          <div
            onClick={() => toggleFeature('semanticCache')}
            className="p-3.5 rounded-[7px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex items-center justify-between cursor-pointer select-none"
          >
            <div className="min-w-0 pr-4">
              <span className="font-semibold text-[13px] text-[var(--text-primary)] block">
                Response Semantic Cache
              </span>
              <span className="text-[11.5px] text-[var(--text-muted)] block mt-0.5">
                Reuses completed model answers for identical requests.
              </span>
            </div>
            <input
              type="checkbox"
              checked={config.semanticCache}
              onChange={() => {}}
              className="w-4 h-4 accent-blue-600 rounded cursor-pointer shrink-0"
            />
          </div>

          <div
            onClick={() => toggleFeature('haFailover')}
            className="p-3.5 rounded-[7px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex items-center justify-between cursor-pointer select-none"
          >
            <div className="min-w-0 pr-4">
              <span className="font-semibold text-[13px] text-[var(--text-primary)] block">
                Instant HA Auto-Failover
              </span>
              <span className="text-[11.5px] text-[var(--text-muted)] block mt-0.5">
                Reroutes 429 and 500 error spikes to standby providers.
              </span>
            </div>
            <input
              type="checkbox"
              checked={config.haFailover}
              onChange={() => {}}
              className="w-4 h-4 accent-blue-600 rounded cursor-pointer shrink-0"
            />
          </div>
        </div>

        <div className="pt-2 flex items-center gap-3">
          <Link to="/token-saver">
            <Button variant="secondary" size="compact" rightIcon={<ArrowRight className="w-3.5 h-3.5" />}>
              Token Saver
            </Button>
          </Link>
          <Link to="/routing">
            <Button variant="ghost" size="compact">
              Routing
            </Button>
          </Link>
        </div>
      </div>
    </div>
  )
}
