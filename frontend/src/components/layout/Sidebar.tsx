import { useState, useEffect } from 'react'
import { NavLink } from 'react-router-dom'
import {
  Activity,
  Terminal,
  Server,
  GitFork,
  KeyRound,
  Globe,
  Key,
  FileText,
  Settings,
  ChevronLeft,
  ChevronRight,
  Sparkles,
  X,
  Zap,
  BookOpen,
  Wrench,
  Layers,
} from 'lucide-react'

interface NavItem {
  to: string
  label: string
  icon: React.ComponentType<{ className?: string }>
}

interface NavGroup {
  title: string
  items: NavItem[]
}

const NAV_GROUPS: NavGroup[] = [
  {
    title: 'CORE GATEWAY',
    items: [
      { to: '/overview', label: 'Overview', icon: Activity },
      { to: '/providers', label: 'Providers', icon: Server },
      { to: '/routing', label: 'Routing', icon: GitFork },
      { to: '/playground', label: 'Playground', icon: Sparkles },
      { to: '/api-keys', label: 'API Keys', icon: Key },
    ],
  },
  {
    title: 'ACCELERATION & AI',
    items: [
      { to: '/boost', label: 'Boost', icon: Zap },
      { to: '/learn', label: 'Learn', icon: BookOpen },
      { to: '/token-saver', label: 'Token Saver', icon: Zap },
      { to: '/tools', label: 'Tools & MCP', icon: Wrench },
    ],
  },
  {
    title: 'SECURITY & VAULT',
    items: [
      { to: '/vault', label: 'Vault', icon: KeyRound },
      { to: '/proxies', label: 'Proxies', icon: Globe },
      { to: '/credential-pools', label: 'Pools', icon: Layers },
    ],
  },
  {
    title: 'OBSERVABILITY',
    items: [
      { to: '/console', label: 'Console', icon: Terminal },
      { to: '/audit-logs', label: 'Audit Logs', icon: FileText },
      { to: '/settings', label: 'Settings', icon: Settings },
    ],
  },
]

export interface SidebarProps {
  isOpen: boolean
  onClose: () => void
  isCollapsed: boolean
  onToggleCollapse: () => void
}

export function Sidebar({ isOpen, onClose, isCollapsed, onToggleCollapse }: SidebarProps) {
  const [port, setPort] = useState('8080')

  useEffect(() => {
    if (window.location.port) {
      setPort(window.location.port)
    }
  }, [])

  return (
    <>
      {isOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/60 lg:hidden backdrop-blur-xs"
          onClick={onClose}
          aria-hidden="true"
        />
      )}

      <aside
        className={`fixed top-0 bottom-0 left-0 z-40 bg-[var(--bg-sidebar)] border-r border-[var(--border-subtle)] flex flex-col transition-all duration-200 lg:translate-x-0 ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        } ${isCollapsed ? 'w-[64px]' : 'w-[230px]'}`}
      >
        <div className="h-[52px] px-3 flex items-center justify-between border-b border-[var(--border-subtle)] shrink-0">
          <div className="flex items-center gap-2.5 overflow-hidden">
            <img
              src="/logo.svg"
              alt="EkaRouter"
              className="w-6 h-6 shrink-0 object-contain"
            />
            {!isCollapsed && (
              <div className="flex flex-col min-w-0">
                <span className="font-semibold text-[13.5px] tracking-tight text-[var(--text-primary)] leading-tight truncate">
                  EkaRouter
                </span>
                <span className="text-[10px] text-[var(--text-muted)] font-mono font-medium truncate">
                  AI Gateway
                </span>
              </div>
            )}
          </div>

          <div className="flex items-center">
            <button
              type="button"
              onClick={onToggleCollapse}
              title={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
              className="hidden lg:flex p-1 rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors cursor-pointer"
            >
              {isCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="p-1 rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] lg:hidden cursor-pointer"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        <nav className="flex-1 overflow-y-auto px-2 py-3 space-y-4">
          {NAV_GROUPS.map((group) => (
            <div key={group.title}>
              {!isCollapsed && (
                <h4 className="px-2 mb-1.5 text-[10px] font-bold tracking-wider text-[var(--text-muted)] uppercase">
                  {group.title}
                </h4>
              )}
              <ul className="space-y-0.5">
                {group.items.map((item) => {
                  const Icon = item.icon
                  return (
                    <li key={item.to}>
                      <NavLink
                        to={item.to}
                        onClick={onClose}
                        title={isCollapsed ? item.label : undefined}
                        className={({ isActive }) =>
                          `group relative flex items-center ${
                            isCollapsed ? 'justify-center px-0' : 'gap-2.5 px-2.5'
                          } py-1.5 text-[12px] font-medium rounded-[6px] transition-all select-none ${
                            isActive
                              ? 'bg-[var(--brand-subtle)] text-[var(--text-primary)] font-semibold border-l-2 border-[var(--brand-primary)]'
                              : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)]/70'
                          }`
                        }
                      >
                        {({ isActive }) => (
                          <>
                            <Icon
                              className={`w-4 h-4 shrink-0 transition-colors ${
                                isActive ? 'text-[var(--brand-text)]' : 'text-[var(--text-muted)] group-hover:text-[var(--text-primary)]'
                              }`}
                            />
                            {!isCollapsed && (
                              <span className="truncate">{item.label}</span>
                            )}
                          </>
                        )}
                      </NavLink>
                    </li>
                  )
                })}
              </ul>
            </div>
          ))}
        </nav>

        <div className="p-2 border-t border-[var(--border-subtle)]">
          <div
            className={`rounded-[7px] bg-[var(--bg-panel)]/70 border border-[var(--border-subtle)] ${
              isCollapsed ? 'p-2 flex justify-center' : 'p-2 flex items-center justify-between'
            }`}
          >
            <div className="flex items-center gap-2 overflow-hidden">
              <span className="relative flex h-2 w-2 shrink-0">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[var(--status-success)] opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-[var(--status-success)]" />
              </span>
              {!isCollapsed && (
                <div className="flex flex-col min-w-0">
                  <span className="text-[11px] font-medium text-[var(--text-primary)] leading-tight truncate">
                    Gateway v1.0.0
                  </span>
                  <span className="text-[10px] font-mono text-[var(--text-muted)] truncate">
                    Port {port}
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
      </aside>
    </>
  )
}
