import { useState, useEffect, useRef, useMemo, FormEvent } from 'react'
import {
  Sparkles,
  Send,
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
  RefreshCw,
  Key,
  Zap,
} from 'lucide-react'
import { api } from '../../api/client.ts'
import { Account } from '../../types/api.ts'
import { SearchableSelect, SearchableOption } from '../../components/ui/SearchableSelect.tsx'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'

interface ProviderPreset {
  id: string
  name: string
  category: string
  defaultEndpoint: string
  defaultModels: string[]
  format: 'gemini' | 'openai' | 'anthropic'
  keyPlaceholder: string
  description: string
}

const AI_PROVIDER_PRESETS: ProviderPreset[] = [
  {
    id: 'gemini',
    name: 'Google Gemini',
    category: 'Provider Utama',
    defaultEndpoint: 'https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent',
    defaultModels: [
      'gemini-2.5-flash',
      'gemini-2.5-pro',
      'gemini-3.6-flash',
      'gemini-2.5-flash-lite',
      'gemini-2.0-flash',
    ],
    format: 'gemini',
    keyPlaceholder: 'AIzaSy...',
    description: 'Direct Google GenAI v1beta API',
  },
  {
    id: 'gemini-cli',
    name: 'Google Gemini CLI / Antigravity',
    category: 'Provider Utama',
    defaultEndpoint: 'https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent',
    defaultModels: [
      'gemini-2.5-flash',
      'gemini-2.5-pro',
      'gemini-3.6-flash',
    ],
    format: 'gemini',
    keyPlaceholder: 'AIzaSy... / OAuth Token',
    description: 'Gemini CLI & Antigravity Client',
  },
  {
    id: 'openai',
    name: 'OpenAI',
    category: 'Provider Utama',
    defaultEndpoint: 'https://api.openai.com/v1/chat/completions',
    defaultModels: ['gpt-4o-mini', 'gpt-4o', 'gpt-4.1-mini', 'o3-mini', 'o1'],
    format: 'openai',
    keyPlaceholder: 'sk-proj-...',
    description: 'Official OpenAI Chat Completions',
  },
  {
    id: 'anthropic',
    name: 'Anthropic Claude',
    category: 'Provider Utama',
    defaultEndpoint: 'https://api.anthropic.com/v1/messages',
    defaultModels: [
      'claude-3-7-sonnet-20250219',
      'claude-3-5-sonnet-20241022',
      'claude-3-5-haiku-20241022',
    ],
    format: 'anthropic',
    keyPlaceholder: 'sk-ant-...',
    description: 'Anthropic Messages API',
  },
  {
    id: 'deepseek',
    name: 'DeepSeek Official',
    category: 'Provider Utama',
    defaultEndpoint: 'https://api.deepseek.com/chat/completions',
    defaultModels: ['deepseek-chat', 'deepseek-reasoner'],
    format: 'openai',
    keyPlaceholder: 'sk-...',
    description: 'DeepSeek V3 & DeepSeek R1 API',
  },
  {
    id: 'groq',
    name: 'Groq Cloud',
    category: 'Ultra-Fast Inference (LPU)',
    defaultEndpoint: 'https://api.groq.com/openai/v1/chat/completions',
    defaultModels: [
      'llama-3.3-70b-versatile',
      'llama-3.1-8b-instant',
      'mixtral-8x7b-32768',
      'gemma2-9b-it',
    ],
    format: 'openai',
    keyPlaceholder: 'gsk_...',
    description: 'Groq Tensor Streaming Processor',
  },
  {
    id: 'cerebras',
    name: 'Cerebras Inference',
    category: 'Ultra-Fast Inference (LPU)',
    defaultEndpoint: 'https://api.cerebras.ai/v1/chat/completions',
    defaultModels: ['llama3.1-8b', 'llama3.1-70b', 'llama-3.3-70b'],
    format: 'openai',
    keyPlaceholder: 'csk-...',
    description: 'Wafer-scale ultrafast token generation',
  },
  {
    id: 'chutes',
    name: 'Chutes AI',
    category: 'Ultra-Fast Inference (LPU)',
    defaultEndpoint: 'https://chutes.ai/api/v1/chat/completions',
    defaultModels: ['deepseek-ai/DeepSeek-V3', 'deepseek-ai/DeepSeek-R1', 'chutes-llama-3.3-70b'],
    format: 'openai',
    keyPlaceholder: 'cpk_...',
    description: 'Decentralized serverless GPU compute',
  },
  {
    id: 'openrouter',
    name: 'OpenRouter Aggregator',
    category: 'Aggregators & Router',
    defaultEndpoint: 'https://openrouter.ai/api/v1/chat/completions',
    defaultModels: [
      'google/gemini-2.5-flash',
      'deepseek/deepseek-chat',
      'deepseek/deepseek-r1',
      'meta-llama/llama-3.3-70b-instruct',
      'anthropic/claude-3.5-sonnet',
    ],
    format: 'openai',
    keyPlaceholder: 'sk-or-v1-...',
    description: 'Unified gateway ke 200+ models',
  },
  {
    id: 'together',
    name: 'Together AI',
    category: 'Aggregators & Router',
    defaultEndpoint: 'https://api.together.xyz/v1/chat/completions',
    defaultModels: [
      'meta-llama/Llama-3.3-70B-Instruct-Turbo',
      'deepseek-ai/DeepSeek-V3',
      'mistralai/Mixtral-8x7B-Instruct-v0.1',
    ],
    format: 'openai',
    keyPlaceholder: '...',
    description: 'Together AI Cloud inference',
  },
  {
    id: 'huggingface',
    name: 'Hugging Face Inference',
    category: 'Aggregators & Router',
    defaultEndpoint: 'https://api-inference.huggingface.co/v1/chat/completions',
    defaultModels: ['Qwen/Qwen2.5-72B-Instruct', 'meta-llama/Llama-3.3-70B-Instruct'],
    format: 'openai',
    keyPlaceholder: 'hf_...',
    description: 'Serverless Hugging Face Inference API',
  },
  {
    id: 'mistral',
    name: 'Mistral AI',
    category: 'Open-Weights & Reasoning',
    defaultEndpoint: 'https://api.mistral.ai/v1/chat/completions',
    defaultModels: ['mistral-large-latest', 'mistral-small-latest', 'codestral-latest'],
    format: 'openai',
    keyPlaceholder: '...',
    description: 'La Plateforme Mistral AI',
  },
  {
    id: 'nvidia',
    name: 'NVIDIA NIM',
    category: 'Open-Weights & Reasoning',
    defaultEndpoint: 'https://integrate.api.nvidia.com/v1/chat/completions',
    defaultModels: ['meta/llama-3.3-70b-instruct', 'deepseek-ai/deepseek-r1', 'mistralai/mistral-large-2-instruct'],
    format: 'openai',
    keyPlaceholder: 'nvapi-...',
    description: 'NVIDIA API Catalog microservices',
  },
  {
    id: 'qwen',
    name: 'Qwen Alibaba Cloud',
    category: 'Open-Weights & Reasoning',
    defaultEndpoint: 'https://dashscope-intl.aliyuncs.com/compatible-mode/v1/chat/completions',
    defaultModels: ['qwen-plus', 'qwen-turbo', 'qwen-max', 'qwen2.5-coder-32b-instruct'],
    format: 'openai',
    keyPlaceholder: 'sk-...',
    description: 'DashScope OpenAI-compatible API',
  },
  {
    id: 'xai',
    name: 'xAI (Grok)',
    category: 'Provider Utama',
    defaultEndpoint: 'https://api.x.ai/v1/chat/completions',
    defaultModels: ['grok-2-1212', 'grok-2-vision-1212', 'grok-beta'],
    format: 'openai',
    keyPlaceholder: 'xai-...',
    description: 'Elon Musk xAI Grok API',
  },
  {
    id: 'moonshot',
    name: 'Moonshot AI (Kimi)',
    category: 'Open-Weights & Reasoning',
    defaultEndpoint: 'https://api.moonshot.cn/v1/chat/completions',
    defaultModels: ['moonshot-v1-8k', 'moonshot-v1-32k', 'moonshot-v1-128k'],
    format: 'openai',
    keyPlaceholder: 'sk-...',
    description: 'Kimi Moonshot AI API',
  },
  {
    id: 'cloudflare',
    name: 'Cloudflare Workers AI',
    category: 'Edge & Local Cloud',
    defaultEndpoint: 'https://api.cloudflare.com/client/v4/accounts/{account_id}/ai/v1/chat/completions',
    defaultModels: ['@cf/meta/llama-3.3-70b-instruct', '@cf/deepseek-ai/deepseek-r1-distill-qwen-32b'],
    format: 'openai',
    keyPlaceholder: 'Bearer API Token',
    description: 'Serverless inference di edge Cloudflare',
  },
  {
    id: 'ollama',
    name: 'Ollama Local Host',
    category: 'Edge & Local Cloud',
    defaultEndpoint: 'http://localhost:11434/v1/chat/completions',
    defaultModels: ['llama3.2', 'deepseek-r1', 'qwen2.5-coder', 'mistral', 'phi4'],
    format: 'openai',
    keyPlaceholder: 'Bisa dikosongkan (Opsional)',
    description: 'Direct local LLM server di localhost:11434',
  },
  {
    id: 'github',
    name: 'GitHub Models',
    category: 'Developer & Free Tiers',
    defaultEndpoint: 'https://models.inference.ai.azure.com/chat/completions',
    defaultModels: ['gpt-4o', 'gpt-4o-mini', 'Phi-3.5-mini-instruct'],
    format: 'openai',
    keyPlaceholder: 'ghp_... (GitHub Token)',
    description: 'Azure AI backed GitHub Models catalog',
  },
  {
    id: 'airforce',
    name: 'Airforce AI',
    category: 'Developer & Free Tiers',
    defaultEndpoint: 'https://api.airforce/v1/chat/completions',
    defaultModels: ['llama-3.3-70b', 'deepseek-r1', 'chatgpt-4o-latest'],
    format: 'openai',
    keyPlaceholder: 'sk-... / Token',
    description: 'Community AI model proxy',
  },
  {
    id: 'mimofree',
    name: 'MiMo Free',
    category: 'Developer & Free Tiers',
    defaultEndpoint: 'https://api.mimofree.com/v1/chat/completions',
    defaultModels: ['mimo-free-v1'],
    format: 'openai',
    keyPlaceholder: 'Token / Key',
    description: 'Xiaomi MiMo Free Tier Provider',
  },
  {
    id: 'devin',
    name: 'Devin AI',
    category: 'Developer & Free Tiers',
    defaultEndpoint: 'https://api.devin.ai/v1/chat/completions',
    defaultModels: ['devin-default'],
    format: 'openai',
    keyPlaceholder: 'devin_...',
    description: 'Devin AI Coding Agent proxy endpoint',
  },
  {
    id: 'custom',
    name: 'Custom / Proxy Relay',
    category: 'Custom & Self-Hosted',
    defaultEndpoint: 'https://api.openai.com/v1/chat/completions',
    defaultModels: ['custom-model'],
    format: 'openai',
    keyPlaceholder: 'sk-... (Opsional)',
    description: 'Custom proxy, Deno relay, atau server pribadi',
  },
]

const PROMPT_PRESETS = [
  { label: 'Health Check', prompt: 'Halo! Sebutkan nama modelmu dan berikan 1 kalimat pembuka.' },
  { label: 'Kecepatan & Latency', prompt: 'Hitung hasil dari: 47 dikali 83 = ?' },
  { label: 'Kreativitas', prompt: 'Tulis 2 baris sajak singkat bertema kecerdasan buatan dan router.' },
]

export function PlaygroundPage() {
  const [selectedPresetId, setSelectedPresetId] = useState('gemini')
  const [testMethod, setTestMethod] = useState<'pool_rotation' | 'manual_key'>('pool_rotation')
  const [poolScope, setPoolScope] = useState<'all_rotation' | 'single_account'>('all_rotation')
  const [selectedAccountId, setSelectedAccountId] = useState('')

  const [apiKey, setApiKey] = useState('')
  const [showApiKey, setShowApiKey] = useState(false)
  const [selectedModel, setSelectedModel] = useState('gemini-2.5-flash')
  const [customModel, setCustomModel] = useState('')
  const [customEndpoint, setCustomEndpoint] = useState('')
  const [temperature, setTemperature] = useState(0.7)
  const [maxTokens, setMaxTokens] = useState(1024)

  const [savedAccounts, setSavedAccounts] = useState<Account[]>([])
  const [routerApiKey, setRouterApiKey] = useState('')

  const [prompt, setPrompt] = useState('Halo! Sebutkan nama modelmu dan berikan 1 kalimat pembuka.')
  const [isLoading, setIsLoading] = useState(false)
  const [latency, setLatency] = useState<number | null>(null)
  const [responseStatus, setResponseStatus] = useState<number | null>(null)
  const [responseText, setResponseText] = useState('')
  const [errorDetails, setErrorDetails] = useState<string | null>(null)
  const [executedInfo, setExecutedInfo] = useState<string | null>(null)
  const [rawRequest, setRawRequest] = useState<any>(null)
  const [rawResponse, setRawResponse] = useState<any>(null)
  const [showRawInspector, setShowRawInspector] = useState(false)
  const [isCopied, setIsCopied] = useState(false)

  const abortControllerRef = useRef<AbortController | null>(null)

  const activePreset = useMemo(() => {
    return AI_PROVIDER_PRESETS.find((p) => p.id === selectedPresetId) || AI_PROVIDER_PRESETS[0]
  }, [selectedPresetId])

  const effectiveModel = customModel.trim() || selectedModel

  const selectedProviderAccounts = useMemo(() => {
    return savedAccounts.filter((a) => {
      const accProv = a.provider_id.toLowerCase().trim()
      const targetProv = selectedPresetId.toLowerCase().trim()
      if (targetProv === 'gemini' || targetProv === 'gemini-cli') {
        return accProv.includes('gemini') || accProv.includes('antigravity') || accProv.includes('google')
      }
      return accProv === targetProv || accProv.includes(targetProv) || targetProv.includes(accProv)
    })
  }, [savedAccounts, selectedPresetId])

  const activeAccountsCount = useMemo(() => {
    return selectedProviderAccounts.filter((a) => a.state === 'active' && a.enabled).length
  }, [selectedProviderAccounts])

  const providerOptions: SearchableOption[] = useMemo(() => {
    return AI_PROVIDER_PRESETS.map((p) => {
      const matched = savedAccounts.filter((a) => {
        const accProv = a.provider_id.toLowerCase().trim()
        const targetProv = p.id.toLowerCase().trim()
        if (targetProv === 'gemini' || targetProv === 'gemini-cli') {
          return accProv.includes('gemini') || accProv.includes('antigravity') || accProv.includes('google')
        }
        return accProv === targetProv || accProv.includes(targetProv) || targetProv.includes(accProv)
      })
      const count = matched.length
      const activeCount = matched.filter((a) => a.state === 'active' && a.enabled).length

      let sub = p.id
      if (count > 0) {
        sub = `${count} Akun (${activeCount} Aktif di Pool)`
      }

      return {
        value: p.id,
        label: p.name,
        sublabel: sub,
        category: p.category,
        icon: <ProviderLogo providerId={p.id} name={p.name} size="sm" />,
      }
    })
  }, [savedAccounts])

  const accountOptions: SearchableOption[] = useMemo(() => {
    return selectedProviderAccounts.map((a) => {
      const isCooling = a.state === 'cooling_down'
      const isDown = a.state === 'disabled' || a.state === 'unavailable'
      const stateLabel = isCooling ? 'Sedang Cooldown' : isDown ? 'Non-Aktif' : 'Profil Aktif'
      return {
        value: a.id,
        label: a.name,
        sublabel: `${a.proxy_name || a.proxy_url ? `${a.proxy_name || 'Proxy Aktif'} · ` : ''}${a.masked_secret || a.id.slice(0, 8)}`,
        category: stateLabel,
        icon: <ProviderLogo providerId={a.provider_id} name={a.name} size="sm" />,
      }
    })
  }, [selectedProviderAccounts])

  const modelOptions: SearchableOption[] = useMemo(() => {
    const models = activePreset.defaultModels || []
    return models.map((m) => ({
      value: m,
      label: m,
      sublabel: activePreset.id,
    }))
  }, [activePreset])

  useEffect(() => {
    loadSavedAccounts()
  }, [])

  useEffect(() => {
    if (activePreset.defaultModels.length > 0 && !activePreset.defaultModels.includes(selectedModel)) {
      setSelectedModel(activePreset.defaultModels[0])
    }
  }, [selectedPresetId, activePreset])

  useEffect(() => {
    if (selectedProviderAccounts.length > 0) {
      if (!selectedAccountId || !selectedProviderAccounts.some((a) => a.id === selectedAccountId)) {
        const firstActive = selectedProviderAccounts.find((a) => a.state === 'active') || selectedProviderAccounts[0]
        setSelectedAccountId(firstActive.id)
      }
    } else {
      setSelectedAccountId('')
    }
  }, [selectedPresetId, selectedProviderAccounts])

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
    setExecutedInfo(null)
    setRawRequest(null)
    setRawResponse(null)

    const controller = new AbortController()
    abortControllerRef.current = controller
    const startTime = performance.now()

    try {
      if (testMethod === 'pool_rotation') {
        if (poolScope === 'single_account' && selectedAccountId) {
          const selectedAcc = selectedProviderAccounts.find((a) => a.id === selectedAccountId)
          const url = `/api/accounts/${selectedAccountId}/test`
          setRawRequest({ url, method: 'POST', targetAccountId: selectedAccountId, accountName: selectedAcc?.name })

          const res = await api.post<{ healthy: boolean; latency: number; message: string }>(url, {})
          const elapsed = Math.round(performance.now() - startTime)
          setLatency(res.latency || elapsed)
          setResponseStatus(res.healthy ? 200 : 503)
          setRawResponse(res)

          setExecutedInfo(`Uji Langsung Akun: ${selectedAcc?.name || selectedAccountId} (${selectedAcc?.provider_id})`)

          if (res.healthy) {
            setResponseText(`Akun Sehat & Siap Digunakan!\nStatus: Terkoneksi ke upstream\nLatency: ${res.latency || elapsed}ms\nPesan: ${res.message || 'Respons OK'}`)
          } else {
            setErrorDetails(`Akun Mengalami Gangguan Upstream:\n${res.message || 'Gagal memvalidasi kredensial'}`)
          }
        } else {
          const token = localStorage.getItem('session_token') || routerApiKey
          const url = '/v1/chat/completions'
          const reqHeaders: Record<string, string> = {
            'Content-Type': 'application/json',
          }
          if (token) {
            reqHeaders['Authorization'] = `Bearer ${token}`
          }

          const targetModel = `${activePreset.id}/${effectiveModel}`
          const reqBody = {
            model: targetModel,
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

          setExecutedInfo(`Auto-Rotasi Pool: ${activePreset.name} (${selectedProviderAccounts.length} Akun di Pool)`)

          if (!res.ok) {
            const errMessage = data?.error?.message || data?.error || `HTTP ${res.status}: ${res.statusText}`
            throw new Error(errMessage)
          }

          const reply = data?.choices?.[0]?.message?.content || JSON.stringify(data, null, 2)
          setResponseText(reply)
        }
      } else {
        if (!apiKey.trim() && selectedPresetId !== 'custom' && selectedPresetId !== 'ollama') {
          throw new Error('Masukkan API Key untuk pengujian langsung dari browser Chrome')
        }

        setExecutedInfo(`Client Chrome Direct: ${activePreset.name}`)

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
        } else if (activePreset.format === 'anthropic') {
          const url = customEndpoint.trim() || activePreset.defaultEndpoint
          const reqHeaders: Record<string, string> = {
            'Content-Type': 'application/json',
            'x-api-key': apiKey.trim(),
            'anthropic-version': '2023-06-01',
            'anthropic-dangerous-direct-browser-access': 'true',
            'dangerously-allow-browser': 'true',
          }

          const reqBody = {
            model: effectiveModel,
            max_tokens: maxTokens,
            messages: [{ role: 'user', content: prompt }],
            temperature,
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
            const errMessage = data?.error?.message || `HTTP ${res.status}: ${res.statusText}`
            throw new Error(errMessage)
          }

          const reply = data?.content?.[0]?.text || JSON.stringify(data, null, 2)
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
            <h2 className="text-[17px] font-bold text-[var(--text-primary)]">AI Playground & Auto-Rotation Tester</h2>
            <p className="text-[12px] text-[var(--text-muted)]">
              Uji coba AI Provider dengan rotasi otomatis antar akun/key tersimpan di EkaRouter atau uji API key langsung dari browser Chrome.
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-[11px] font-mono px-2 py-1 rounded-[5px] bg-[#171821] border border-[#2b2e3b] text-[#9ca3af]">
            Total Akun Tersimpan: <strong className="text-[#ea580c]">{savedAccounts.length}</strong>
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 items-start">
        <div className="lg:col-span-5 space-y-4">
          <div className="bg-[var(--bg-surface)] border border-[var(--border-strong)] rounded-[10px] p-4 space-y-4">
            <h3 className="text-[13.5px] font-bold text-[var(--text-primary)] flex items-center gap-2 pb-2 border-b border-[var(--border-subtle)]">
              <Layers className="w-4 h-4 text-[#ea580c]" />
              <span>Konfigurasi Pengujian Provider</span>
            </h3>

            <div className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)]">
                    Target AI Provider ({AI_PROVIDER_PRESETS.length} Provider Tersedia)
                  </label>
                  <span className="text-[10px] font-mono font-semibold text-[#f97316] bg-[#2a1d17] px-1.5 py-0.5 rounded border border-[#ea580c]/30">
                    {activePreset.format.toUpperCase()}
                  </span>
                </div>
                <SearchableSelect
                  options={providerOptions}
                  value={selectedPresetId}
                  onChange={(val) => {
                    setSelectedPresetId(val)
                    setCustomModel('')
                  }}
                  placeholder="Pilih AI Provider..."
                  searchPlaceholder="Cari provider (gemini, claude, deepseek, groq, openai)..."
                />
                <p className="text-[10.5px] text-[#787d90] mt-1 truncate">
                  {activePreset.description}
                </p>
              </div>

              <div>
                <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                  Metode Pengujian untuk {activePreset.name}
                </label>
                <div className="grid grid-cols-2 gap-1.5 p-1 bg-[#121318] border border-[#2b2e3b] rounded-[7px]">
                  <button
                    type="button"
                    onClick={() => setTestMethod('pool_rotation')}
                    className={`px-2.5 py-1.5 text-[11.5px] font-medium rounded-[5px] flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
                      testMethod === 'pool_rotation'
                        ? 'bg-[#ea580c] text-white shadow-sm font-semibold'
                        : 'text-[#9ca3af] hover:text-white hover:bg-[#1c1d25]'
                    }`}
                  >
                    <RefreshCw className="w-3.5 h-3.5" />
                    <span>Auto-Rotasi Pool ({selectedProviderAccounts.length})</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => setTestMethod('manual_key')}
                    className={`px-2.5 py-1.5 text-[11.5px] font-medium rounded-[5px] flex items-center justify-center gap-1.5 transition-all cursor-pointer ${
                      testMethod === 'manual_key'
                        ? 'bg-[#ea580c] text-white shadow-sm font-semibold'
                        : 'text-[#9ca3af] hover:text-white hover:bg-[#1c1d25]'
                    }`}
                  >
                    <Key className="w-3.5 h-3.5" />
                    <span>API Key Langsung</span>
                  </button>
                </div>
              </div>

              {testMethod === 'pool_rotation' ? (
                <div className="p-3 rounded-[8px] bg-[#14151c] border border-[#262936] space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Zap className="w-4 h-4 text-[#ea580c]" />
                      <span className="text-[12px] font-bold text-[var(--text-primary)]">
                        Pool EkaRouter: {activePreset.name}
                      </span>
                    </div>
                    <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-[#1e2029] text-[#f97316] font-semibold border border-[#ea580c]/20">
                      {activeAccountsCount} Aktif / {selectedProviderAccounts.length} Total
                    </span>
                  </div>

                  {selectedProviderAccounts.length === 0 ? (
                    <div className="p-2.5 rounded-[6px] bg-[#1c1d25] text-[11px] text-[#9ca3af] leading-relaxed">
                      Belum ada akun tersimpan untuk <strong>{activePreset.name}</strong> di database EkaRouter. Anda dapat menambahkan akun di menu <strong>Providers</strong> atau gunakan tab <strong>API Key Langsung</strong> di atas.
                    </div>
                  ) : (
                    <div className="space-y-2.5">
                      <div className="flex items-center gap-2 pt-1">
                        <button
                          type="button"
                          onClick={() => setPoolScope('all_rotation')}
                          className={`flex-1 py-1.5 px-2 text-[11px] rounded-[5px] border text-center transition-colors cursor-pointer ${
                            poolScope === 'all_rotation'
                              ? 'bg-[#2a1d17] border-[#ea580c] text-[#f97316] font-medium'
                              : 'bg-[#181920] border-[#2d303e] text-[#8e93a6] hover:text-white'
                          }`}
                        >
                          Rotasi Otomatis (Semua Akun)
                        </button>
                        <button
                          type="button"
                          onClick={() => setPoolScope('single_account')}
                          className={`flex-1 py-1.5 px-2 text-[11px] rounded-[5px] border text-center transition-colors cursor-pointer ${
                            poolScope === 'single_account'
                              ? 'bg-[#2a1d17] border-[#ea580c] text-[#f97316] font-medium'
                              : 'bg-[#181920] border-[#2d303e] text-[#8e93a6] hover:text-white'
                          }`}
                        >
                          Fokus 1 Akun Spesifik
                        </button>
                      </div>

                      {poolScope === 'single_account' ? (
                        <div>
                          <label className="block text-[11px] font-semibold text-[#8e93a6] mb-1">
                            Pilih Akun {activePreset.name} yang Ingin Diuji:
                          </label>
                          <SearchableSelect
                            options={accountOptions}
                            value={selectedAccountId}
                            onChange={(val) => setSelectedAccountId(val)}
                            placeholder={`Pilih dari ${selectedProviderAccounts.length} akun ${activePreset.name}...`}
                            searchPlaceholder="Cari nama akun atau proxy..."
                          />
                        </div>
                      ) : (
                        <p className="text-[11px] text-[#8e93a6] leading-relaxed">
                          Permintaan akan otomatis di-rotasi (round-robin / priority) oleh EkaRouter ke {activeAccountsCount} akun {activePreset.name} yang aktif dengan proteksi cooldown & failover otomatis.
                        </p>
                      )}
                    </div>
                  )}
                </div>
              ) : (
                <div className="space-y-3">
                  <div>
                    <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                      API Key {activePreset.name} (Direct Client Chrome) *
                    </label>
                    <div className="relative">
                      <input
                        type={showApiKey ? 'text' : 'password'}
                        value={apiKey}
                        onChange={(e) => setApiKey(e.target.value)}
                        placeholder={activePreset.keyPlaceholder || 'sk-...'}
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
                      Key langsung diuji dari browser Chrome ke endpoint provider tanpa melalui server backend.
                    </p>
                  </div>

                  <div>
                    <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                      Custom Proxy / Base URL (Opsional)
                    </label>
                    <input
                      type="text"
                      value={customEndpoint}
                      onChange={(e) => setCustomEndpoint(e.target.value)}
                      placeholder={activePreset.defaultEndpoint}
                      className="w-full h-9 px-3 text-[12px] font-mono rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] focus:outline-none focus:border-[#ea580c] transition-colors"
                    />
                  </div>
                </div>
              )}

              <div>
                <label className="block text-[11.5px] font-semibold text-[var(--text-secondary)] mb-1.5">
                  Pilih Model
                </label>
                <SearchableSelect
                  options={modelOptions}
                  value={selectedModel}
                  onChange={(val) => {
                    setSelectedModel(val)
                    setCustomModel('')
                  }}
                  placeholder="Pilih model..."
                  searchPlaceholder="Cari nama model..."
                />
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
                {customModel.trim() && (
                  <p className="text-[10.5px] text-[#ea580c] mt-1">
                    Model kustom aktif: <span className="font-mono">{customModel.trim()}</span> (menggantikan {selectedModel})
                  </p>
                )}
              </div>

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

                {executedInfo && (
                  <span className="text-[10.5px] font-mono px-2 py-0.5 rounded-[4px] bg-[#1a1b20] border border-[#2c303d] text-[#ea580c]">
                    {executedInfo}
                  </span>
                )}
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
                  <span className="text-[12px] font-mono">
                    {testMethod === 'pool_rotation'
                      ? 'Menghubungi EkaRouter Load-Balancer & Merotasi Kredensial...'
                      : 'Mengirim permintaan langsung ke endpoint provider...'}
                  </span>
                </div>
              ) : errorDetails ? (
                <div className="p-3.5 rounded-[7px] bg-rose-950/20 border border-rose-600/30 text-rose-300 text-[12px] space-y-1 font-mono">
                  <div className="font-bold flex items-center gap-2 text-rose-200">
                    <XCircle className="w-4 h-4 shrink-0" />
                    <span>Error Pengujian</span>
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
                <div className="py-12 text-center text-[#686d80] text-[12px] space-y-1">
                  <p>Belum ada respon pengujian.</p>
                  <p className="text-[11px] text-[#555a6d]">
                    Pilih target provider, pilih mode pengujian (Auto-Rotasi Pool atau API Key Langsung), lalu klik <strong>Run Test</strong>.
                  </p>
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
