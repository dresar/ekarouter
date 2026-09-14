import { useEffect, useState, FormEvent } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Plus,
  Trash2,
  Server,
  AlertCircle,
  Save,
  ChevronUp,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { Provider, Account, VaultCredential, PlatformProvider } from '../../types/api.ts'

export function ProviderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [provider, setProvider] = useState<Provider | null>(null)
  const [accounts, setAccounts] = useState<Account[]>([])
  const [credentials, setCredentials] = useState<VaultCredential[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showAddAccount, setShowAddAccount] = useState(false)
  const [accountForm, setAccountForm] = useState({
    name: '',
    auth_type: 'api_key',
    priority: 1,
    api_key: '',
    secret_key: '',
  })
  const [isSubmittingAccount, setIsSubmittingAccount] = useState(false)
  const [accountError, setAccountError] = useState<string | null>(null)

  const loadData = async () => {
    if (!id) return
    setIsLoading(true)
    setError(null)
    try {
      const [allProvidersRes, platformProvidersRes, allAccountsRes, allCredentialsRes] = await Promise.allSettled([
        api.get<Provider[]>('/api/providers'),
        api.get<PlatformProvider[]>('/api/v1/providers'),
        api.get<Account[]>('/api/accounts'),
        api.get<VaultCredential[]>('/api/v1/credentials'),
      ])

      const allProviders: Provider[] = allProvidersRes.status === 'fulfilled' && Array.isArray(allProvidersRes.value) ? allProvidersRes.value : []
      const platformProviders: PlatformProvider[] = platformProvidersRes.status === 'fulfilled' && Array.isArray(platformProvidersRes.value) ? platformProvidersRes.value : []
      const allAccounts: Account[] = allAccountsRes.status === 'fulfilled' && Array.isArray(allAccountsRes.value) ? allAccountsRes.value : []
      const allCredentials: VaultCredential[] = allCredentialsRes.status === 'fulfilled' && Array.isArray(allCredentialsRes.value) ? allCredentialsRes.value : []

      let found = allProviders.find((p) => p.id === id || p.key === id)
      if (!found) {
        const platFound = platformProviders.find((p) => p.id === id)
        if (platFound) {
          found = {
            id: platFound.id,
            key: platFound.id,
            name: platFound.name,
            kind: platFound.category || 'ai',
            base_url: platFound.base_url,
            enabled: true,
          }
        } else {
          try {
            const single = await api.get<PlatformProvider>(`/api/v1/providers/${id}`)
            if (single && single.id) {
              found = {
                id: single.id,
                key: single.id,
                name: single.name,
                kind: single.category || 'ai',
                base_url: single.base_url,
                enabled: true,
              }
            }
          } catch {
            // provider not found
          }
        }
      }

      if (!found) {
        setError('Provider not found')
        return
      }

      setProvider(found)
      const matchingAccounts = allAccounts.filter((a) => a.provider_id === id || a.provider_id === found?.key)
      setAccounts(matchingAccounts)
      const matchingCreds = allCredentials.filter((c) => c.provider_id === id || c.provider_id === found?.key)
      setCredentials(matchingCreds)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load provider details'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [id])

  const handleCreateAccount = async (e: FormEvent) => {
    e.preventDefault()
    if (!provider) return
    setIsSubmittingAccount(true)
    setAccountError(null)

    try {
      await api.post('/api/accounts', {
        provider_id: provider.id,
        name: accountForm.name,
        auth_type: accountForm.auth_type,
        priority: Number(accountForm.priority),
        state: 'active',
        enabled: true,
        credentials: {
          access_key: accountForm.api_key,
          secret_key: accountForm.secret_key,
        },
      })
      setAccountForm({
        name: '',
        auth_type: 'api_key',
        priority: 1,
        api_key: '',
        secret_key: '',
      })
      setShowAddAccount(false)
      await loadData()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to add account'
      setAccountError(msg)
    } finally {
      setIsSubmittingAccount(false)
    }
  }

  const handleDeleteAccount = async (accountId: string) => {
    try {
      await api.delete(`/api/accounts/${accountId}`)
      setAccounts((prev) => prev.filter((a) => a.id !== accountId))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete account'
      setError(msg)
    }
  }

  const handleDeleteCredential = async (credId: string) => {
    try {
      await api.delete(`/api/v1/credentials/${credId}`)
      setCredentials((prev) => prev.filter((c) => c.id !== credId))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete credential'
      setError(msg)
    }
  }

  const credentialColumns: Column<VaultCredential>[] = [
    {
      key: 'name',
      title: 'Credential Name',
      render: (item) => (
        <span className="font-medium text-[var(--text-primary)]">{item.name}</span>
      ),
    },
    {
      key: 'credential_type',
      title: 'Auth Type',
      render: (item) => (
        <span className="font-mono text-[11px] uppercase text-[var(--text-secondary)]">
          {item.credential_type}
        </span>
      ),
    },
    {
      key: 'masked_value',
      title: 'Masked Secret',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--text-muted)]">
          {item.masked_value || '••••••••'}
        </span>
      ),
    },
    {
      key: 'status',
      title: 'Status',
      render: (item) => (
        <StatusBadge variant={item.status === 'active' ? 'healthy' : 'disabled'}>
          {item.status || 'active'}
        </StatusBadge>
      ),
    },
    {
      key: 'actions',
      title: 'Actions',
      width: '80px',
      render: (item) => (
        <InlineConfirm
          trigger={
            <button
              type="button"
              className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] hover:bg-[var(--status-danger)]/10 transition-colors"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          }
          confirmText="Delete?"
          onConfirm={() => handleDeleteCredential(item.id)}
        />
      ),
    },
  ]

  const accountColumns: Column<Account>[] = [
    {
      key: 'name',
      title: 'Account Name',
      render: (item) => (
        <span className="font-medium text-[var(--text-primary)]">{item.name}</span>
      ),
    },
    {
      key: 'auth_type',
      title: 'Auth Type',
      render: (item) => (
        <span className="font-mono text-[11px] uppercase text-[var(--text-secondary)]">
          {item.auth_type}
        </span>
      ),
    },
    {
      key: 'priority',
      title: 'Priority Tier',
      render: (item) => (
        <span className="font-mono text-[12px] text-[var(--brand-text)] font-semibold">
          Tier {item.priority}
        </span>
      ),
    },
    {
      key: 'state',
      title: 'Circuit State',
      render: (item) => (
        <StatusBadge
          variant={
            item.state === 'active'
              ? 'healthy'
              : item.state === 'cooling_down'
              ? 'cooling_down'
              : 'disabled'
          }
        >
          {item.state}
        </StatusBadge>
      ),
    },
    {
      key: 'actions',
      title: 'Actions',
      width: '80px',
      render: (item) => (
        <InlineConfirm
          trigger={
            <button
              type="button"
              className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] hover:bg-[var(--status-danger)]/10 transition-colors"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          }
          confirmText="Delete?"
          onConfirm={() => handleDeleteAccount(item.id)}
        />
      ),
    },
  ]

  return (
    <div className="space-y-5">
      <PageHeader
        title={provider ? provider.name : 'Provider Details'}
        description={`Backend ID: ${id || ''}`}
        breadcrumbs={[
          { label: 'Providers', to: '/providers' },
          { label: provider?.name || 'Details' },
        ]}
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="compact"
              onClick={() => navigate('/providers')}
              leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}
            >
              Back
            </Button>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadData} />}

      {provider && (
        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-5">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-[var(--border-subtle)]">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded bg-[var(--bg-panel)] flex items-center justify-center text-[var(--brand-text)]">
                <Server className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-[16px] font-semibold text-[var(--text-primary)]">
                  {provider.name}
                </h2>
                <div className="flex items-center gap-2 mt-0.5">
                  <span className="font-mono text-[11px] text-[var(--text-muted)]">
                    key: {provider.key}
                  </span>
                  <span className="text-[var(--border-strong)]">•</span>
                  <span className="font-mono text-[11px] text-[var(--text-muted)]">
                    kind: {provider.kind}
                  </span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <StatusBadge variant={provider.enabled ? 'healthy' : 'disabled'}>
                {provider.enabled ? 'Active Backend' : 'Disabled'}
              </StatusBadge>
            </div>
          </div>

          <div className="mt-4 grid grid-cols-1 sm:grid-cols-2 gap-4 text-[12.5px]">
            <div>
              <span className="text-[11px] font-semibold uppercase text-[var(--text-muted)] block">
                Upstream URL
              </span>
              <span className="font-mono text-[var(--text-secondary)] break-all">
                {provider.base_url}
              </span>
            </div>
            <div>
              <span className="text-[11px] font-semibold uppercase text-[var(--text-muted)] block">
                Active Accounts
              </span>
              <span className="font-mono text-[var(--text-primary)] font-medium">
                {accounts.length + credentials.length} credentials configured
              </span>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
              Configured Accounts & Keys
            </h3>
            <p className="text-[11px] text-[var(--text-muted)]">
              Multi-account credentials with priority failover ordering
            </p>
          </div>
          <Button
            variant="secondary"
            size="compact"
            onClick={() => setShowAddAccount(!showAddAccount)}
            leftIcon={showAddAccount ? <ChevronUp className="w-3.5 h-3.5" /> : <Plus className="w-3.5 h-3.5" />}
          >
            {showAddAccount ? 'Close Form' : 'Add Account'}
          </Button>
        </div>

        {showAddAccount && (
          <form
            onSubmit={handleCreateAccount}
            className="bg-[var(--bg-panel)]/50 border border-[var(--border-strong)] rounded-[8px] p-4 space-y-3.5 animate-in fade-in duration-150"
          >
            <div className="flex items-center justify-between border-b border-[var(--border-strong)] pb-2">
              <span className="text-[12px] font-semibold text-[var(--text-primary)]">
                New Account Credentials
              </span>
              <span className="text-[11px] text-[var(--text-muted)]">
                AES-256-GCM encrypted in SQLite
              </span>
            </div>

            {accountError && (
              <div className="p-2.5 rounded bg-[var(--status-danger)]/10 text-[var(--status-danger)] text-[12px] flex items-center gap-1.5">
                <AlertCircle className="w-3.5 h-3.5" />
                <span>{accountError}</span>
              </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div>
                <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                  Account Name *
                </label>
                <input
                  type="text"
                  required
                  value={accountForm.name}
                  onChange={(e) => setAccountForm({ ...accountForm, name: e.target.value })}
                  placeholder="e.g. Master Production Key"
                  className="w-full px-2.5 py-1.5 text-[12.5px] rounded-[5px] focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                  Auth Type *
                </label>
                <select
                  value={accountForm.auth_type}
                  onChange={(e) => setAccountForm({ ...accountForm, auth_type: e.target.value })}
                  className="w-full px-2.5 py-1.5 text-[12.5px] rounded-[5px] focus:outline-none"
                >
                  <option value="api_key">API Key (Bearer)</option>
                  <option value="oauth">OAuth Token</option>
                  <option value="custom">Custom Header</option>
                </select>
              </div>

              <div>
                <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                  Priority (1 = Highest) *
                </label>
                <input
                  type="number"
                  min="1"
                  max="10"
                  required
                  value={accountForm.priority}
                  onChange={(e) =>
                    setAccountForm({ ...accountForm, priority: parseInt(e.target.value) || 1 })
                  }
                  className="w-full px-2.5 py-1.5 text-[12.5px] rounded-[5px] focus:outline-none"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                  API Key / Access Key *
                </label>
                <input
                  type="password"
                  required
                  autoComplete="new-password"
                  value={accountForm.api_key}
                  onChange={(e) => setAccountForm({ ...accountForm, api_key: e.target.value })}
                  placeholder="sk-..."
                  className="w-full px-2.5 py-1.5 text-[12.5px] font-mono rounded-[5px] focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                  Secret Key (Optional)
                </label>
                <input
                  type="password"
                  autoComplete="new-password"
                  value={accountForm.secret_key}
                  onChange={(e) => setAccountForm({ ...accountForm, secret_key: e.target.value })}
                  placeholder="For basic or dual-key auth"
                  className="w-full px-2.5 py-1.5 text-[12.5px] font-mono rounded-[5px] focus:outline-none"
                />
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="ghost"
                size="compact"
                onClick={() => setShowAddAccount(false)}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="compact"
                isLoading={isSubmittingAccount}
                leftIcon={<Save className="w-3.5 h-3.5" />}
              >
                Save Account
              </Button>
            </div>
          </form>
        )}

        <DataTable
          columns={accountColumns}
          data={accounts}
          isLoading={isLoading}
          keyExtractor={(a) => a.id}
          emptyMessage="No accounts configured for this provider. Add an account above to enable routing."
        />

        {credentials.length > 0 && (
          <div className="space-y-3 pt-4 border-t border-[var(--border-subtle)]">
            <div>
              <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
                Vault Credentials
              </h3>
              <p className="text-[11px] text-[var(--text-muted)]">
                AES-256-GCM encrypted credentials assigned to this provider
              </p>
            </div>
            <DataTable
              columns={credentialColumns}
              data={credentials}
              isLoading={isLoading}
              keyExtractor={(c) => c.id}
              emptyMessage="No vault credentials found."
            />
          </div>
        )}
      </div>
    </div>
  )
}
