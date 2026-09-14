import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { X, Zap, ArrowRight } from 'lucide-react'
import { Button } from '../ui/Button.tsx'

export interface BoostModalProps {
  isOpen: boolean
  onClose: () => void
}

interface BoostState {
  fastLane: boolean
  tokenSaver: boolean
  semanticCache: boolean
  haFailover: boolean
}

const STORAGE_KEY = 'ekarouter_boost_config'

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

export function BoostModal({ isOpen, onClose }: BoostModalProps) {
  const navigate = useNavigate()
  const [boostConfig, setBoostConfig] = useState<BoostState>(loadBoostState)

  useEffect(() => {
    if (!isOpen) return

    const originalBodyOverflow = document.body.style.overflow
    const originalHtmlOverflow = document.documentElement.style.overflow
    const originalBodyOverscroll = document.body.style.overscrollBehavior
    const originalHtmlOverscroll = document.documentElement.style.overscrollBehavior
    const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth

    document.body.style.overflow = 'hidden'
    document.documentElement.style.overflow = 'hidden'
    document.body.style.overscrollBehavior = 'none'
    document.documentElement.style.overscrollBehavior = 'none'

    if (scrollbarWidth > 0) {
      document.body.style.paddingRight = `${scrollbarWidth}px`
    }

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => {
      document.body.style.overflow = originalBodyOverflow
      document.documentElement.style.overflow = originalHtmlOverflow
      document.body.style.overscrollBehavior = originalBodyOverscroll
      document.documentElement.style.overscrollBehavior = originalHtmlOverscroll
      document.body.style.paddingRight = ''
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [isOpen, onClose])

  const toggleFeature = (key: keyof BoostState) => {
    const updated = { ...boostConfig, [key]: !boostConfig[key] }
    setBoostConfig(updated)
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated))
    } catch {}
  }

  if (!isOpen) return null

  const navigateTo = (path: string) => {
    onClose()
    navigate(path)
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      onWheel={(e) => e.stopPropagation()}
    >
      <div
        className="fixed inset-0 bg-black/60 backdrop-blur-xs transition-opacity"
        onClick={onClose}
        onWheel={(e) => {
          e.preventDefault()
          e.stopPropagation()
        }}
        onTouchMove={(e) => {
          e.preventDefault()
          e.stopPropagation()
        }}
        aria-hidden="true"
      />

      <div
        className="relative w-full max-w-[480px] rounded-[10px] bg-[var(--bg-card)] border border-[var(--border-strong)] shadow-2xl overflow-hidden z-10 animate-in zoom-in-95 duration-150"
        onWheel={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-5 py-4 border-b border-[var(--border-subtle)] bg-[var(--bg-card)]">
          <div className="flex items-center gap-2">
            <div className="p-1 rounded-[5px] bg-amber-500/10 text-amber-400 border border-amber-500/20">
              <Zap className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
                Gateway Boost
              </h3>
              <p className="text-[11px] text-[var(--text-muted)]">
                Accelerate throughput and optimize token usage.
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close boost modal"
            className="p-1 rounded-[6px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="p-5 space-y-3">
          <div className="grid grid-cols-3 gap-2 text-center">
            <div className="p-2.5 rounded-[7px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)]">
              <span className="text-[10px] uppercase font-mono text-[var(--text-muted)] block">
                Latency
              </span>
              <span className="text-[14px] font-bold text-emerald-400 font-mono">-42ms</span>
            </div>
            <div className="p-2.5 rounded-[7px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)]">
              <span className="text-[10px] uppercase font-mono text-[var(--text-muted)] block">
                Token Savings
              </span>
              <span className="text-[14px] font-bold text-blue-400 font-mono">34.8%</span>
            </div>
            <div className="p-2.5 rounded-[7px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)]">
              <span className="text-[10px] uppercase font-mono text-[var(--text-muted)] block">
                Failover
              </span>
              <span className="text-[14px] font-bold text-amber-400 font-mono">0ms</span>
            </div>
          </div>

          <div className="space-y-2 pt-1">
            <div
              onClick={() => toggleFeature('fastLane')}
              className="p-3 rounded-[7px] bg-[var(--bg-panel)]/40 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex items-center justify-between cursor-pointer select-none"
            >
              <div className="min-w-0 pr-3">
                <span className="font-semibold text-[12.5px] text-[var(--text-primary)] block">
                  Fast-Lane Streaming
                </span>
                <span className="text-[11px] text-[var(--text-muted)] block">
                  Direct HTTP/2 pipe with minimal proxy overhead.
                </span>
              </div>
              <input
                type="checkbox"
                checked={boostConfig.fastLane}
                onChange={() => {}}
                className="w-4 h-4 accent-blue-600 rounded cursor-pointer shrink-0"
              />
            </div>

            <div
              onClick={() => toggleFeature('tokenSaver')}
              className="p-3 rounded-[7px] bg-[var(--bg-panel)]/40 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex items-center justify-between cursor-pointer select-none"
            >
              <div className="min-w-0 pr-3">
                <span className="font-semibold text-[12.5px] text-[var(--text-primary)] block">
                  Token Saver Compaction
                </span>
                <span className="text-[11px] text-[var(--text-muted)] block">
                  Strips redundant whitespace and system boilerplate before dispatch.
                </span>
              </div>
              <input
                type="checkbox"
                checked={boostConfig.tokenSaver}
                onChange={() => {}}
                className="w-4 h-4 accent-blue-600 rounded cursor-pointer shrink-0"
              />
            </div>

            <div
              onClick={() => toggleFeature('semanticCache')}
              className="p-3 rounded-[7px] bg-[var(--bg-panel)]/40 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex items-center justify-between cursor-pointer select-none"
            >
              <div className="min-w-0 pr-3">
                <span className="font-semibold text-[12.5px] text-[var(--text-primary)] block">
                  Response Semantic Cache
                </span>
                <span className="text-[11px] text-[var(--text-muted)] block">
                  Reuses completed model answers for identical requests.
                </span>
              </div>
              <input
                type="checkbox"
                checked={boostConfig.semanticCache}
                onChange={() => {}}
                className="w-4 h-4 accent-blue-600 rounded cursor-pointer shrink-0"
              />
            </div>

            <div
              onClick={() => toggleFeature('haFailover')}
              className="p-3 rounded-[7px] bg-[var(--bg-panel)]/40 border border-[var(--border-subtle)] hover:border-[var(--brand-primary)]/40 transition-colors flex items-center justify-between cursor-pointer select-none"
            >
              <div className="min-w-0 pr-3">
                <span className="font-semibold text-[12.5px] text-[var(--text-primary)] block">
                  Instant HA Auto-Failover
                </span>
                <span className="text-[11px] text-[var(--text-muted)] block">
                  Reroutes 429 and 500 error spikes to standby providers.
                </span>
              </div>
              <input
                type="checkbox"
                checked={boostConfig.haFailover}
                onChange={() => {}}
                className="w-4 h-4 accent-blue-600 rounded cursor-pointer shrink-0"
              />
            </div>
          </div>
        </div>

        <div className="px-5 py-3 border-t border-[var(--border-subtle)] bg-[var(--bg-panel)]/40 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => navigateTo('/boost')}
              className="text-[11.5px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 cursor-pointer font-medium"
            >
              <span>Full Dashboard</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </button>
            <span className="text-[var(--text-muted)] text-[11px]">•</span>
            <button
              type="button"
              onClick={() => navigateTo('/token-saver')}
              className="text-[11.5px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:underline inline-flex items-center gap-1 cursor-pointer"
            >
              <span>Token Saver</span>
            </button>
          </div>
          <Button
            variant="primary"
            size="compact"
            onClick={onClose}
          >
            Done
          </Button>
        </div>
      </div>
    </div>
  )
}
