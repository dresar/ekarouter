import { useState } from 'react'
import { Copy, Check } from 'lucide-react'

export interface SecretViewerProps {
  value: string
  copyValue?: string
  className?: string
}

export function SecretViewer({ value, copyValue, className = '' }: SecretViewerProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    const textToCopy = copyValue || value
    if (!textToCopy) return

    try {
      await navigator.clipboard.writeText(textToCopy)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {}
  }

  return (
    <div className={`inline-flex items-center gap-1.5 font-mono text-[12px] text-[var(--text-secondary)] bg-[var(--bg-input)] px-2 py-1 rounded-[4px] border border-[var(--border-strong)] ${className}`}>
      <span className="select-all truncate max-w-[200px]">{value}</span>
      <button
        type="button"
        onClick={handleCopy}
        title="Copy to clipboard"
        className="p-0.5 text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors rounded hover:bg-[var(--border-strong)] focus:outline-none"
      >
        {copied ? (
          <Check className="w-3.5 h-3.5 text-[var(--status-success)]" />
        ) : (
          <Copy className="w-3.5 h-3.5" />
        )}
      </button>
    </div>
  )
}
