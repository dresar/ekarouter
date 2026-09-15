import { useState, useEffect, useRef, FormEvent } from 'react'
import {
  Sparkles,
  Send,
  Globe,
  Server,
  Copy,
  Check,
  Clock,
  CheckCircle2,
  XCircle,
  Code,
  Eye,
  EyeOff,
  Layers,
  StopCircle,
} from 'lucide-react'
import { api } from '../../api/client.ts'
import { Account } from '../../types/api.ts'

interface ProviderPreset {
  id: string
  name: string
  defaultEndpoint: string
  defaultModels: string[]
  format: 'gemini' | 'openai'
}

const PROVIDER_PRESETS: ProviderPreset[] = [
  {
    id: 'gemini',
    name: 'Google Gemini',
    defaultEndpoint: 'https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent',
    defaultModels: [
      'gemini-2.5-flash',
      'gemini-2.5-pro',
      'gemini-3.6-flash',
      'gemini-2.5-flash-lite',
      'gemini-2.0-flash',
    ],
    format: 'gemini',
  },
  {
    id: 'openai',
    name: 'OpenAI',
    defaultEndpoint: 'https://api.openai.com/v1/chat/completions',
    defaultModels: ['gpt-4o-mini', 'gpt-4o', 'gpt-4.1-mini', 'o3-mini'],
    format: 'openai',
  },
  {
    id: 'groq',
    name: 'Groq',
    defaultEndpoint: 'https://api.groq.com/openai/v1/chat/completions',
    defaultModels: ['llama-3.3-70b-versatile', 'llama-3.1-8b-instant', 'mixtral-8x7b-32768'],
    format: 'openai',
  },
  {
    id: 'openrouter',
    name: 'OpenRouter',
    defaultEndpoint: 'https://openrouter.ai/api/v1/chat/completions',
    defaultModels: ['google/gemini-2.5-flash', 'deepseek/deepseek-chat', 'meta-llama/llama-3.3-70b-instruct'],
    format: 'openai',
  },
  {
    id: 'custom',
    name: 'Custom / Proxy Relay',
    defaultEndpoint: 'https://api.openai.com/v1/chat/completions',
    defaultModels: ['custom-model'],
    format: 'openai',
  },
]

const PROMPT_PRESETS = [
  { label: 'Health Check', prompt: 'Halo! Sebutkan nama modelmu dan berikan 1 kalimat pembuka.' },
  { label: 'Kecepatan & Latency', prompt: 'Hitung hasil dari: 47 dikali 83 = ?' },
  { label: 'Kreativitas', prompt: 'Tulis 2 baris sajak singkat bertema kecerdasan buatan dan router.' },
]

export function PlaygroundPage() {
  const [testMode, setTestMode] = useState<'browser' | 'gateway'>('browser')

  const [selectedPresetId, setSelectedPresetId] = useState('gemini')
  const [apiKey, setApiKey] = useState('')
  const [showApiKey, setShowApiKey] = useState(false)
  const [selectedModel, setSelectedModel] = useState('gemini-2.5-flash')
  const [customModel, setCustomModel] = useState('')
  const [customEndpoint, setCustomEndpoint] = useState('')
  const [temperature, setTemperature] = useState(0.7)
  const [maxTokens, setMaxTokens] = useState(1024)

  const [gatewayModel, setGatewayModel] = useState('')
  const [availableGatewayModels, setAvailableGatewayModels] = useState<string[]>([])
  const [routerApiKey, setRouterApiKey] = useState('')

  const [savedAccounts, setSavedAccounts] = useState<Account[]>([])

  const [prompt, setPrompt] = useState('Halo! Sebutkan nama modelmu dan berikan 1 kalimat pembuka.')
  const [isLoading, setIsLoading] = useState(false)
  const [latency, setLatency] = useState<number | null>(null)
  const [responseStatus, setResponseStatus] = useState<number | null>(null)
  const [responseText, setResponseText] = useState('')
  const [errorDetails, setErrorDetails] = useState<string | null>(null)
  const [rawRequest, setRawRequest] = useState<any>(null)
  const [rawResponse, setRawResponse] = useState<any>(null)
  const [showRawInspector, setShowRawInspector] = useState(false)
  const [isCopied, setIsCopied] = useState(false)

  const abortControllerRef = useRef<AbortController | null>(null)

  const activePreset = PROVIDER_PRESETS.find((p) => p.id === selectedPresetId) || PROVIDER_PRESETS[0]
  const effectiveModel = customModel.trim() || selectedModel

  useEffect(() => {
    loadSavedAccounts()
    loadGatewayModels()
  }, [])

  useEffect(() => {
    if (activePreset.defaultModels.length > 0 && !activePreset.defaultModels.includes(selectedModel)) {
      setSelectedModel(activePreset.defaultModels[0])
    }
  }, [selectedPresetId])

  const loadSavedAccounts = async () => {
    try {
      const [accsRes, keysRes] = await Promise.allSettled([
        api.get<Account[]>('/api/accounts'),
        api.get<any[]>('/api/keys'),
      ])

      if (accsRes.status === 'fulfilled' && Array.isArray(accsRes.value)) {
        setSavedAccounts(accsRes.value)
      }
      if (keysRes.status === 'fulfilled' && Array.isArray(keysRes.value) && keysRes.value.length > 0) {
        setRouterApiKey(keysRes.value[0].key || '')
      }
    } catch {}
  }

  const loadGatewayModels = async () => {
    try {
      const [modelsRes, routesRes] = await Promise.allSettled([
        api.get<any[]>('/api/models'),
        api.get<any[]>('/api/routes'),
      ])

      const list: string[] = []
      if (modelsRes.status === 'fulfilled' && Array.isArray(modelsRes.value)) {
        modelsRes.value.filter((m) => m.enabled).forEach((m) => {
          if (m.external_name && !list.includes(m.external_name)) list.push(m.external_name)
          if (m.id && !list.includes(m.id)) list.push(m.id)
        })
      }
      if (routesRes.status === 'fulfilled' && Array.isArray(routesRes.value)) {
        routesRes.value.filter((r) => r.enabled).forEach((r) => {
          if (r.name && !list.includes(r.name)) list.push(r.name)
        })
      }

      if (list.length > 0) {
        setAvailableGatewayModels(list)
        setGatewayModel(list[0])
      } else {
        setAvailableGatewayModels(['gemini-2.5-flash', 'gemini-3.6-flash', 'auto'])
        setGatewayModel('gemini-2.5-flash')
      }
    } catch {
      setAvailableGatewayModels(['gemini-2.5-flash', 'auto'])
      setGatewayModel('gemini-2.5-flash')
    }
  }

  const handleSelectSavedAccount = (accId: string) => {
    const found = savedAccounts.find((a) => a.id === accId)
    if (!found) return
    if (found.provider_id.includes('gemini')) {
      setSelectedPresetId('gemini')
    } else if (found.provider_id.includes('groq')) {
      setSelectedPresetId('groq')
    } else if (found.provider_id.includes('openrouter')) {
      setSelectedPresetId('openrouter')
    } else if (found.provider_id.includes('openai')) {
      setSelectedPresetId('openai')
    }
    if (found.proxy_url) {
      setCustomEndpoint(found.proxy_url)
    }
  }

  const handleStop = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = null
      setIsLoading(false)
    }
  }

  const handleExecuteTest = async (e?: FormEvent) => {
    if (e) e.preventDefault()
    if (!prompt.trim()) return

    setIsLoading(true)
    setLatency(null)
    setResponseStatus(null)
    setResponseText('')
    setErrorDetails(null)
    setRawRequest(null)
    setRawResponse(null)

    const controller = new AbortController()
    abortControllerRef.current = controller
    const startTime = performance.now()

    try {
      if (testMode === 'browser') {
        if (!apiKey.trim() && selectedPresetId !== 'custom') {
          throw new Error('Masukkan API Key untuk pengujian langsung dari browser Chrome')
        }

        if (activePreset.format === 'gemini') {
          let url = customEndpoint.trim()
            ? customEndpoint.trim()
            : `https://generativelanguage.googleapis.com/v1beta/models/${encodeURIComponent(effectiveModel)}:generateContent?key=${encodeURIComponent(apiKey.trim())}`

          if (customEndpoint.trim() && !url.includes('key=') && apiKey.trim()) {
            url += (url.includes('?') ? '&' : '?') + `key=${encodeURIComponent(apiKey.trim())}`
          }

          const reqBody = {
            contents: [{ parts: [{ text: prompt }] }],
            generationConfig: {
              temperature,
              maxOutputTokens: maxTokens,
            },
          }

          setRawRequest({ url, method: 'POST', headers: { 'Content-Type': 'application/json' }, body: reqBody })

          const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(reqBody),
            signal: controller.signal,
          })

          const elapsed = Math.round(performance.now() - startTime)
          setLatency(elapsed)
          setResponseStatus(res.status)

          const data = await res.json()
          setRawResponse(data)

          if (!res.ok) {
            const errMessage = data?.error?.message || `HTTP ${res.status}: ${res.statusText}`
            throw new Error(errMessage)
          }

          const reply = data?.candidates?.[0]?.content?.parts?.[0]?.text || JSON.stringify(data, null, 2)
          setResponseText(reply)
        } else {
          const url = customEndpoint.trim() || activePreset.defaultEndpoint
          const reqHeaders: Record<string, string> = {
            'Content-Type': 'application/json',
          }
          if (apiKey.trim()) {
            reqHeaders['Authorization'] = `Bearer ${apiKey.trim()}`
          }

          const reqBody = {
            model: effectiveModel,
            messages: [{ role: 'user', content: prompt }],
            temperature,
            max_tokens: maxTokens,
          }

          setRawRequest({ url, method: 'POST', headers: reqHeaders, body: reqBody })

          const res = await fetch(url, {
            method: 'POST',
            headers: reqHeaders,
            body: JSON.stringify(reqBody),
            signal: controller.signal,
          })

          const elapsed = Math.round(performance.now() - startTime)
          setLatency(elapsed)
          setResponseStatus(res.status)

          const data = await res.json()
          setRawResponse(data)

          if (!res.ok) {
            const errMessage = data?.error?.message || data?.message || `HTTP ${res.status}: ${res.statusText}`
            throw new Error(errMessage)
          }

          const reply = data?.choices?.[0]?.message?.content || JSON.stringify(data, null, 2)
          setResponseText(reply)
        }
      } else {
        const url = '/v1/chat/completions'
        const reqHeaders: Record<string, string> = {
          'Content-Type': 'application/json',
        }
        if (routerApiKey) {
          reqHeaders['Authorization'] = `Bearer ${routerApiKey}`
        }

        const reqBody = {
          model: gatewayModel || 'auto',
          messages: [{ role: 'user', content: prompt }],
          temperature,
          max_tokens: maxTokens,
        }

        setRawRequest({ url, method: 'POST', headers: reqHeaders, body: reqBody })

        const res = await fetch(url, {
          method: 'POST',
          headers: reqHeaders,
          body: JSON.stringify(reqBody),
          signal: controller.signal,
        })

        const elapsed = Math.round(performance.now() - startTime)
        setLatency(elapsed)
        setResponseStatus(res.status)

        const data = await res.json()
        setRawResponse(data)

        if (!res.ok) {
          const errMessage = data?.error?.message || data?.error || `HTTP ${res.status}`
          throw new Error(errMessage)
        }

        const reply = data?.choices?.[0]?.message?.content || JSON.stringify(data, null, 2)
        setResponseText(reply)
      }
    } catch (err: unknown) {
      const elapsed = Math.round(performance.now() - startTime)
      setLatency(elapsed)
      if (err instanceof Error && err.name === 'AbortError') {
        setErrorDetails('Pengujian dibatalkan oleh pengguna.')
      } else {
        const msg = err instanceof Error ? err.message : 'Permintaan gagal dieksekusi'
        setErrorDetails(msg)
      }
    } finally {
      setIsLoading(false)
      abortControllerRef.current = null
    }
  }

  const copyResponse = () => {
    if (!responseText) return
    navigator.clipboard.writeText(responseText)
    setIsCopied(true)
    setTimeout(() => setIsCopied(false), 2000)
  }

  return (
    <div className="space-y-4 pb-24">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] p-4">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-[8px] bg-[#2a1d17] border border-[#ea580c]/30 flex items-center justify-center text-[#f97316] shrink-0">
            <Sparkles className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-[17px] font-bold text-[var(--text-primary)]">AI Playground & Key Tester</h2>
            <p className="text-[12px] text-[var(--text-muted)]">
              {testMode === 'browser'
                ? 'Mode Client Chrome Saja — 100% permintaan dikirim langsung dari browser ke AI provider tanpa lewat backend.'
                : 'Mode EkaRouter Gateway — Menguji combo routing, fallback multi-model, dan token saver melalui server gateway.'}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-1.5 p-1 rounded-[8px] bg-[#14151b] border border-[#2b2f3a] shrink-0 self-start sm:self-auto">
          <button
            type="button"
            onClick={() => setTestMode('browser')}
            className={`px-3 py-1.5 text-[12px] font-medium rounded-[6px] transition-all flex items-center gap-1.5 cursor-pointer ${
              testMode === 'browser'
                ? 'bg-[#ea580c] text-white shadow-sm'
                : 'text-[#9ca3af] hover:text-white hover:bg-[#1e2027]'
            }`}
          >
            <Globe className="w-3.5 h-3.5" />
            <span>Browser Direct (Chrome Saja)</span>
          </button>
          <button
            type="button"
            onClick={() => setTestMode('gateway')}
            className={`px-3 py-1.5 text-[12px] font-medium rounded-[6px] transition-all flex items-center gap-1.5 cursor-pointer ${
              testMode === 'gateway'
                ? 'bg-[#ea580c] text-white shadow-sm'
                : 'text-[#9ca3af] hover:text-white hover:bg-[#1e2027]'
            }`}
          >
            <Server className="w-3.5 h-3.5" />
            <span>EkaRouter Gateway</span>
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 items-start">
        <div className="lg:col-span-5 space-y-4">
          <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] p-4 space-y-4">
            <h3 className="text-[13.5px] font-bold text-[var(--text-primary)] flex items-center gap-2 pb-2 border-b border-[var(--border-subtle)]">
              <Layers className="w-4 h-4 text-[#ea580c]" />
              <span>Konfigurasi Pengujian</span>
            </h3>

            {testMode === 'browser' ? (
              <div className="space-y-3.5">
                <div>
                  <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                    Target AI Provider
                  </label>
                  <div className="grid grid-cols-2 gap-1.5">
                    {PROVIDER_PRESETS.map((p) => (
                      <button
                        key={p.id}
                        type="button"
                        onClick={() => setSelectedPresetId(p.id)}
                        className={`px-2.5 py-2 text-[12px] font-medium rounded-[6px] border text-left transition-colors cursor-pointer ${
                          selectedPresetId === p.id
                            ? 'bg-[#2a1d17] border-[#ea580c]/50 text-[#f97316]'
                            : 'bg-[#181920] border-[#2c303d] text-[#a0a4b5] hover:bg-[#20222a] hover:text-white'
                        }`}
                      >
                        {p.name}
                      </button>
                    ))}
                  </div>
                </div>

                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)]">
                      API Key (Diuji langsung dari Chrome) *
                    </label>
                    {savedAccounts.length > 0 && (
                      <div className="relative inline-block">
                        <select
                          onChange={(e) => handleSelectSavedAccount(e.target.value)}
                          defaultValue=""
                          className="text-[10.5px] font-medium text-[#ea580c] bg-transparent hover:underline focus:outline-none cursor-pointer"
                        >
                          <option value="" disabled>
                            Pilih Profil Tersimpan
                          </option>
                          {savedAccounts.map((a) => (
                            <option key={a.id} value={a.id} className="bg-[#181920] text-white">
                              {a.name} ({a.provider_id})
                            </option>
                          ))}
                        </select>
                      </div>
                    )}
                  </div>
                  <div className="relative">
                    <input
                      type={showApiKey ? 'text' : 'password'}
                      value={apiKey}
                      onChange={(e) => setApiKey(e.target.value)}
                      placeholder={selectedPresetId === 'gemini' ? 'AIzaSy...' : 'sk-...'}
                      className="w-full h-9 pl-3 pr-9 text-[13px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                    />
                    <button
                      type="button"
                      onClick={() => setShowApiKey(!showApiKey)}
                      className="absolute inset-y-0 right-0 pr-2.5 flex items-center text-[#8e93a6] hover:text-white cursor-pointer"
                    >
                      {showApiKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                  <p className="text-[10.5px] text-[#787d90] mt-1">
                    Key tidak dikirim ke backend EkaRouter. Permintaan langsung dari browser ke server provider.
                  </p>
                </div>

                <div>
                  <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                    Pilih Model
                  </label>
                  <select
                    value={selectedModel}
                    onChange={(e) => {
                      setSelectedModel(e.target.value)
                      setCustomModel('')
                    }}
                    className="w-full h-9 px-3 text-[13px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                  >
                    {activePreset.defaultModels.map((m) => (
                      <option key={m} value={m}>
                        {m}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                    Nama Model Kustom (Opsional)
                  </label>
                  <input
                    type="text"
                    value={customModel}
                    onChange={(e) => setCustomModel(e.target.value)}
                    placeholder="Contoh: gemini-3.6-flash atau gpt-4o"
                    className="w-full h-9 px-3 text-[12.5px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                  />
                </div>

                <div>
                  <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                    Custom Proxy / Base URL (Opsional)
                  </label>
                  <input
                    type="text"
                    value={customEndpoint}
                    onChange={(e) => setCustomEndpoint(e.target.value)}
                    placeholder="Contoh: https://sour-mussel-7365.deno.net"
                    className="w-full h-9 px-3 text-[12px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                  />
                </div>
              </div>
            ) : (
              <div className="space-y-3.5">
                <div>
                  <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                    Pilih Gateway Route / Model
                  </label>
                  <select
                    value={gatewayModel}
                    onChange={(e) => setGatewayModel(e.target.value)}
                    className="w-full h-9 px-3 text-[13px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                  >
                    {availableGatewayModels.map((m) => (
                      <option key={m} value={m}>
                        {m}
                      </option>
                    ))}
                  </select>
                  <p className="text-[11px] text-[var(--text-muted)] mt-1">
                    Permintaan akan melewati load balancer EkaRouter dengan failover otomatis ke akun cadangan.
                  </p>
                </div>
              </div>
            )}

            <div className="pt-2 border-t border-[var(--border-subtle)] space-y-3">
              <div>
                <div className="flex items-center justify-between text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1">
                  <span>Temperature</span>
                  <span className="font-mono text-[#ea580c]">{temperature}</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.1"
                  value={temperature}
                  onChange={(e) => setTemperature(parseFloat(e.target.value))}
                  className="w-full accent-[#ea580c] cursor-pointer"
                />
              </div>

              <div>
                <div className="flex items-center justify-between text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1">
                  <span>Max Tokens</span>
                  <span className="font-mono text-[#ea580c]">{maxTokens}</span>
                </div>
                <input
                  type="number"
                  min="64"
                  max="8192"
                  step="64"
                  value={maxTokens}
                  onChange={(e) => setMaxTokens(parseInt(e.target.value) || 1024)}
                  className="w-full h-8 px-3 text-[12px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c]"
                />
              </div>
            </div>
          </div>
        </div>

        <div className="lg:col-span-7 space-y-4">
          <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] p-4 space-y-3">
            <div className="flex items-center justify-between gap-2 flex-wrap pb-1">
              <span className="text-[12px] font-semibold text-[var(--text-secondary)]">
                Preset Pertanyaan Cepat:
              </span>
              <div className="flex items-center gap-1.5 flex-wrap">
                {PROMPT_PRESETS.map((p) => (
                  <button
                    key={p.label}
                    type="button"
                    onClick={() => setPrompt(p.prompt)}
                    className="px-2 py-1 text-[11px] font-medium rounded-[5px] bg-[#1a1b22] hover:bg-[#232530] border border-[#2d313e] text-[#b0b4c5] hover:text-[#ea580c] transition-colors cursor-pointer"
                  >
                    {p.label}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <textarea
                rows={4}
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                onKeyDown={(e) => {
                  if (e.ctrlKey && e.key === 'Enter') {
                    handleExecuteTest()
                  }
                }}
                placeholder="Tulis instruksi atau pertanyaan untuk menguji model... (Tekan Ctrl+Enter untuk kirim)"
                className="w-full p-3 text-[13px] rounded-[7px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors leading-relaxed"
              />
            </div>

            <div className="flex items-center justify-between gap-3 pt-1">
              <div className="text-[11px] text-[var(--text-muted)] hidden sm:block">
                <span>Tekan </span>
                <kbd className="px-1.5 py-0.5 rounded bg-[#1f2129] border border-[#363a48] font-mono text-[10px]">
                  Ctrl + Enter
                </kbd>
                <span> untuk menjalankan tes</span>
              </div>

              <div className="flex items-center gap-2 ml-auto">
                {isLoading ? (
                  <button
                    type="button"
                    onClick={handleStop}
                    className="h-8 px-4 text-[12px] font-semibold rounded-[6px] bg-rose-950/40 border border-rose-600/50 text-rose-300 hover:bg-rose-900/50 flex items-center gap-1.5 transition-colors cursor-pointer"
                  >
                    <StopCircle className="w-3.5 h-3.5" />
                    <span>Stop</span>
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={() => handleExecuteTest()}
                    className="h-8 px-5 text-[12.5px] font-semibold rounded-[6px] bg-[#ea580c] hover:bg-[#f97316] text-white flex items-center gap-1.5 transition-colors cursor-pointer shadow-sm"
                  >
                    <Send className="w-3.5 h-3.5" />
                    <span>Run Test</span>
                  </button>
                )}
              </div>
            </div>
          </div>

          <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] overflow-hidden">
            <div className="flex items-center justify-between px-4 py-2.5 bg-[#14151b] border-b border-[var(--border-subtle)]">
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-[12.5px] font-bold text-[var(--text-primary)]">Hasil Respon</span>

                {responseStatus !== null && (
                  <span
                    className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10.5px] font-mono font-medium ${
                      responseStatus >= 200 && responseStatus < 300
                        ? 'bg-emerald-950/40 text-emerald-400 border border-emerald-600/30'
                        : 'bg-rose-950/40 text-rose-400 border border-rose-600/30'
                    }`}
                  >
                    {responseStatus >= 200 && responseStatus < 300 ? (
                      <CheckCircle2 className="w-3 h-3" />
                    ) : (
                      <XCircle className="w-3 h-3" />
                    )}
                    HTTP {responseStatus}
                  </span>
                )}

                {latency !== null && (
                  <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-[4px] bg-[#1a1b20] border border-[#2c303d] text-[11px] font-mono text-[#9ca3af]">
                    <Clock className="w-3 h-3 text-[#ea580c]" />
                    {latency}ms
                  </span>
                )}

                <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[4px] bg-[#1a1b20] text-[#717686]">
                  {testMode === 'browser' ? 'Direct Chrome' : 'Gateway'}
                </span>
              </div>

              <div className="flex items-center gap-1.5">
                {responseText && (
                  <button
                    type="button"
                    onClick={copyResponse}
                    className="h-7 px-2 text-[11px] font-medium rounded-[5px] bg-[#1e2027] hover:bg-[#282a35] border border-[#353947] text-[#c0c4d4] flex items-center gap-1 transition-colors cursor-pointer"
                  >
                    {isCopied ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                    <span>{isCopied ? 'Tersalin' : 'Copy'}</span>
                  </button>
                )}

                <button
                  type="button"
                  onClick={() => setShowRawInspector(!showRawInspector)}
                  className={`h-7 px-2 text-[11px] font-medium rounded-[5px] border flex items-center gap-1 transition-colors cursor-pointer ${
                    showRawInspector
                      ? 'bg-[#2a1d17] border-[#ea580c]/50 text-[#f97316]'
                      : 'bg-[#1e2027] hover:bg-[#282a35] border-[#353947] text-[#c0c4d4]'
                  }`}
                >
                  <Code className="w-3 h-3" />
                  <span>Inspect JSON</span>
                </button>
              </div>
            </div>

            <div className="p-4 min-h-[160px]">
              {isLoading ? (
                <div className="py-12 flex flex-col items-center justify-center gap-2 text-[var(--text-muted)]">
                  <div className="w-6 h-6 rounded-full border-2 border-[#383c4b] border-t-[#ea580c] animate-spin" />
                  <span className="text-[12px] font-mono">Mengirim permintaan langsung...</span>
                </div>
              ) : errorDetails ? (
                <div className="p-3.5 rounded-[7px] bg-rose-950/20 border border-rose-600/30 text-rose-300 text-[12px] space-y-1 font-mono">
                  <div className="font-bold flex items-center gap-2 text-rose-200">
                    <XCircle className="w-4 h-4 shrink-0" />
                    <span>Error Permintaan</span>
                  </div>
                  <p className="whitespace-pre-wrap break-all text-[11.5px] leading-relaxed pt-1">
                    {errorDetails}
                  </p>
                </div>
              ) : responseText ? (
                <div className="text-[13px] text-[#e0e2eb] whitespace-pre-wrap leading-relaxed font-sans">
                  {responseText}
                </div>
              ) : (
                <div className="py-12 text-center text-[#686d80] text-[12px]">
                  Belum ada respon. Masukkan API key & klik <strong>Run Test</strong> untuk memulai pengujian.
                </div>
              )}
            </div>

            {showRawInspector && (rawRequest || rawResponse) && (
              <div className="border-t border-[var(--border-subtle)] bg-[#101115] p-4 space-y-3">
                {rawRequest && (
                  <div>
                    <span className="text-[10.5px] font-semibold uppercase text-[#787d90] tracking-wider block mb-1">
                      Raw Request Payload
                    </span>
                    <pre className="p-2.5 rounded-[6px] bg-[#16171d] border border-[#272a34] text-[11px] font-mono text-[#a5abbf] overflow-x-auto max-h-48 overflow-y-auto">
                      {JSON.stringify(rawRequest, null, 2)}
                    </pre>
                  </div>
                )}

                {rawResponse && (
                  <div>
                    <span className="text-[10.5px] font-semibold uppercase text-[#787d90] tracking-wider block mb-1">
                      Raw Response JSON
                    </span>
                    <pre className="p-2.5 rounded-[6px] bg-[#16171d] border border-[#272a34] text-[11px] font-mono text-[#a5abbf] overflow-x-auto max-h-56 overflow-y-auto">
                      {JSON.stringify(rawResponse, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
