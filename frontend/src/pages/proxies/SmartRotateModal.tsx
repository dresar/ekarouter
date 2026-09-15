import { useState, useEffect, FormEvent } from 'react'
import { AlertCircle, RotateCw, Gauge, Zap, CheckCircle2 } from 'lucide-react'
import { Button } from '../../components/ui/Button.tsx'
import { ProviderSelect, ProviderOption } from '../../components/ui/ProviderSelect.tsx'
import { api } from '../../api/client.ts'

interface SmartRotateModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
}

export function SmartRotateModal({
  isOpen,
  onClose,
  onSuccess,
}: SmartRotateModalProps) {
  const [providerId, setProviderId] = useState('')
  const [providers, setProviders] = useState<ProviderOption[]>([])
  const [strategy, setStrategy] = useState<'lowest_latency' | 'round_robin'>('lowest_latency')
  const [isRotating, setIsRotating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<{
    accounts_bound: number
    proxies_tested: number
    best_latency_ms: number
    strategy: string
  } | null>(null)

  useEffect(() => {
    if (isOpen) {
      api.get<Array<{ id: string; name: string; kind?: string }>>('/api/providers')
        .then((res) => {
          if (Array.isArray(res)) {
            setProviders(res.map((p) => ({ id: p.id, name: p.name, kind: p.kind })))
          }
        })
        .catch(() => {})
    }
  }, [isOpen])

  if (!isOpen) return null

  const handleSmartRotate = async (e: FormEvent) => {
    e.preventDefault()
    if (!providerId) {
      setError('Please select a target AI provider')
      return
    }

    setIsRotating(true)
    setError(null)
    setResult(null)

    try {
      const res = await api.post<{
        accounts_bound: number
        proxies_tested: number
        best_latency_ms: number
        strategy: string
        status: string
      }>('/api/proxies/smart-rotate', {
        provider_id: providerId,
        strategy,
      })

      setResult({
        accounts_bound: res.accounts_bound,
        proxies_tested: res.proxies_tested,
        best_latency_ms: res.best_latency_ms,
        strategy: res.strategy,
      })

      setTimeout(() => {
        onSuccess()
      }, 1200)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Smart proxy rotation failed'
      setError(msg)
    } finally {
      setIsRotating(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
      <div
        className="w-full max-w-[480px] bg-[#1a1c23] border border-[#2e323e] rounded-[12px] shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-2 px-4 py-3 bg-[#15171d] border-b border-[#262934] select-none">
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              onClick={onClose}
              className="w-3 h-3 rounded-full bg-[#ff5f56] hover:brightness-110 transition-all border border-[#e0443e]/40"
              title="Close"
            />
            <span className="w-3 h-3 rounded-full bg-[#ffbd2e] border border-[#dea123]/40" />
            <span className="w-3 h-3 rounded-full bg-[#27c93f] border border-[#1aab29]/40" />
          </div>
          <span className="text-[13.5px] font-semibold text-[#f3f4f6] ml-2">
            Smart Proxy Rotation & Latency Optimizer
          </span>
        </div>

        <form onSubmit={handleSmartRotate} className="p-5 space-y-4">
          {error && (
            <div className="p-2.5 rounded-[6px] bg-[var(--status-danger)]/15 border border-[var(--status-danger)]/30 flex items-center gap-2 text-[12px] text-[var(--status-danger)]">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {result && (
            <div className="p-3 rounded-[8px] bg-[var(--status-success)]/15 border border-[var(--status-success)]/30 space-y-1.5 text-[12px] text-[#f3f4f6]">
              <div className="flex items-center gap-1.5 font-bold text-[var(--status-success)]">
                <CheckCircle2 className="w-4 h-4" />
                <span>Smart Optimization Completed</span>
              </div>
              <div className="text-[11.5px] text-[#d1d5db]">
                Bound {result.accounts_bound} accounts across {result.proxies_tested} tested proxies.
                {result.best_latency_ms > 0 && (
                  <span className="ml-1 text-[var(--status-success)] font-semibold">
                    (Top latency: {result.best_latency_ms}ms)
                  </span>
                )}
              </div>
            </div>
          )}

          <div className="space-y-1.5">
            <label className="block text-[12px] font-semibold text-[#e5e7eb]">
              Target Provider *
            </label>
            <ProviderSelect
              value={providerId}
              onChange={setProviderId}
              providers={providers}
              placeholder="Select AI Provider to optimize..."
            />
            <p className="text-[11px] text-[#9ca3af]">
              Select which provider's account keys will receive optimized proxy routing.
            </p>
          </div>

          <div className="space-y-2">
            <label className="block text-[12px] font-semibold text-[#e5e7eb]">
              Optimization Strategy
            </label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setStrategy('lowest_latency')}
                className={`p-3 rounded-[8px] border text-left transition-all cursor-pointer ${
                  strategy === 'lowest_latency'
                    ? 'bg-[#262a36] border-[#ff6940] text-[#f3f4f6]'
                    : 'bg-[#1f222b] border-[#333745] text-[#9ca3af] hover:text-[#e5e7eb]'
                }`}
              >
                <div className="flex items-center gap-1.5 font-semibold text-[12px] mb-1">
                  <Gauge className="w-3.5 h-3.5 text-[#ff6940]" />
                  <span>Lowest Latency</span>
                </div>
                <div className="text-[10.5px] text-[#9ca3af] leading-tight">
                  Benchmarks all proxies and assigns the fastest tunnels.
                </div>
              </button>

              <button
                type="button"
                onClick={() => setStrategy('round_robin')}
                className={`p-3 rounded-[8px] border text-left transition-all cursor-pointer ${
                  strategy === 'round_robin'
                    ? 'bg-[#262a36] border-[#ff6940] text-[#f3f4f6]'
                    : 'bg-[#1f222b] border-[#333745] text-[#9ca3af] hover:text-[#e5e7eb]'
                }`}
              >
                <div className="flex items-center gap-1.5 font-semibold text-[12px] mb-1">
                  <RotateCw className="w-3.5 h-3.5 text-[#38bdf8]" />
                  <span>Round-Robin</span>
                </div>
                <div className="text-[10.5px] text-[#9ca3af] leading-tight">
                  Evenly distributes all active proxy IPs across keys.
                </div>
              </button>
            </div>
          </div>

          <div className="pt-3 flex items-center justify-end gap-2 border-t border-[#262934]">
            <Button
              type="button"
              variant="ghost"
              size="compact"
              onClick={onClose}
              className="text-[#9ca3af] hover:text-[#f3f4f6]"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isRotating}
              leftIcon={<Zap className="w-3.5 h-3.5" />}
              className="bg-[#2d313d] hover:bg-[#383d4c] text-[#f3f4f6] px-5 font-semibold"
            >
              Run Smart Analysis & Rotate
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
