import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext.tsx'
import { ThemeProvider } from './context/ThemeContext.tsx'
import { AppShell } from './components/layout/AppShell.tsx'
import { ErrorBoundary } from './components/ui/ErrorBoundary.tsx'

import { LoginPage } from './pages/auth/LoginPage.tsx'
import { OverviewPage } from './pages/overview/OverviewPage.tsx'
import { LiveConsolePage } from './pages/console/LiveConsolePage.tsx'
import { ProvidersListPage } from './pages/providers/ProvidersListPage.tsx'
import { ProviderCreatePage } from './pages/providers/ProviderCreatePage.tsx'
import { ProviderDetailPage } from './pages/providers/ProviderDetailPage.tsx'
import { RoutingListPage } from './pages/routing/RoutingListPage.tsx'
import { RouteCreatePage } from './pages/routing/RouteCreatePage.tsx'
import { RouteDetailPage } from './pages/routing/RouteDetailPage.tsx'
import { BoostPage } from './pages/boost/BoostPage.tsx'
import { TokenSaverPage } from './pages/tokensaver/TokenSaverPage.tsx'
import { VaultListPage } from './pages/vault/VaultListPage.tsx'
import { VaultCreatePage } from './pages/vault/VaultCreatePage.tsx'
import { VaultDetailPage } from './pages/vault/VaultDetailPage.tsx'
import { ToolsListPage } from './pages/tools/ToolsListPage.tsx'
import { ToolCreatePage } from './pages/tools/ToolCreatePage.tsx'
import { ToolDetailPage } from './pages/tools/ToolDetailPage.tsx'
import { ProxiesListPage } from './pages/proxies/ProxiesListPage.tsx'
import { ProxyCreatePage } from './pages/proxies/ProxyCreatePage.tsx'
import { FreeTiersPage } from './pages/freetiers/FreeTiersPage.tsx'
import { PoolsListPage } from './pages/pools/PoolsListPage.tsx'
import { PoolDetailPage } from './pages/pools/PoolDetailPage.tsx'
import { ApiKeysPage } from './pages/apikeys/ApiKeysPage.tsx'
import { UsagePage } from './pages/usage/UsagePage.tsx'
import { AuditLogsPage } from './pages/audit/AuditLogsPage.tsx'
import { QuotaPage } from './pages/quota/QuotaPage.tsx'
import { SettingsPage } from './pages/settings/SettingsPage.tsx'
import { BackupPage } from './pages/backup/BackupPage.tsx'
import { ApiDocsPage } from './pages/apidocs/ApiDocsPage.tsx'
import { NotFoundPage } from './pages/error/NotFoundPage.tsx'

function ProtectedLayout() {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--bg-app)]">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 rounded-full border-2 border-[var(--border-strong)] border-t-[var(--brand-primary)] animate-spin" />
          <span className="text-[12px] font-mono text-[var(--text-muted)]">
            Authenticating Session...
          </span>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname)}`} replace />
  }

  return <AppShell />
}

export function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider>
        <AuthProvider>
          <BrowserRouter>
            <Routes>
              <Route path="/login" element={<LoginPage />} />

              <Route element={<ProtectedLayout />}>
                <Route path="/" element={<Navigate to="/overview" replace />} />
                <Route path="/overview" element={<OverviewPage />} />
                <Route path="/console" element={<LiveConsolePage />} />
                <Route path="/providers" element={<ProvidersListPage />} />
                <Route path="/providers/new" element={<ProviderCreatePage />} />
                <Route path="/providers/:id" element={<ProviderDetailPage />} />
                <Route path="/routing" element={<RoutingListPage />} />
                <Route path="/routing/new" element={<RouteCreatePage />} />
                <Route path="/routing/:id" element={<RouteDetailPage />} />
                <Route path="/models" element={<Navigate to="/providers" replace />} />
                <Route path="/boost" element={<BoostPage />} />
                <Route path="/token-saver" element={<TokenSaverPage />} />
                <Route path="/vault" element={<VaultListPage />} />
                <Route path="/vault/new" element={<VaultCreatePage />} />
                <Route path="/vault/:id" element={<VaultDetailPage />} />
                <Route path="/tools" element={<ToolsListPage />} />
                <Route path="/tools/new" element={<ToolCreatePage />} />
                <Route path="/tools/:id" element={<ToolDetailPage />} />
                <Route path="/proxies" element={<ProxiesListPage />} />
                <Route path="/proxies/new" element={<ProxyCreatePage />} />
                <Route path="/free-tiers" element={<FreeTiersPage />} />
                <Route path="/credential-pools" element={<PoolsListPage />} />
                <Route path="/credential-pools/:id" element={<PoolDetailPage />} />
                <Route path="/api-keys" element={<ApiKeysPage />} />
                <Route path="/usage" element={<UsagePage />} />
                <Route path="/audit-logs" element={<AuditLogsPage />} />
                <Route path="/quota" element={<QuotaPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/backup" element={<BackupPage />} />
                <Route path="/api-docs" element={<ApiDocsPage />} />
                <Route path="*" element={<NotFoundPage />} />
              </Route>
            </Routes>
          </BrowserRouter>
        </AuthProvider>
      </ThemeProvider>
    </ErrorBoundary>
  )
}
