import { useState } from 'react'
import { Outlet, Link } from 'react-router-dom'
import { ShieldAlert, X, ArrowRight } from 'lucide-react'
import { useAuth } from '../../context/AuthContext.tsx'
import { Sidebar } from './Sidebar.tsx'
import { Topbar } from './Topbar.tsx'

export function AppShell() {
  const { user } = useAuth()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [isCollapsed, setIsCollapsed] = useState(false)
  const [showSecurityNotice, setShowSecurityNotice] = useState(true)

  return (
    <div className="min-h-screen flex flex-col bg-[var(--bg-app)]">
      <Sidebar
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        isCollapsed={isCollapsed}
        onToggleCollapse={() => setIsCollapsed(!isCollapsed)}
      />

      <div
        className={`flex-1 flex flex-col min-w-0 transition-all duration-200 ${
          isCollapsed ? 'lg:pl-[64px]' : 'lg:pl-[230px]'
        }`}
      >
        <Topbar onToggleSidebar={() => setSidebarOpen(!sidebarOpen)} />

        {user?.is_default_password && showSecurityNotice && (
          <div className="bg-amber-950/20 border-b border-amber-600/30 px-4 py-2.5 flex items-center justify-between text-[12px] text-amber-200/90 shrink-0">
            <div className="flex items-center gap-2.5">
              <ShieldAlert className="w-4 h-4 text-amber-400 shrink-0" />
              <div className="flex items-center gap-1.5 flex-wrap">
                <span>Default credentials detected. Update your password in Settings.</span>
              </div>
            </div>
            <div className="flex items-center gap-3 shrink-0 ml-2">
              <Link
                to="/settings"
                className="flex items-center gap-1 text-[11.5px] font-medium text-amber-300 hover:text-amber-100 hover:underline"
              >
                <span>Go to Settings</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </Link>
              <button
                type="button"
                onClick={() => setShowSecurityNotice(false)}
                className="p-1 rounded text-amber-400/80 hover:text-amber-200 transition-colors"
                aria-label="Dismiss security notice"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        )}

        <main className="flex-1 p-4 md:p-6 w-full max-w-7xl mx-auto bg-grid">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
