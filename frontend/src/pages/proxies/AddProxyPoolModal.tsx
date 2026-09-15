import { useState, useEffect, FormEvent } from 'react'
import { AlertCircle } from 'lucide-react'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'
import { ProxyProfile } from '../../types/api.ts'

interface AddProxyPoolModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
  initialData?: ProxyProfile | null
}

export function AddProxyPoolModal({
  isOpen,
  onClose,
  onSuccess,
  initialData,
}: AddProxyPoolModalProps) {
  const [name, setName] = useState('')
  const [proxyUrl, setProxyUrl] = useState('')
  const [noProxy, setNoProxy] = useState('')
  const [active, setActive] = useState(true)
  const [strictProxy, setStrictProxy] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (initialData) {
      setName(initialData.name || '')
      setProxyUrl(
        initialData.proxy_url ||
          (initialData.host ? `${initialData.scheme || 'http'}://${initialData.host}:${initialData.port}` : '')
      )
      setNoProxy(initialData.no_proxy || '')
      setActive(initialData.enabled ?? true)
      setStrictProxy(initialData.strict_proxy ?? false)
    } else {
      setName('')
      setProxyUrl('')
      setNoProxy('')
      setActive(true)
      setStrictProxy(false)
    }
    setError(null)
  }, [initialData, isOpen])

  if (!isOpen) return null

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !proxyUrl.trim()) {
      setError('Name and Proxy URL are required')
      return
    }

    setIsSubmitting(true)
    setError(null)

    try {
      if (initialData?.id) {
        await api.put(`/api/proxies/${initialData.id}`, {
          name: name.trim(),
          proxy_url: proxyUrl.trim(),
          no_proxy: noProxy.trim(),
          enabled: active,
          strict_proxy: strictProxy,
        })
      } else {
        await api.post('/api/proxies', {
          name: name.trim(),
          proxy_url: proxyUrl.trim(),
          no_proxy: noProxy.trim(),
          enabled: active,
          strict_proxy: strictProxy,
        })
      }
      onSuccess()
      onClose()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to save proxy pool'
      setError(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
      <div
        className="w-full max-w-[440px] bg-[#1a1c23] border border-[#2e323e] rounded-[12px] shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150"
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
            {initialData ? 'Edit Proxy Pool' : 'Add Proxy Pool'}
          </span>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          {error && (
            <div className="p-2.5 rounded-[6px] bg-[var(--status-danger)]/15 border border-[var(--status-danger)]/30 flex items-center gap-2 text-[12px] text-[var(--status-danger)]">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div className="space-y-1">
            <label className="block text-[12px] font-semibold text-[#e5e7eb]">
              Name
            </label>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Office Proxy"
              className="w-full px-3 py-2 text-[12.5px] rounded-[6px] bg-[#232630] border border-[#333745] text-[#f3f4f6] placeholder-[#6b7280] focus:outline-none focus:border-[#4b5563]"
            />
          </div>

          <div className="space-y-1">
            <label className="block text-[12px] font-semibold text-[#e5e7eb]">
              Proxy URL
            </label>
            <input
              type="text"
              required
              value={proxyUrl}
              onChange={(e) => setProxyUrl(e.target.value)}
              placeholder="http://127.0.0.1:7897"
              className="w-full px-3 py-2 text-[12.5px] font-mono rounded-[6px] bg-[#232630] border border-[#333745] text-[#f3f4f6] placeholder-[#6b7280] focus:outline-none focus:border-[#4b5563]"
            />
          </div>

          <div className="space-y-1">
            <label className="block text-[12px] font-semibold text-[#e5e7eb]">
              No Proxy
            </label>
            <input
              type="text"
              value={noProxy}
              onChange={(e) => setNoProxy(e.target.value)}
              placeholder="localhost,127.0.0.1,.internal"
              className="w-full px-3 py-2 text-[12.5px] font-mono rounded-[6px] bg-[#232630] border border-[#333745] text-[#f3f4f6] placeholder-[#6b7280] focus:outline-none focus:border-[#4b5563]"
            />
            <p className="text-[11px] text-[#9ca3af]">
              Comma-separated hosts/domains to bypass proxy
            </p>
          </div>

          <div className="pt-2 border-t border-[#262934] space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-[12.5px] font-semibold text-[#f3f4f6]">Active</div>
                <div className="text-[11px] text-[#9ca3af]">
                  Inactive pools are ignored by runtime resolution.
                </div>
              </div>
              <button
                type="button"
                onClick={() => setActive(!active)}
                className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full transition-colors duration-200 ease-in-out focus:outline-none ${
                  active ? 'bg-[#ff6940]' : 'bg-[#374151]'
                }`}
              >
                <span
                  className={`pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out mt-0.5 ${
                    active ? 'translate-x-4 ml-0.5' : 'translate-x-0.5'
                  }`}
                />
              </button>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <div className="text-[12.5px] font-semibold text-[#f3f4f6]">Strict Proxy</div>
                <div className="text-[11px] text-[#9ca3af]">
                  Fail request if proxy is unreachable instead of falling back to direct.
                </div>
              </div>
              <button
                type="button"
                onClick={() => setStrictProxy(!strictProxy)}
                className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full transition-colors duration-200 ease-in-out focus:outline-none ${
                  strictProxy ? 'bg-[#ff6940]' : 'bg-[#374151]'
                }`}
              >
                <span
                  className={`pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out mt-0.5 ${
                    strictProxy ? 'translate-x-4 ml-0.5' : 'translate-x-0.5'
                  }`}
                />
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
              isLoading={isSubmitting}
              className="bg-[#2d313d] hover:bg-[#383d4c] text-[#f3f4f6] px-5 font-semibold"
            >
              Save
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
