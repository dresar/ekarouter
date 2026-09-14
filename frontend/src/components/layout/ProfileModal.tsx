import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { X, ShieldCheck, KeyRound, Settings, LogOut, User } from 'lucide-react'
import { useAuth } from '../../context/AuthContext.tsx'
import { Button } from '../ui/Button.tsx'

export interface ProfileModalProps {
  isOpen: boolean
  onClose: () => void
}

export function ProfileModal({ isOpen, onClose }: ProfileModalProps) {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

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

  if (!isOpen) return null

  const handleLogout = () => {
    onClose()
    logout()
  }

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
        className="relative w-full max-w-[420px] rounded-[10px] bg-[var(--bg-card)] border border-[var(--border-strong)] shadow-2xl overflow-hidden z-10 animate-in zoom-in-95 duration-150"
        onWheel={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-5 py-4 border-b border-[var(--border-subtle)]">
          <div className="flex items-center gap-2">
            <User className="w-4 h-4 text-[var(--brand-text)]" />
            <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
              Operator Profile
            </h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close profile modal"
            className="p-1 rounded-[6px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="p-5 space-y-4">
          <div className="flex items-center gap-3.5 p-3 rounded-[8px] bg-[var(--bg-panel)]/60 border border-[var(--border-subtle)]">
            <div className="w-11 h-11 rounded-full bg-blue-900/40 border border-blue-500/30 text-blue-300 flex items-center justify-center font-bold text-[15px] shrink-0">
              {user?.username?.[0]?.toUpperCase() || 'A'}
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center justify-between gap-2">
                <span className="font-semibold text-[14px] text-[var(--text-primary)] truncate">
                  {user?.username || 'admin'}
                </span>
                <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded-[4px] bg-[var(--brand-subtle)] text-[var(--brand-text)] border border-[var(--brand-primary)]/20">
                  {user?.role || 'admin'}
                </span>
              </div>
              <p className="text-[11.5px] text-[var(--text-muted)] mt-0.5 truncate">
                Active gateway administrative session.
              </p>
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between px-3 py-2 rounded-[6px] bg-[var(--bg-panel)]/30 border border-[var(--border-subtle)] text-[11.5px]">
              <div className="flex items-center gap-2 text-[var(--text-secondary)]">
                <ShieldCheck className="w-3.5 h-3.5 text-emerald-400" />
                <span>Vault Security</span>
              </div>
              <span className="font-mono text-[10.5px] text-emerald-400">AES-256-GCM</span>
            </div>

            <div className="flex items-center justify-between px-3 py-2 rounded-[6px] bg-[var(--bg-panel)]/30 border border-[var(--border-subtle)] text-[11.5px]">
              <div className="flex items-center gap-2 text-[var(--text-secondary)]">
                <KeyRound className="w-3.5 h-3.5 text-blue-400" />
                <span>Session Status</span>
              </div>
              <span className="font-mono text-[10.5px] text-[var(--text-primary)]">Authenticated</span>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-2 pt-1">
            <Button
              variant="secondary"
              size="compact"
              onClick={() => navigateTo('/settings')}
              leftIcon={<Settings className="w-3.5 h-3.5" />}
              className="w-full justify-center"
            >
              Settings
            </Button>
            <Button
              variant="secondary"
              size="compact"
              onClick={() => navigateTo('/vault')}
              leftIcon={<KeyRound className="w-3.5 h-3.5" />}
              className="w-full justify-center"
            >
              Vault
            </Button>
          </div>
        </div>

        <div className="px-5 py-3 border-t border-[var(--border-subtle)] bg-[var(--bg-panel)]/40 flex items-center justify-between">
          <span className="text-[11px] text-[var(--text-muted)]">
            EkaRouter Gateway v1.0.0
          </span>
          <button
            type="button"
            onClick={handleLogout}
            className="inline-flex items-center gap-1.5 px-2.5 py-1 text-[11.5px] font-medium text-[var(--status-danger)] hover:bg-[var(--status-danger)]/10 rounded-[5px] transition-colors cursor-pointer"
          >
            <LogOut className="w-3.5 h-3.5" />
            <span>Sign Out</span>
          </button>
        </div>
      </div>
    </div>
  )
}
