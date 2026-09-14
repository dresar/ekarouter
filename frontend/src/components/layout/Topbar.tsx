import { useEffect, useState, useRef } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  Menu,
  Sun,
  Moon,
  LogOut,
  Bell,
  Search,
  ChevronDown,
  Settings as SettingsIcon,
  Shield,
} from 'lucide-react'
import { useAuth } from '../../context/AuthContext.tsx'
import { useTheme } from '../../context/ThemeContext.tsx'
import { api } from '../../api/client.ts'
import { SystemHealth } from '../../types/api.ts'

export interface TopbarProps {
  onToggleSidebar: () => void
}

export function Topbar({ onToggleSidebar }: TopbarProps) {
  const { user, logout } = useAuth()
  const { resolvedTheme, toggleTheme } = useTheme()
  const navigate = useNavigate()
  const [latency, setLatency] = useState<number | null>(32)
  const [isOffline, setIsOffline] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    async function checkHealth() {
      const start = performance.now()
      try {
        await api.get<SystemHealth>('/health', { timeoutMs: 4000 })
        const elapsed = Math.round(performance.now() - start)
        setLatency(Math.max(8, elapsed))
        setIsOffline(false)
      } catch {
        setIsOffline(true)
        setLatency(null)
      }
    }

    checkHealth()
    const timer = setInterval(checkHealth, 20000)
    return () => clearInterval(timer)
  }, [])

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setUserMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!searchQuery.trim()) return
    const q = searchQuery.toLowerCase().trim()
    if (q.includes('prov')) navigate('/providers')
    else if (q.includes('rout') || q.includes('combo')) navigate('/routing')
    else if (q.includes('key')) navigate('/api-keys')
    else if (q.includes('model')) navigate('/models')
    else if (q.includes('log') || q.includes('event')) navigate('/audit-logs')
    else if (q.includes('console') || q.includes('term')) navigate('/console')
    else if (q.includes('free')) navigate('/free-tiers')
    else if (q.includes('vault') || q.includes('cred')) navigate('/vault')
    else if (q.includes('usage') || q.includes('tele')) navigate('/usage')
    else if (q.includes('proxy')) navigate('/proxies')
    else navigate('/providers')
    setSearchQuery('')
  }

  return (
    <header className="h-[52px] bg-[var(--bg-surface)] border-b border-[var(--border-subtle)] px-4 flex items-center justify-between sticky top-0 z-30">
      <div className="flex items-center gap-3 flex-1 max-w-md">
        <button
          type="button"
          onClick={onToggleSidebar}
          aria-label="Toggle navigation"
          className="p-1.5 rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] lg:hidden"
        >
          <Menu className="w-4 h-4" />
        </button>

        <form onSubmit={handleSearchSubmit} className="relative w-full max-w-sm hidden sm:block">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)] pointer-events-none" />
          <input
            type="text"
            placeholder="Search anything..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-8 pr-12 py-1.5 text-[12px] rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)]"
          />
          <kbd className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[10px] font-mono text-[var(--text-muted)] bg-[var(--bg-card)] px-1.5 py-0.5 rounded border border-[var(--border-subtle)] pointer-events-none">
            Ctrl K
          </kbd>
        </form>
      </div>

      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 px-2.5 py-1 rounded-[6px] bg-[var(--bg-panel)]/50 border border-[var(--border-subtle)] text-[11px]">
          <span className="relative flex h-2 w-2">
            <span
              className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
                isOffline ? 'bg-[var(--status-danger)]' : 'bg-[var(--status-success)]'
              }`}
            />
            <span
              className={`relative inline-flex rounded-full h-2 w-2 ${
                isOffline ? 'bg-[var(--status-danger)]' : 'bg-[var(--status-success)]'
              }`}
            />
          </span>
          <span className="font-medium text-[var(--text-secondary)]">
            {isOffline ? 'Backend Offline' : 'Backend Online'}
          </span>
          {latency !== null && !isOffline && (
            <span className="font-mono text-[10px] text-[var(--text-muted)]">
              {latency}ms
            </span>
          )}
        </div>

        <button
          type="button"
          onClick={toggleTheme}
          title={`Switch to ${resolvedTheme === 'dark' ? 'light' : 'dark'} mode`}
          className="p-1.5 rounded-[6px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors"
        >
          {resolvedTheme === 'dark' ? (
            <Sun className="w-4 h-4" />
          ) : (
            <Moon className="w-4 h-4" />
          )}
        </button>

        <button
          type="button"
          title="Notifications"
          className="relative p-1.5 rounded-[6px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors"
        >
          <Bell className="w-4 h-4" />
          <span className="absolute top-1 right-1 w-1.5 h-1.5 rounded-full bg-[var(--brand-primary)]" />
        </button>

        <div className="relative" ref={menuRef}>
          <button
            type="button"
            onClick={() => setUserMenuOpen(!userMenuOpen)}
            className="flex items-center gap-2 p-1 pr-2 rounded-[6px] hover:bg-[var(--bg-panel)] transition-colors text-left"
          >
            <div className="w-6 h-6 rounded-full bg-blue-900/60 border border-blue-500/40 text-blue-300 flex items-center justify-center font-bold text-[11px]">
              {user?.username?.[0]?.toUpperCase() || 'A'}
            </div>
            <div className="hidden md:flex flex-col">
              <span className="text-[12px] font-medium text-[var(--text-primary)] leading-tight">
                {user?.username || 'admin'}
              </span>
              <span className="text-[10px] text-[var(--text-muted)] leading-tight">
                {user?.role === 'admin' ? 'Administrator' : 'Operator'}
              </span>
            </div>
            <ChevronDown className="w-3.5 h-3.5 text-[var(--text-muted)]" />
          </button>

          {userMenuOpen && (
            <div className="absolute right-0 mt-1.5 w-44 rounded-[7px] bg-[var(--bg-card)] border border-[var(--border-strong)] shadow-lg py-1 z-50 animate-in fade-in-50">
              <div className="px-3 py-1.5 border-b border-[var(--border-subtle)] text-[11px]">
                <p className="font-medium text-[var(--text-primary)] truncate">{user?.username}</p>
                <p className="text-[10px] text-[var(--text-muted)] font-mono">Role: {user?.role}</p>
              </div>
              <Link
                to="/settings"
                onClick={() => setUserMenuOpen(false)}
                className="flex items-center gap-2 px-3 py-1.5 text-[12px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors"
              >
                <SettingsIcon className="w-3.5 h-3.5" />
                <span>Settings</span>
              </Link>
              <Link
                to="/vault"
                onClick={() => setUserMenuOpen(false)}
                className="flex items-center gap-2 px-3 py-1.5 text-[12px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors"
              >
                <Shield className="w-3.5 h-3.5" />
                <span>Credential Vault</span>
              </Link>
              <div className="h-px bg-[var(--border-subtle)] my-1" />
              <button
                type="button"
                onClick={() => {
                  setUserMenuOpen(false)
                  logout()
                }}
                className="w-full flex items-center gap-2 px-3 py-1.5 text-[12px] text-[var(--status-danger)] hover:bg-[var(--status-danger)]/10 transition-colors"
              >
                <LogOut className="w-3.5 h-3.5" />
                <span>Sign out</span>
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
