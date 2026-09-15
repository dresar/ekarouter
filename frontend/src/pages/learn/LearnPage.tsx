import { useState, useEffect } from 'react'
import {
  BookOpen,
  Copy,
  Check,
  Cpu,
  Terminal,
  Shield,
  Zap,
  Sparkles,
  Layers,
  Info,
  Server,
  KeyRound,
  Plus,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

interface DocSection {
  id: string
  title: string
  icon: React.ComponentType<{ className?: string }>
}

interface AIDocsResponse {
  project: string
  version: string
  content: string
  endpoints: Record<string, string>
}

interface AIKeyInfo {
  id: string
  name: string
  prefix: string
  scopes: string
}

interface ModelItem {
  id: string
  external_name: string
  display_name: string
  context_limit: number
  provider_name: string
}

interface RouteItem {
  id: string
  name: string
  strategy: string
}

interface ModelsResponse {
  models: ModelItem[]
  combos: RouteItem[]
  total_models: number
  total_combos: number
}

export function LearnPage() {
  const [activeTab, setActiveTab] = useState<'docs' | 'models' | 'prompt' | 'mcp' | 'skills' | 'quickstart'>('docs')
  const [copiedKey, setCopiedKey] = useState<string | null>(null)
  const [docsData, setDocsData] = useState<AIDocsResponse | null>(null)
  const [apiKeys, setApiKeys] = useState<AIKeyInfo[]>([])
  const [modelsData, setModelsData] = useState<ModelsResponse | null>(null)
  const [liveGeneratedKey, setLiveGeneratedKey] = useState<string | null>(null)
  const [isGeneratingKey, setIsGeneratingKey] = useState(false)
  const [toastMessage, setToastMessage] = useState<string | null>(null)

  const showToast = (msg: string) => {
    setToastMessage(msg)
    setTimeout(() => setToastMessage(null), 2000)
  }

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedKey(id)
    showToast('✓ Disalin!')
    setTimeout(() => setCopiedKey(null), 2000)
  }

  const handleGenerateKey = async () => {
    setIsGeneratingKey(true)
    try {
      const res = await api.post<{ success: boolean; key: string; name: string }>('/api/ai/keys', {
        name: 'Learn AI Client',
      })
      if (res && res.key) {
        setLiveGeneratedKey(res.key)
        showToast('✓ Dibuat!')
      }
    } catch {
      showToast('✕ Gagal!')
    } finally {
      setIsGeneratingKey(false)
    }
  }

  useEffect(() => {
    api.get<AIDocsResponse>('/api/ai/docs?format=json')
      .then(setDocsData)
      .catch(() => {})

    api.get<{ keys: AIKeyInfo[] }>('/api/ai/keys')
      .then((res) => {
        if (res && res.keys) setApiKeys(res.keys)
      })
      .catch(() => {})

    api.get<ModelsResponse>('/api/ai/models')
      .then(setModelsData)
      .catch(() => {})
  }, [])

  const currentAuthKey = liveGeneratedKey || 'YOUR_API_KEY'

  const masterPrompt = `You are an autonomous AI engineer connected to EkaRouter (Enterprise AI Gateway & Smart Model Router).

Base URL: http://localhost:8080
Completions: http://localhost:8080/v1/chat/completions
Models: http://localhost:8080/v1/models
MCP Server: http://localhost:8080/mcp (or /mcp/sse)

Guidelines:
1. Use combo routes (fast, smart, code) for automatic multi-provider fallback.
2. Maintain "X-Token-Saver: conservative" header to compress context and reduce tokens.
3. Enforce "ui-ux-text" microcopy standard: 1-word placeholders, 1-2 word buttons, 1-sentence subtitles.`

  const mcpConfig = `{
  "mcpServers": {
    "ekarouter": {
      "url": "http://localhost:8080/mcp/sse"
    }
  }
}`

  const curlSnippet = `curl -X POST http://localhost:8080/v1/chat/completions \\
  -H "Authorization: Bearer ${currentAuthKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "fast",
    "messages": [{"role": "user", "content": "Hello EkaRouter"}]
  }'`

  const pythonSnippet = `from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="${currentAuthKey}"
)

res = client.chat.completions.create(
    model="fast",
    messages=[{"role": "user", "content": "Hello EkaRouter"}]
)
print(res.choices[0].message.content)`

  const SECTIONS: DocSection[] = [
    { id: 'docs', title: 'Dokumentasi', icon: BookOpen },
    { id: 'models', title: 'Model', icon: Server },
    { id: 'prompt', title: 'Prompt', icon: Cpu },
    { id: 'mcp', title: 'MCP', icon: Layers },
    { id: 'skills', title: 'Skill', icon: Sparkles },
    { id: 'quickstart', title: 'Integrasi', icon: Terminal },
  ]

  return (
    <div className="space-y-4">
      <PageHeader
        title="Dokumentasi AI"
        description="Dokumentasi resmi, master prompt AI, dan protokol MCP."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'Learn' },
        ]}
        metadata={
          <span className="flex items-center gap-1.5 text-blue-400 text-[12px] font-mono">
            <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
            Gateway v1.0.0
          </span>
        }
      />

      {toastMessage && (
        <div className="fixed bottom-5 right-5 z-50 px-3 py-1.5 rounded-[6px] bg-blue-600 text-white text-[12px] font-medium shadow-lg animate-in fade-in duration-200">
          {toastMessage}
        </div>
      )}

      {/* Navigation Tabs */}
      <div className="flex items-center gap-1.5 border-b border-[var(--border-subtle)] pb-2 overflow-x-auto">
        {SECTIONS.map((sec) => {
          const Icon = sec.icon
          const isActive = activeTab === sec.id
          return (
            <button
              key={sec.id}
              type="button"
              onClick={() => setActiveTab(sec.id as any)}
              className={`flex items-center gap-2 px-3 py-1.5 rounded-[6px] text-[12px] font-medium transition-colors cursor-pointer select-none ${
                isActive
                  ? 'bg-[var(--brand-subtle)] text-[var(--brand-text)] border border-[var(--brand-primary)]/30'
                  : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-panel)]'
              }`}
            >
              <Icon className="w-3.5 h-3.5" />
              <span>{sec.title}</span>
            </button>
          )
        })}
      </div>

      {/* Tab: Docs */}
      {activeTab === 'docs' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <div className="lg:col-span-2 space-y-4">
            <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
                    Arsitektur Sistem
                  </h3>
                  <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                    Alur gateway, fallback dinamis, dan optimasi token.
                  </p>
                </div>
                <Button
                  variant="secondary"
                  size="compact"
                  onClick={() => copyToClipboard(docsData?.content || '', 'docs-content')}
                  leftIcon={copiedKey === 'docs-content' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                >
                  Salin
                </Button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-[12px]">
                <div className="p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-1">
                  <span className="font-semibold text-[var(--text-primary)] flex items-center gap-1.5">
                    <Server className="w-3.5 h-3.5 text-blue-400" />
                    Ingress Gateway
                  </span>
                  <p className="text-[var(--text-muted)] leading-relaxed">
                    Mendukung format OpenAI standar di <code className="text-blue-400">/v1/chat/completions</code> dengan SSE streaming.
                  </p>
                </div>

                <div className="p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-1">
                  <span className="font-semibold text-[var(--text-primary)] flex items-center gap-1.5">
                    <Zap className="w-3.5 h-3.5 text-amber-400" />
                    Token Saver
                  </span>
                  <p className="text-[var(--text-muted)] leading-relaxed">
                    Mengompresi prompt dan diff kode hingga 40% sebelum dikirim ke provider upstream.
                  </p>
                </div>

                <div className="p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-1">
                  <span className="font-semibold text-[var(--text-primary)] flex items-center gap-1.5">
                    <Shield className="w-3.5 h-3.5 text-emerald-400" />
                    Vault Kredensial
                  </span>
                  <p className="text-[var(--text-muted)] leading-relaxed">
                    Kunci provider dienkripsi menggunakan AES-256-GCM dengan rotasi cerdas dan cooldown.
                  </p>
                </div>

                <div className="p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-1">
                  <span className="font-semibold text-[var(--text-primary)] flex items-center gap-1.5">
                    <Layers className="w-3.5 h-3.5 text-purple-400" />
                    Routing Combo
                  </span>
                  <p className="text-[var(--text-muted)] leading-relaxed">
                    Otomatis beralih ke provider cadangan saat upstream mengalami limit atau galat.
                  </p>
                </div>
              </div>

              <div className="rounded-[6px] bg-[#0f172a] p-3 text-[#f1f5f9] font-mono text-[11.5px] overflow-x-auto max-h-[300px]">
                <pre>{docsData?.content || 'Memuat dokumentasi...'}</pre>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-3">
              <h4 className="text-[13px] font-semibold text-[var(--text-primary)] flex items-center gap-2">
                <Info className="w-4 h-4 text-blue-400" />
                Endpoint Publik
              </h4>
              <div className="space-y-2 text-[12px]">
                <div className="p-2 rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
                  <span className="text-[10px] font-bold text-blue-400 block font-mono">GET</span>
                  <code className="text-[11.5px] text-[var(--text-primary)]">/api/ai/docs</code>
                  <span className="text-[10.5px] text-[var(--text-muted)] block mt-0.5">Dokumentasi JSON</span>
                </div>
                <div className="p-2 rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
                  <span className="text-[10px] font-bold text-blue-400 block font-mono">GET</span>
                  <code className="text-[11.5px] text-[var(--text-primary)]">/api/ai/models</code>
                  <span className="text-[10.5px] text-[var(--text-muted)] block mt-0.5">Katalog Model</span>
                </div>
                <div className="p-2 rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
                  <span className="text-[10px] font-bold text-blue-400 block font-mono">GET</span>
                  <code className="text-[11.5px] text-[var(--text-primary)]">/api/ai/prompt</code>
                  <span className="text-[10.5px] text-[var(--text-muted)] block mt-0.5">Master Prompt</span>
                </div>
                <div className="p-2 rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
                  <span className="text-[10px] font-bold text-emerald-400 block font-mono">POST</span>
                  <code className="text-[11.5px] text-[var(--text-primary)]">/api/ai/keys</code>
                  <span className="text-[10.5px] text-[var(--text-muted)] block mt-0.5">Generate API Key</span>
                </div>
                <div className="p-2 rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
                  <span className="text-[10px] font-bold text-emerald-400 block font-mono">POST</span>
                  <code className="text-[11.5px] text-[var(--text-primary)]">/mcp</code>
                  <span className="text-[10.5px] text-[var(--text-muted)] block mt-0.5">JSON-RPC MCP</span>
                </div>
                <div className="p-2 rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
                  <span className="text-[10px] font-bold text-emerald-400 block font-mono">GET</span>
                  <code className="text-[11.5px] text-[var(--text-primary)]">/mcp/sse</code>
                  <span className="text-[10.5px] text-[var(--text-muted)] block mt-0.5">MCP SSE Stream</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab: Models & Combos */}
      {activeTab === 'models' && (
        <div className="space-y-4">
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
                  Routing Combo
                </h3>
                <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                  Virtual model combo dengan failover otomatis antar provider.
                </p>
              </div>
              <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-purple-500/20 text-purple-300">
                {modelsData?.total_combos || 0} Combo
              </span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              {(modelsData?.combos || [
                { id: 'fast', name: 'fast', strategy: 'priority' },
                { id: 'smart', name: 'smart', strategy: 'priority' },
                { id: 'code', name: 'code', strategy: 'weighted' },
              ]).map((combo) => (
                <div key={combo.id} className="p-3.5 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-1">
                  <div className="flex items-center justify-between">
                    <code className="text-[13px] font-bold text-blue-400 font-mono">{combo.name}</code>
                    <span className="text-[10px] px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-300 font-mono uppercase">
                      {combo.strategy}
                    </span>
                  </div>
                  <p className="text-[11px] text-[var(--text-muted)]">
                    Gunakan model: &quot;{combo.name}&quot; untuk failover otomatis.
                  </p>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
                  Katalog Model
                </h3>
                <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                  Model aktif yang terdaftar di upstream providers.
                </p>
              </div>
              <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-blue-500/20 text-blue-300">
                {modelsData?.total_models || 0} Model
              </span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {(modelsData?.models || []).map((m) => (
                <div key={m.id} className="p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="text-[12px] font-semibold text-[var(--text-primary)] truncate">
                      {m.display_name}
                    </span>
                    <span className="text-[10px] text-[var(--text-muted)] font-mono">
                      {m.provider_name}
                    </span>
                  </div>
                  <code className="text-[11px] text-blue-400 font-mono block">
                    {m.external_name}
                  </code>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Tab: Prompt */}
      {activeTab === 'prompt' && (
        <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
                Master Prompt AI
              </h3>
              <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                Gunakan prompt ini pada Cursor, Claude Code, atau Antigravity.
              </p>
            </div>
            <Button
              variant="primary"
              size="compact"
              onClick={() => copyToClipboard(masterPrompt, 'master-prompt')}
              leftIcon={copiedKey === 'master-prompt' ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
            >
              Salin
            </Button>
          </div>

          <div className="rounded-[6px] bg-[#0f172a] p-4 text-[#f1f5f9] font-mono text-[12px] overflow-x-auto">
            <pre className="whitespace-pre-wrap">{masterPrompt}</pre>
          </div>
        </div>
      )}

      {/* Tab: MCP */}
      {activeTab === 'mcp' && (
        <div className="space-y-4">
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
                  Konfigurasi MCP
                </h3>
                <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                  Tambahkan ke file konfigurasi Claude Desktop atau Cursor.
                </p>
              </div>
              <Button
                variant="primary"
                size="compact"
                onClick={() => copyToClipboard(mcpConfig, 'mcp-config')}
                leftIcon={copiedKey === 'mcp-config' ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
              >
                Salin
              </Button>
            </div>

            <div className="rounded-[6px] bg-[#0f172a] p-4 text-[#f1f5f9] font-mono text-[12px]">
              <pre>{mcpConfig}</pre>
            </div>
          </div>

          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-3">
            <h4 className="text-[13px] font-semibold text-[var(--text-primary)]">
              Daftar Tool MCP
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 text-[12px]">
              {[
                { name: 'generate_api_key', desc: 'Buat API key baru untuk agent eksternal.' },
                { name: 'get_api_keys', desc: 'Ambil API key gateway.' },
                { name: 'chat_completion', desc: 'Eksekusi inferensi multi-provider.' },
                { name: 'list_available_models', desc: 'Daftar model dan combo aktif.' },
                { name: 'get_project_docs', desc: 'Ambil dokumentasi resmi proyek.' },
                { name: 'get_ai_prompt', desc: 'Ambil master prompt sistem AI.' },
                { name: 'compact_tokens', desc: 'Kompresi token via TokenSaver.' },
                { name: 'check_gateway_health', desc: 'Periksa status kesehatan gateway.' },
              ].map((tool) => (
                <div key={tool.name} className="p-3 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)]">
                  <code className="text-blue-400 font-semibold block text-[11.5px]">{tool.name}</code>
                  <span className="text-[var(--text-muted)] mt-1 block text-[11px]">{tool.desc}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Tab: Skills */}
      {activeTab === 'skills' && (
        <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-4">
          <div>
            <h3 className="text-[14px] font-semibold text-[var(--text-primary)]">
              Skill AI
            </h3>
            <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
              Standar baku dan kapabilitas AI untuk proyek EkaRouter.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-[12px]">
            <div className="p-4 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-2">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-blue-400 font-mono">ui-ux-text</span>
                <span className="text-[10px] px-2 py-0.5 rounded bg-blue-500/20 text-blue-300 font-mono">Standar</span>
              </div>
              <p className="text-[var(--text-muted)] text-[11.5px]">
                Standar wajib mikro-kopi UI: placeholder 1 kata, tombol 1-2 kata, heading 1-3 kata, tanpa teks mubazir.
              </p>
            </div>

            <div className="p-4 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-2">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-emerald-400 font-mono">ekarouter</span>
                <span className="text-[10px] px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 font-mono">Gateway</span>
              </div>
              <p className="text-[var(--text-muted)] text-[11.5px]">
                Setup env, penemuan model, pemeriksaan kesehatan gateway, dan konfigurasi API key.
              </p>
            </div>

            <div className="p-4 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-2">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-purple-400 font-mono">ekarouter-chat</span>
                <span className="text-[10px] px-2 py-0.5 rounded bg-purple-500/20 text-purple-300 font-mono">Inference</span>
              </div>
              <p className="text-[var(--text-muted)] text-[11.5px]">
                Eksekusi chat completions OpenAI-compatible, streaming SSE, tool calling, dan combo fallback.
              </p>
            </div>

            <div className="p-4 rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] space-y-2">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-amber-400 font-mono">ekarouter-tokensaver</span>
                <span className="text-[10px] px-2 py-0.5 rounded bg-amber-500/20 text-amber-300 font-mono">Optimizer</span>
              </div>
              <p className="text-[var(--text-muted)] text-[11.5px]">
                Kompaksi aman untuk diff git, baris berulang, dan keluaran tool demi menghemat token.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Tab: Quickstart / Integrasi */}
      {activeTab === 'quickstart' && (
        <div className="space-y-4">
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-[13px] font-semibold text-[var(--text-primary)] flex items-center gap-2">
                  <KeyRound className="w-4 h-4 text-emerald-400" />
                  Kunci Gateway
                  {apiKeys.length > 0 && !liveGeneratedKey && (
                    <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-300">
                      {apiKeys.length} Kunci
                    </span>
                  )}
                </h4>
                <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                  {liveGeneratedKey ? 'Kunci aktif siap digunakan.' : 'Buat kunci API untuk langsung mencoba inferensi.'}
                </p>
              </div>
              <Button
                variant="primary"
                size="compact"
                onClick={handleGenerateKey}
                isLoading={isGeneratingKey}
                leftIcon={<Plus className="w-3.5 h-3.5" />}
              >
                Buat Key
              </Button>
            </div>

            {liveGeneratedKey && (
              <div className="p-3 rounded-[6px] bg-emerald-950/20 border border-emerald-500/30 text-[12px] flex items-center justify-between">
                <div className="font-mono text-emerald-300 truncate max-w-md">
                  {liveGeneratedKey}
                </div>
                <Button
                  variant="secondary"
                  size="compact"
                  onClick={() => copyToClipboard(liveGeneratedKey, 'live-key')}
                  leftIcon={copiedKey === 'live-key' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                >
                  Salin
                </Button>
              </div>
            )}
          </div>

          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-[13px] font-semibold text-[var(--text-primary)]">
                  Contoh cURL
                </h4>
                <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                  Inferensi langsung via terminal.
                </p>
              </div>
              <Button
                variant="secondary"
                size="compact"
                onClick={() => copyToClipboard(curlSnippet, 'curl-code')}
                leftIcon={copiedKey === 'curl-code' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              >
                Salin
              </Button>
            </div>
            <div className="rounded-[6px] bg-[#0f172a] p-4 text-[#f1f5f9] font-mono text-[11.5px]">
              <pre>{curlSnippet}</pre>
            </div>
          </div>

          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-5 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-[13px] font-semibold text-[var(--text-primary)]">
                  Contoh Python
                </h4>
                <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                  Gunakan OpenAI Python SDK.
                </p>
              </div>
              <Button
                variant="secondary"
                size="compact"
                onClick={() => copyToClipboard(pythonSnippet, 'python-code')}
                leftIcon={copiedKey === 'python-code' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              >
                Salin
              </Button>
            </div>
            <div className="rounded-[6px] bg-[#0f172a] p-4 text-[#f1f5f9] font-mono text-[11.5px]">
              <pre>{pythonSnippet}</pre>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
