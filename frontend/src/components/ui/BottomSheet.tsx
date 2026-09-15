import { useEffect, ReactNode } from 'react'
import { X } from 'lucide-react'

export interface BottomSheetProps {
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
  maxWidth = 'max-w-xl',
}: BottomSheetProps) {
  useEffect(() => {
    if (!isOpen) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      document.body.style.overflow = prevOverflow
    }
  }, [isOpen, onClose])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center">
      <div
        className="fixed inset-0 bg-black/70 backdrop-blur-xs transition-opacity animate-in fade-in duration-200"
        onClick={onClose}
      />

      <div
        className={`relative z-10 w-full ${maxWidth} max-h-[88vh] bg-[#14151a] border border-[#2d3039] rounded-t-[16px] sm:rounded-[14px] sm:mb-6 shadow-2xl flex flex-col overflow-hidden animate-in slide-in-from-bottom duration-250`}
      >
        <div className="w-10 h-1 bg-[#373a46] rounded-full mx-auto mt-2.5 mb-1 shrink-0" />

        <div className="flex items-center justify-between px-5 py-3 border-b border-[#252832]">
          <div className="flex items-center gap-2.5">
            <div className="flex items-center gap-1.5 mr-1">
              <span className="w-3 h-3 rounded-full bg-[#ff5f56]" />
              <span className="w-3 h-3 rounded-full bg-[#ffbd2e]" />
              <span className="w-3 h-3 rounded-full bg-[#27c93f]" />
            </div>
            <div>
              <h3 className="text-[14.5px] font-semibold text-[#f0f2f7] leading-tight">
                {title}
              </h3>
              {description && (
                <p className="text-[11.5px] text-[#858999] mt-0.5 leading-snug">
                  {description}
                </p>
              )}
            </div>
          </div>

          <button
            type="button"
            onClick={onClose}
            className="p-1.5 text-[#858999] hover:text-[#f0f2f7] hover:bg-[#20222a] rounded-[6px] transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="p-5 overflow-y-auto flex-1 text-[#d8dae5]">
          {children}
        </div>
      </div>
    </div>
  )
}
