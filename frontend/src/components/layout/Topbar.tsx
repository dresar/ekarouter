import { useState } from 'react'
import {
  Menu,
  Sun,
  Moon,
  Zap,
} from 'lucide-react'
import { useAuth } from '../../context/AuthContext.tsx'
import { useTheme } from '../../context/ThemeContext.tsx'
import { ProfileModal } from './ProfileModal.tsx'
import { BoostModal } from './BoostModal.tsx'

export interface TopbarProps {
  onToggleSidebar: () => void
}

export function Topbar({ onToggleSidebar }: TopbarProps) {
  const { user } = useAuth()
  const { resolvedTheme, toggleTheme } = useTheme()
  const [profileModalOpen, setProfileModalOpen] = useState(false)
  const [boostModalOpen, setBoostModalOpen] = useState(false)

  return (
    <>
      <header className="h-[52px] bg-[var(--bg-surface)] border-b border-[var(--border-subtle)] px-4 flex items-center justify-between sticky top-0 z-30">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={onToggleSidebar}
            aria-label="Toggle navigation"
            className="p-1.5 rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] lg:hidden cursor-pointer"
          >
            <Menu className="w-4 h-4" />
          </button>
        </div>

        <div className="flex items-center gap-2.5">
          <button
            type="button"
            onClick={() => setBoostModalOpen(true)}
            title="Gateway Boost Acceleration"
            className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-[6px] bg-amber-500/10 hover:bg-amber-500/15 border border-amber-500/25 text-amber-400 text-[11.5px] font-semibold transition-colors cursor-pointer select-none"
          >
            <Zap className="w-3.5 h-3.5 fill-amber-400" />
            <span>Boost</span>
          </button>

          <button
            type="button"
            onClick={toggleTheme}
            title={`Switch to ${resolvedTheme === 'dark' ? 'light' : 'dark'} mode`}
            className="p-1.5 rounded-[6px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors cursor-pointer"
          >
            {resolvedTheme === 'dark' ? (
              <Sun className="w-4 h-4" />
            ) : (
              <Moon className="w-4 h-4" />
            )}
          </button>

          <button
            type="button"
            onClick={() => setProfileModalOpen(true)}
            title="Operator Profile"
            className="flex items-center gap-2 p-1 pr-2 rounded-[6px] hover:bg-[var(--bg-panel)] transition-colors cursor-pointer text-left select-none border border-transparent hover:border-[var(--border-subtle)]"
          >
            <div className="w-6 h-6 rounded-full bg-blue-900/60 border border-blue-500/40 text-blue-300 flex items-center justify-center font-bold text-[11px] shrink-0">
              {user?.username?.[0]?.toUpperCase() || 'A'}
            </div>
            <div className="hidden sm:flex flex-col">
              <span className="text-[12px] font-medium text-[var(--text-primary)] leading-tight truncate max-w-[100px]">
                {user?.username || 'admin'}
              </span>
            </div>
          </button>
        </div>
      </header>

      <ProfileModal
        isOpen={profileModalOpen}
        onClose={() => setProfileModalOpen(false)}
      />

      <BoostModal
        isOpen={boostModalOpen}
        onClose={() => setBoostModalOpen(false)}
      />
    </>
  )
}
