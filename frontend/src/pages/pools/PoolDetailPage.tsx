import { useEffect, useState, FormEvent } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Layers,
  Play,
  Pause,
  RotateCw,
  Activity,
  Plus,
  Trash2,
  Save,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { StatusBadge } from '../../components/ui/StatusBadge.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { InlineConfirm } from '../../components/ui/InlineConfirm.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { CredentialPool, PoolMember, Account } from '../../types/api.ts'
import { SearchableSelect } from '../../components/ui/SearchableSelect.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'

export function PoolDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [pool, setPool] = useState<CredentialPool | null>(null)
  const [members, setMembers] = useState<PoolMember[]>([])
  const [availableAccounts, setAvailableAccounts] = useState<Account[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showAddMember, setShowAddMember] = useState(false)
  const [memberForm, setMemberForm] = useState({
    account_id: '',
    priority: 1,
    weight: 10,
  })
  const [isSubmittingMember, setIsSubmittingMember] = useState(false)
  const [actionFeedback, setActionFeedback] = useState<string | null>(null)

  const loadData = async () => {
    if (!id) return
    setIsLoading(true)
    setError(null)
    try {
      const [poolData, membersData, accountsData] = await Promise.all([
        api.get<CredentialPool>(`/api/credential-pools/${id}`),
        api.get<PoolMember[]>(`/api/credential-pools/${id}/credentials`).catch(() => []),
        api.get<Account[]>('/api/accounts').catch(() => []),
      ])
      setPool(poolData)
      setMembers(membersData || [])
      setAvailableAccounts(accountsData || [])
      if (accountsData && accountsData.length > 0 && !memberForm.account_id) {
        setMemberForm((prev) => ({ ...prev, account_id: accountsData[0].id }))
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load pool'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [id])

  const handleAction = async (action: 'pause' | 'resume' | 'rotate' | 'health') => {
    if (!id) return
    setActionFeedback(null)
    try {
      await api.post(`/api/credential-pools/${id}/${action}`)
      setActionFeedback(`Pool ${action} executed successfully`)
      await loadData()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : `Action ${action} failed`
      setError(msg)
    }
  }

  const handleAddMember = async (e: FormEvent) => {
    e.preventDefault()
    if (!id) return
    setIsSubmittingMember(true)
    try {
      await api.post(`/api/credential-pools/${id}/credentials`, {
        account_id: memberForm.account_id,
        priority: Number(memberForm.priority),
        weight: Number(memberForm.weight),
      })
      setShowAddMember(false)
      await loadData()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to add pool member'
      setError(msg)
    } finally {
      setIsSubmittingMember(false)
    }
  }

  const handleRemoveMember = async (memberId: string) => {
    if (!id) return
    try {
      await api.delete(`/api/credential-pools/${id}/credentials/${memberId}`)
      setMembers((prev) => prev.filter((m) => m.id !== memberId))
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to remove member'
      setError(msg)
    }
  }

  const columns: Column<PoolMember>[] = [
    {
      key: 'account_id',
      title: 'Account ID',
      render: (item) => {
        const acc = availableAccounts.find((a) => a.id === item.account_id)
        return (
          <div className="flex flex-col">
            <span className="font-medium text-[var(--text-primary)]">
              {acc ? acc.name : item.account_id}
            </span>
            <span className="text-[11px] font-mono text-[var(--text-muted)]">{item.account_id}</span>
          </div>
        )
      },
    },
    {
      key: 'priority',
      title: 'Priority',
      render: (item) => (
        <span className="font-mono text-[12px] text-[var(--brand-text)] font-semibold">
          Tier {item.priority}
        </span>
      ),
    },
    {
      key: 'weight',
      title: 'Weight',
      render: (item) => (
        <span className="font-mono text-[12px] text-[var(--text-secondary)]">
          {item.weight}
        </span>
      ),
    },
    {
      key: 'status',
      title: 'Status',
      render: (item) => (
        <StatusBadge variant={item.status === 'active' ? 'healthy' : 'warning'}>
          {item.status || 'Active'}
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
              className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--status-danger)] transition-colors"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          }
          confirmText="Remove?"
          onConfirm={() => handleRemoveMember(item.id)}
        />
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title={pool ? pool.name : 'Credential Pool'}
        description={`Cluster ID: ${id || ''}`}
        breadcrumbs={[
          { label: 'HA Pools', to: '/credential-pools' },
          { label: pool?.name || 'Inspect' },
        ]}
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="compact"
              onClick={() => navigate('/credential-pools')}
              leftIcon={<ArrowLeft className="w-3.5 h-3.5" />}
            >
              Back
            </Button>
            <Button
              variant="secondary"
              size="compact"
              onClick={() => handleAction('health')}
              leftIcon={<Activity className="w-3.5 h-3.5" />}
            >
              Health Check
            </Button>
            <Button
              variant="secondary"
              size="compact"
              onClick={() => handleAction('rotate')}
              leftIcon={<RotateCw className="w-3.5 h-3.5" />}
            >
              Rotate Now
            </Button>
            {pool?.status === 'active' ? (
              <Button
                variant="secondary"
                size="compact"
                onClick={() => handleAction('pause')}
                leftIcon={<Pause className="w-3.5 h-3.5" />}
              >
                Pause Pool
              </Button>
            ) : (
              <Button
                variant="primary"
                size="compact"
                onClick={() => handleAction('resume')}
                leftIcon={<Play className="w-3.5 h-3.5" />}
              >
                Resume Pool
              </Button>
            )}
          </div>
        }
      />

      {error && <ErrorBanner message={error} />}

      {actionFeedback && (
        <div className="p-3 rounded-[6px] bg-[var(--status-success)]/10 border border-[var(--status-success)]/30 text-[12.5px] text-[var(--status-success)]">
          {actionFeedback}
        </div>
      )}

      {pool && (
        <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[8px] p-5">
          <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)]">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded bg-[var(--bg-panel)] flex items-center justify-center text-[var(--brand-text)]">
                <Layers className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-[16px] font-semibold text-[var(--text-primary)]">{pool.name}</h2>
                <div className="flex items-center gap-2 mt-0.5 text-[11px] font-mono text-[var(--text-muted)]">
                  <span>provider: {pool.provider_id}</span>
                  <span>•</span>
                  <span>env: {pool.environment}</span>
                  <span>•</span>
                  <span>strategy: {pool.strategy || 'round_robin'}</span>
                </div>
              </div>
            </div>

            <StatusBadge variant={pool.status === 'active' ? 'healthy' : 'disabled'}>
              {pool.status || 'Active'}
            </StatusBadge>
          </div>
        </div>
      )}

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">Pool Members</h3>
            <p className="text-[11px] text-[var(--text-muted)]">
              Accounts enrolled into this high-availability rotation group
            </p>
          </div>
          <Button
            variant="secondary"
            size="compact"
            onClick={() => setShowAddMember(!showAddMember)}
            leftIcon={<Plus className="w-3.5 h-3.5" />}
          >
            {showAddMember ? 'Close' : 'Add Member'}
          </Button>
        </div>

        {showAddMember && (
          <form
            onSubmit={handleAddMember}
            className="bg-[var(--bg-panel)]/50 border border-[var(--border-strong)] rounded-[8px] p-4 space-y-3 animate-in fade-in duration-150"
          >
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div>
                <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                  Select Account *
                </label>
                <SearchableSelect
                  value={memberForm.account_id}
                  onChange={(val) => setMemberForm({ ...memberForm, account_id: val })}
                  placeholder="Select Account..."
                  options={availableAccounts.map((a) => ({
                    value: a.id,
                    label: a.name,
                    sublabel: `(${a.provider_id})`,
                    icon: <ProviderLogo providerId={a.provider_id} name={a.name} size="sm" />,
                  }))}
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                  Priority (1 = Highest)
                </label>
                <input
                  type="number"
                  min="1"
                  max="10"
                  value={memberForm.priority}
                  onChange={(e) =>
                    setMemberForm({ ...memberForm, priority: parseInt(e.target.value) || 1 })
                  }
                  className="w-full px-2.5 py-1.5 text-[12.5px] rounded focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-[var(--text-secondary)] mb-1">
                  Weight (1 - 100)
                </label>
                <input
                  type="number"
                  min="1"
                  max="100"
                  value={memberForm.weight}
                  onChange={(e) =>
                    setMemberForm({ ...memberForm, weight: parseInt(e.target.value) || 10 })
                  }
                  className="w-full px-2.5 py-1.5 text-[12.5px] rounded focus:outline-none"
                />
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="ghost"
                size="compact"
                onClick={() => setShowAddMember(false)}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="compact"
                isLoading={isSubmittingMember}
                leftIcon={<Save className="w-3.5 h-3.5" />}
              >
                Enroll Member
              </Button>
            </div>
          </form>
        )}

        <DataTable
          columns={columns}
          data={members}
          isLoading={isLoading}
          keyExtractor={(m) => m.id}
          emptyMessage="No members in this pool yet. Click 'Add Member' to enroll accounts."
        />
      </div>
    </div>
  )
}
