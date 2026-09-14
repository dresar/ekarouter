import { useEffect, ReactNode } from 'react'
import { X } from 'lucide-react'

interface BottomSheetProps {
  isOpen: boolean
  onClose: () => void
  title: string
  description?: string
  children: ReactNode
  maxWidth?: string
}

export function BottomSheet({
  isOpen,
  onClose,
  title,
  description,
  children,
  maxWidth = 'max-w-2xl',
}: BottomSheetProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      <div
        className="fixed inset-0 bg-black/65 backdrop-blur-[2px] transition-opacity duration-200"
        onClick={onClose}
      />

      <div className="fixed inset-x-0 bottom-0 z-50 flex justify-center pointer-events-none">
        <div
          className={`pointer-events-auto w-full ${maxWidth} bg-[var(--bg-card)] border-t border-x border-[var(--border-strong)] rounded-t-[16px] shadow-2xl flex flex-col max-h-[85vh] transform transition-transform duration-300 ease-out animate-in slide-in-from-bottom`}
        >
          <div className="pt-2.5 pb-1 flex justify-center shrink-0">
            <div className="w-10 h-1 rounded-full bg-[var(--border-strong)]" />
          </div>

          <div className="flex items-center justify-between px-5 py-3 border-b border-[var(--border-subtle)] shrink-0">
            <div>
              <h3 className="text-[15px] font-semibold text-[var(--text-primary)]">
                {title}
              </h3>
              {description && (
                <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                  {description}
                </p>
              )}
            </div>
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 rounded-[6px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          <div className="overflow-y-auto px-5 py-4 space-y-4">
            {children}
          </div>
        </div>
      </div>
    </div>
  )
}
