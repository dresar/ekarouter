import { useState, FormEvent } from 'react'
import { AlertCircle, Upload, CheckCircle2 } from 'lucide-react'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

interface BatchImportModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
}

export function BatchImportModal({
  isOpen,
  onClose,
  onSuccess,
}: BatchImportModalProps) {
  const [rawText, setRawText] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<{ imported: number; failed: number } | null>(null)

  if (!isOpen) return null

  const validLines = rawText
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0 && !l.startsWith('#'))

  const handleImport = async (e: FormEvent) => {
    e.preventDefault()
    if (validLines.length === 0) {
      setError('Please paste at least one valid proxy URL')
      return
    }

    setIsSubmitting(true)
    setError(null)
    setResult(null)

    try {
      const res = await api.post<{ imported: number; failed: number; status: string }>(
        '/api/proxies/batch',
        { raw_text: rawText }
      )
      setResult({ imported: res.imported, failed: res.failed })
      setTimeout(() => {
        onSuccess()
        onClose()
      }, 900)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Batch import failed'
      setError(msg)
    } finally {
      setIsSubmitting(false)
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
            Batch Import Proxies
          </span>
        </div>

        <form onSubmit={handleImport} className="p-5 space-y-4">
          {error && (
            <div className="p-2.5 rounded-[6px] bg-[var(--status-danger)]/15 border border-[var(--status-danger)]/30 flex items-center gap-2 text-[12px] text-[var(--status-danger)]">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {result && (
            <div className="p-2.5 rounded-[6px] bg-[var(--status-success)]/15 border border-[var(--status-success)]/30 flex items-center gap-2 text-[12px] text-[var(--status-success)]">
              <CheckCircle2 className="w-4 h-4 shrink-0" />
              <span>
                Successfully imported {result.imported} proxies ({result.failed} skipped/failed).
              </span>
            </div>
          )}

          <div className="space-y-1">
            <div className="flex items-center justify-between">
              <label className="block text-[12px] font-semibold text-[#e5e7eb]">
                Proxy List (One per line)
              </label>
              <span className="text-[11px] font-mono text-[#9ca3af]">
                {validLines.length} detected
              </span>
            </div>
            <textarea
              rows={8}
              required
              value={rawText}
              onChange={(e) => setRawText(e.target.value)}
              placeholder={`http://127.0.0.1:7897\nhttp://username:password@10.0.0.50:8080\nsocks5://node1.proxy.net:1080\nhttps://custom-relay.workers.dev`}
              className="w-full p-3 font-mono text-[12px] rounded-[6px] bg-[#232630] border border-[#333745] text-[#f3f4f6] placeholder-[#6b7280] focus:outline-none focus:border-[#4b5563] resize-y leading-relaxed"
            />
            <p className="text-[11px] text-[#9ca3af]">
              Supports HTTP, HTTPS, SOCKS5 proxies with optional embedded user:password credentials.
            </p>
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
              isLoading={isSubmitting}
              leftIcon={<Upload className="w-3.5 h-3.5" />}
              className="bg-[#2d313d] hover:bg-[#383d4c] text-[#f3f4f6] px-5 font-semibold"
            >
              Import {validLines.length > 0 ? `(${validLines.length})` : ''}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
