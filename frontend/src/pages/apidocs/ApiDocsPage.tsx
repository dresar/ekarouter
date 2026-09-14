import { useState } from 'react'
import { Terminal, Copy, Check } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader.tsx'

interface DocEndpoint {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  path: string
  tag: string
  summary: string
  description: string
  requestBody?: string
  responseStatus: number
  responseSample: string
}

const ENDPOINTS: DocEndpoint[] = [
  {
    method: 'POST',
    path: '/v1/chat/completions',
    tag: 'Ingress AI Gateway',
    summary: 'OpenAI Chat Completions',
    description: 'Execute standard or SSE streaming chat completions routed across upstream providers.',
    requestBody: '{\n  "model": "gpt-4o",\n  "messages": [{"role": "user", "content": "Hello"}],\n  "stream": false\n}',
    responseStatus: 200,
    responseSample: '{\n  "id": "chatcmpl-8f3a",\n  "object": "chat.completion",\n  "model": "gpt-4o",\n  "choices": [{\n    "index": 0,\n    "message": {"role": "assistant", "content": "Hello! How can I assist you today?"},\n    "finish_reason": "stop"\n  }]\n}',
  },
  {
    method: 'GET',
    path: '/v1/models',
    tag: 'Ingress AI Gateway',
    summary: 'List Ingress Models',
    description: 'Returns OpenAI-compatible model catalog entries available through the gateway.',
    responseStatus: 200,
    responseSample: '{\n  "object": "list",\n  "data": [\n    {"id": "gpt-4o", "object": "model", "owned_by": "openai"}\n  ]\n}',
  },
  {
    method: 'GET',
    path: '/api/providers',
    tag: 'Providers',
    summary: 'List Configured Providers',
    description: 'Retrieve all configured upstream AI backends.',
    responseStatus: 200,
    responseSample: '[\n  {\n    "id": "prov_openai",\n    "key": "openai",\n    "name": "OpenAI Commercial",\n    "kind": "openai",\n    "base_url": "https://api.openai.com/v1",\n    "enabled": true\n  }\n]',
  },
  {
    method: 'POST',
    path: '/api/providers',
    tag: 'Providers',
    summary: 'Create or Update Provider',
    description: 'Register an upstream AI backend and connection endpoint.',
    requestBody: '{\n  "key": "groq",\n  "name": "Groq Cloud",\n  "kind": "groq",\n  "base_url": "https://api.groq.com/openai/v1",\n  "enabled": true\n}',
    responseStatus: 200,
    responseSample: '{"success": true}',
  },
  {
    method: 'GET',
    path: '/api/routes',
    tag: 'Routing',
    summary: 'List Model Routes',
    description: 'Returns virtual routing combos with target accounts and fallback strategies.',
    responseStatus: 200,
    responseSample: '[\n  {\n    "id": "route_gpt4o",\n    "name": "gpt-4o",\n    "strategy": "priority",\n    "enabled": true,\n    "items": []\n  }\n]',
  },
  {
    method: 'GET',
    path: '/api/v1/credentials',
    tag: 'Credential Vault',
    summary: 'List Vault Credentials',
    description: 'Returns masked API keys and tokens stored in the AES-256-GCM vault.',
    responseStatus: 200,
    responseSample: '[\n  {\n    "id": "cred_123",\n    "name": "Production Key",\n    "provider_id": "openai",\n    "masked_value": "sk-****abcd",\n    "environment": "production",\n    "health_state": "healthy"\n  }\n]',
  },
  {
    method: 'POST',
    path: '/api/v1/tools/{id}/execute',
    tag: 'Tools',
    summary: 'Execute Parameterized Tool',
    description: 'Dispatches SSRF-safe parameterized tool execution with variable substitution.',
    requestBody: '{\n  "parameters": {\n    "to": "user@example.com"\n  }\n}',
    responseStatus: 200,
    responseSample: '{\n  "status_code": 200,\n  "latency_ms": 120,\n  "body": {"success": true}\n}',
  },
  {
    method: 'GET',
    path: '/health',
    tag: 'System',
    summary: 'Process Health Status',
    description: 'Process health, database readiness, and uptime indicator.',
    responseStatus: 200,
    responseSample: '{"status": "ok", "database": "ok", "uptime": "12h45m"}',
  },
]

export function ApiDocsPage() {
  const [selectedEndpoint, setSelectedEndpoint] = useState<DocEndpoint>(ENDPOINTS[0])
  const [copied, setCopied] = useState(false)

  const apiBaseUrl =
    (import.meta.env.VITE_API_BASE_URL as string) ||
    (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080')

  const curl = selectedEndpoint.requestBody
    ? `curl -X ${selectedEndpoint.method} ${apiBaseUrl}${selectedEndpoint.path} \\\n  -H "Authorization: Bearer eka_live_token" \\\n  -H "Content-Type: application/json" \\\n  -d '${selectedEndpoint.requestBody.replace(/\n/g, '\n  ')}'`
    : `curl -X ${selectedEndpoint.method} ${apiBaseUrl}${selectedEndpoint.path} \\\n  -H "Authorization: Bearer eka_live_token"`

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(curl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {}
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="API Reference"
        description="Endpoint schemas, expected payload formats, and cURL snippets for EkaRouter APIs."
        breadcrumbs={[
          { label: 'Home', to: '/overview' },
          { label: 'API Reference' },
        ]}
        metadata={
          <span>
            {ENDPOINTS.length} documented endpoints
          </span>
        }
      />

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        <div className="lg:col-span-4 bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3 space-y-2 h-[640px] overflow-y-auto">
          <span className="text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)] px-2 block">
            Endpoints
          </span>
          <div className="space-y-1">
            {ENDPOINTS.map((ep, idx) => {
              const isSelected = selectedEndpoint.path === ep.path && selectedEndpoint.method === ep.method
              return (
                <button
                  key={idx}
                  type="button"
                  onClick={() => setSelectedEndpoint(ep)}
                  className={`w-full text-left p-2 rounded-[5px] transition-colors flex items-center gap-2 text-[12px] ${
                    isSelected
                      ? 'bg-[var(--brand-subtle)] text-[var(--text-primary)] font-semibold border-l-2 border-[var(--brand-primary)]'
                      : 'text-[var(--text-secondary)] hover:bg-[var(--bg-panel)]'
                  }`}
                >
                  <span
                    className={`font-mono text-[10px] font-bold px-1.5 py-0.2 rounded ${
                      ep.method === 'GET'
                        ? 'text-emerald-400 bg-emerald-500/10'
                        : ep.method === 'POST'
                        ? 'text-blue-400 bg-blue-500/10'
                        : 'text-amber-400 bg-amber-500/10'
                    }`}
                  >
                    {ep.method}
                  </span>
                  <span className="font-mono text-[11px] truncate flex-1">{ep.path}</span>
                </button>
              )
            })}
          </div>
        </div>

        <div className="lg:col-span-8 space-y-4">
          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-4">
            <div className="flex items-center justify-between pb-3 border-b border-[var(--border-subtle)]">
              <div className="flex items-center gap-2.5">
                <span
                  className={`font-mono text-[12px] font-bold px-2 py-0.5 rounded border ${
                    selectedEndpoint.method === 'GET'
                      ? 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
                      : selectedEndpoint.method === 'POST'
                      ? 'text-blue-400 bg-blue-500/10 border-blue-500/20'
                      : 'text-amber-400 bg-amber-500/10 border-amber-500/20'
                  }`}
                >
                  {selectedEndpoint.method}
                </span>
                <span className="font-mono text-[13.5px] font-semibold text-[var(--text-primary)]">
                  {selectedEndpoint.path}
                </span>
              </div>
              <span className="text-[10.5px] font-mono text-[var(--text-muted)] bg-[var(--bg-panel)] px-2 py-0.5 rounded border border-[var(--border-subtle)]">
                {selectedEndpoint.tag}
              </span>
            </div>

            <p className="text-[12px] text-[var(--text-secondary)]">
              {selectedEndpoint.description}
            </p>

            {selectedEndpoint.requestBody && (
              <div className="space-y-1.5">
                <span className="text-[11px] font-semibold uppercase text-[var(--text-muted)] block">
                  Request Payload (JSON)
                </span>
                <pre className="bg-[#07090E] p-3 rounded-[6px] border border-[var(--border-subtle)] text-[11px] font-mono text-sky-300 overflow-x-auto leading-relaxed select-all">
                  {selectedEndpoint.requestBody}
                </pre>
              </div>
            )}

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-semibold uppercase text-[var(--text-muted)]">
                  Response (HTTP {selectedEndpoint.responseStatus})
                </span>
              </div>
              <pre className="bg-[#07090E] p-3 rounded-[6px] border border-[var(--border-subtle)] text-[11px] font-mono text-emerald-400 overflow-x-auto leading-relaxed select-all">
                {selectedEndpoint.responseSample}
              </pre>
            </div>
          </div>

          <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-4 space-y-2.5">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Terminal className="w-4 h-4 text-[var(--brand-text)]" />
                <span className="text-[13px] font-semibold text-[var(--text-primary)]">
                  Generated cURL Command
                </span>
              </div>
              <button
                type="button"
                onClick={handleCopy}
                className="text-[11px] text-[var(--brand-text)] hover:underline inline-flex items-center gap-1 font-medium"
              >
                {copied ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-[var(--status-success)]" />
                    <span>Copied</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>Copy cURL</span>
                  </>
                )}
              </button>
            </div>

            <pre className="bg-[#07090E] p-3 rounded-[6px] border border-[var(--border-subtle)] text-[11px] font-mono text-[var(--text-secondary)] overflow-x-auto leading-relaxed select-all">
              {curl}
            </pre>
          </div>
        </div>
      </div>
    </div>
  )
}
