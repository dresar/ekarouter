import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Wrench, Plus, ExternalLink, RefreshCw, Search, FileCode } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { DataTable, Column } from '../../components/ui/DataTable.tsx'
import { Tabs } from '../../components/ui/Tabs.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { ErrorBanner } from '../../components/ui/ErrorBanner.tsx'
import { api } from '../../api/client.ts'
import { ToolDefinition, RequestTemplate } from '../../types/api.ts'

export function ToolsListPage() {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'tools' | 'templates'>('tools')
  const [tools, setTools] = useState<ToolDefinition[]>([])
  const [templates, setTemplates] = useState<RequestTemplate[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')

  const loadData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [toolsData, templatesData] = await Promise.all([
        api.get<ToolDefinition[]>('/api/v1/tools').catch(() => []),
        api.get<RequestTemplate[]>('/api/v1/request-templates').catch(() => []),
      ])
      setTools(toolsData || [])
      setTemplates(templatesData || [])
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load tools'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const toolColumns: Column<ToolDefinition>[] = [
    {
      key: 'name',
      title: 'Tool Name',
      render: (item) => (
        <div className="flex flex-col">
          <Link
            to={`/tools/${item.id}`}
            className="font-semibold text-[13px] text-[var(--text-primary)] hover:text-[var(--brand-text)] hover:underline"
          >
            {item.name}
          </Link>
          <span className="text-[10.5px] text-[var(--text-muted)] truncate max-w-xs">
            {item.description || item.url_template}
          </span>
        </div>
      ),
    },
    {
      key: 'category',
      title: 'Category',
      render: (item) => (
        <span className="px-2 py-0.5 rounded-[4px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[10.5px] font-mono text-[var(--text-secondary)] capitalize">
          {item.category}
        </span>
      ),
    },
    {
      key: 'method',
      title: 'Method',
      render: (item) => (
        <span
          className={`font-mono text-[10.5px] font-bold px-1.5 py-0.5 rounded ${
            item.method === 'GET'
              ? 'text-emerald-400 bg-emerald-500/10'
              : item.method === 'POST'
              ? 'text-blue-400 bg-blue-500/10'
              : item.method === 'DELETE'
              ? 'text-rose-400 bg-rose-500/10'
              : 'text-amber-400 bg-amber-500/10'
          }`}
        >
          {item.method}
        </span>
      ),
    },
    {
      key: 'url_template',
      title: 'URL Template',
      render: (item) => (
        <span className="font-mono text-[11px] text-[var(--text-muted)] truncate max-w-xs block">
          {item.url_template}
        </span>
      ),
    },
    {
      key: 'actions',
      title: 'Actions',
      width: '80px',
      render: (item) => (
        <button
          type="button"
          onClick={() => navigate(`/tools/${item.id}`)}
          title="Test execution & inspect schema"
          className="p-1 rounded text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
        >
          <ExternalLink className="w-4 h-4" />
        </button>
      ),
    },
  ]

  const templateColumns: Column<RequestTemplate>[] = [
    {
      key: 'name',
      title: 'Template Name',
      render: (item) => (
        <span className="font-semibold text-[13px] text-[var(--text-primary)]">{item.name}</span>
      ),
    },
    {
      key: 'method',
      title: 'Method',
      render: (item) => (
        <span className="font-mono text-[10.5px] font-bold text-[var(--brand-text)]">
          {item.method}
        </span>
      ),
    },
    {
      key: 'path',
      title: 'Path',
      render: (item) => (
        <span className="font-mono text-[11.5px] text-[var(--text-secondary)]">
          {item.path}
        </span>
      ),
    },
  ]

  const filteredTools = tools.filter(
    (t) =>
      t.name.toLowerCase().includes(search.toLowerCase()) ||
      t.category.toLowerCase().includes(search.toLowerCase()) ||
      t.url_template.toLowerCase().includes(search.toLowerCase())
  )

  const filteredTemplates = templates.filter(
    (tpl) =>
      tpl.name.toLowerCase().includes(search.toLowerCase()) ||
      tpl.path.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="space-y-4">
      <PageHeader
        title="Tools & Templates"
        description="Parameterized HTTP tools with SSRF guards and reusable developer request templates."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Tools & Templates' },
        ]}
        metadata={
          <span>
            {tools.length} tools &bull; {templates.length} templates
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="compact"
              onClick={loadData}
              isLoading={isLoading}
              leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
            >
              Refresh
            </Button>
            <Link to="/tools/new">
              <Button variant="primary" size="compact" leftIcon={<Plus className="w-3.5 h-3.5" />}>
                Define Tool
              </Button>
            </Link>
          </div>
        }
      />

      {error && <ErrorBanner message={error} onRetry={loadData} />}

      <Tabs
        items={[
          { key: 'tools', label: 'Generic Tools', icon: <Wrench className="w-4 h-4" />, badge: tools.length },
          { key: 'templates', label: 'Request Templates', icon: <FileCode className="w-4 h-4" />, badge: templates.length },
        ]}
        activeKey={activeTab}
        onChange={(k) => setActiveTab(k as 'tools' | 'templates')}
      />

      <div className="flex items-center gap-2.5 bg-[var(--bg-surface)] p-2.5 rounded-[8px] border border-[var(--border-subtle)]">
        <Search className="w-3.5 h-3.5 text-[var(--text-muted)] ml-1 pointer-events-none" />
        <input
          type="text"
          placeholder={`Search ${activeTab === 'tools' ? 'tools' : 'templates'} by name or URL...`}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full px-2 py-1 text-[12px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--brand-primary)]"
        />
      </div>

      {activeTab === 'tools' ? (
        <DataTable
          columns={toolColumns}
          data={filteredTools}
          isLoading={isLoading}
          keyExtractor={(t) => t.id}
          emptyMessage="No generic tools defined yet. Click 'Define Tool' to create one."
        />
      ) : (
        <DataTable
          columns={templateColumns}
          data={filteredTemplates}
          isLoading={isLoading}
          keyExtractor={(tpl) => tpl.id}
          emptyMessage="No request templates configured."
        />
      )}
    </div>
  )
}
