import { useState, useEffect } from 'react'
import { NavLink } from 'react-router-dom'
import {
  Activity,
  Terminal,
  Server,
  GitFork,
  Cpu,
  Zap,
  KeyRound,
  Wrench,
  Globe,
  Sparkles,
  Layers,
  Key,
  BarChart3,
  FileText,
  Gauge,
  Settings,
  Database,
  BookOpen,
  ChevronLeft,
  ChevronRight,
  X,
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
    title: 'OVERVIEW',
    items: [
      { to: '/overview', label: 'Command Center', icon: Activity },
      { to: '/console', label: 'Live Console', icon: Terminal },
    ],
  },
  {
    title: 'AI GATEWAY',
    items: [
      { to: '/providers', label: 'AI Providers', icon: Server },
      { to: '/routing', label: 'Routing Combos', icon: GitFork },
      { to: '/models', label: 'Model Catalog', icon: Cpu },
      { to: '/token-saver', label: 'Token Saver', icon: Zap },
    ],
  },
  {
    title: 'INFRASTRUCTURE',
    items: [
      { to: '/vault', label: 'Credential Vault', icon: KeyRound },
      { to: '/tools', label: 'Tools & Templates', icon: Wrench },
      { to: '/proxies', label: 'Outbound Proxies', icon: Globe },
      { to: '/free-tiers', label: 'Free Tiers', icon: Sparkles },
      { to: '/credential-pools', label: 'HA Pools', icon: Layers },
    ],
  },
  {
    title: 'OBSERVABILITY',
    items: [
      { to: '/api-keys', label: 'Ingress API Keys', icon: Key },
      { to: '/usage', label: 'Usage & Telemetry', icon: BarChart3 },
      { to: '/audit-logs', label: 'Audit Logs', icon: FileText },
      { to: '/quota', label: 'Quotas & Circuits', icon: Gauge },
    ],
  },
  {
    title: 'SYSTEM',
    items: [
      { to: '/settings', label: 'Settings', icon: Settings },
      { to: '/backup', label: 'Database Backup', icon: Database },
      { to: '/api-docs', label: 'API Reference', icon: BookOpen },
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
                  Universal AI Gateway
                </span>
              </div>
            )}
          </div>

          <div className="flex items-center">
            <button
              type="button"
              onClick={onToggleCollapse}
              title={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
              className="hidden lg:flex p-1 rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)] transition-colors"
            >
              {isCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="p-1 rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] lg:hidden"
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
            {!isCollapsed && (
              <ChevronRight className="w-3.5 h-3.5 text-[var(--text-muted)] shrink-0" />
            )}
          </div>
        </div>
      </aside>
    </>
  )
}
