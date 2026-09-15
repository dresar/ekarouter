import { useState, useEffect } from 'react'

interface ProviderLogoProps {
  providerId: string
  name?: string
  customIcon?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
}

const PROVIDER_ICON_MAP: Record<string, string> = {
  '01-ai': '/providers/01-ai.svg',
  '9inf': '/providers/9inf.svg',
  'node_9inference_cloud': '/providers/9inf.svg',
  'abuseipdb': '/providers/abuseipdb.svg',
  'agentrouter': '/providers/agentrouter.png',
  'ai21': '/providers/ai21.svg',
  'airforce': '/providers/airforce.png',
  'alicode': '/providers/alicode.png',
  'alicode-intl': '/providers/alicode-intl.png',
  'alims-intl': '/providers/alims-intl.png',
  'alitp-intl': '/providers/alitp-intl.png',
  'amp': '/providers/amp.png',
  'anthropic': '/providers/anthropic.png',
  'anthropic-m': '/providers/anthropic-m.png',
  'antigravity': '/providers/antigravity.png',
  'anyscale': '/providers/anyscale.svg',
  'api-airforce': '/providers/api-airforce.png',
  'apx': '/providers/apx.svg',
  'node_apinex': '/providers/apx.svg',
  'assemblyai': '/providers/assemblyai.png',
  'aws-polly': '/providers/aws-polly.png',
  'azure': '/providers/azure.png',
  'bai': '/providers/bai.svg',
  'node_b_ai_api': '/providers/bai.svg',
  'baichuan': '/providers/baichuan.svg',
  'baidu': '/providers/baidu.png',
  'baseten': '/providers/baseten.svg',
  'bazaarlink': '/providers/bazaarlink.png',
  'bedrock': '/providers/bedrock.svg',
  'betterstack': '/providers/betterstack.svg',
  'black-forest-labs': '/providers/black-forest-labs.png',
  'blackbox': '/providers/blackbox.png',
  'bluesminds': '/providers/bluesminds.png',
  'brave-search': '/providers/brave-search.png',
  'byteplus': '/providers/byteplus.png',
  'cartesia': '/providers/cartesia.png',
  'cerebras': '/providers/cerebras.png',
  'chroma': '/providers/chroma.svg',
  'chromadb': '/providers/chroma.svg',
  'chutes': '/providers/chutes.png',
  'clarifai': '/providers/clarifai.svg',
  'claude': '/providers/anthropic.png',
  'cline': '/providers/cline.png',
  'clinepass': '/providers/clinepass.png',
  'cloudflare': '/providers/cloudflare.svg',
  'cloudflare-ai': '/providers/cloudflare.svg',
  'cloudflare-r2': '/providers/cloudflare.svg',
  'codebuddy': '/providers/codebuddy-intl.png',
  'codebuddy-cn': '/providers/codebuddy-cn.png',
  'codebuddy-intl': '/providers/codebuddy-intl.png',
  'codex': '/providers/openai.svg',
  'cohere': '/providers/cohere.png',
  'comfyui': '/providers/comfyui.png',
  'commandcode': '/providers/commandcode.png',
  'continue': '/providers/continue.png',
  'copilot': '/providers/copilot.png',
  'coqui': '/providers/coqui.svg',
  'coqui-tts': '/providers/coqui.svg',
  'cursor': '/providers/cursor.png',
  'custom': '/providers/generic.svg',
  'dahl': '/providers/dahl.svg',
  'node_dahl_global': '/providers/dahl.svg',
  'dalle': '/providers/openai.svg',
  'dall-e': '/providers/openai.svg',
  'deepgram': '/providers/deepgram.png',
  'deepinfra': '/providers/deepinfra.svg',
  'deepseek': '/providers/deepseek.png',
  'deepseek-tui': '/providers/deepseek-tui.png',
  'devin': '/providers/devin.svg',
  'devin-cli': '/providers/devin.svg',
  'devin-free': '/providers/devin.svg',
  'discord': '/providers/discord.svg',
  'docker': '/providers/docker.svg',
  'doubao': '/providers/doubao.svg',
  'drizzle': '/providers/drizzle.svg',
  'droid': '/providers/droid.png',
  'edge-tts': '/providers/edge-tts.png',
  'edgetts': '/providers/edge-tts.png',
  'elevenlabs': '/providers/elevenlabs.png',
  'exa': '/providers/exa.png',
  'fal-ai': '/providers/fal-ai.png',
  'featherless': '/providers/featherless.png',
  'firecrawl': '/providers/firecrawl.png',
  'fireworks': '/providers/fireworks.png',
  'fish-audio': '/providers/fish-audio.png',
  'gemini': '/providers/gemini.png',
  'gemini-cli': '/providers/gemini-cli.png',
  'generic': '/providers/generic.svg',
  'generic-rest': '/providers/generic.svg',
  'generic-rest-api': '/providers/generic.svg',
  'github': '/providers/github.png',
  'github-copilot': '/providers/copilot.png',
  'github-models': '/providers/github.png',
  'gitlab': '/providers/gitlab.png',
  'glm': '/providers/glm.png',
  'glm-cn': '/providers/glm-cn.png',
  'google': '/providers/gemini.png',
  'google-cloud': '/providers/google.svg',
  'gcp': '/providers/google.svg',
  'r2': '/providers/cloudflare.svg',
  'supabase-storage': '/providers/supabase.svg',
  'google-pse': '/providers/google-pse.png',
  'google-tts': '/providers/google-tts.png',
  'grok': '/providers/xai.png',
  'grok-cli': '/providers/grok-cli.png',
  'grok-web': '/providers/grok-web.png',
  'groq': '/providers/groq.svg',
  'hermes': '/providers/hermes.png',
  'holver': '/providers/openai.svg',
  'huggingface': '/providers/huggingface.png',
  'hunyuan': '/providers/hunyuan.svg',
  'hyperbolic': '/providers/hyperbolic.png',
  'iflow': '/providers/iflow.png',
  'inworld': '/providers/inworld.png',
  'ipinfo': '/providers/ipinfo.png',
  'jcode': '/providers/jcode.png',
  'jina': '/providers/jina-reader.png',
  'jina-ai': '/providers/jina-ai.png',
  'jina-reader': '/providers/jina-reader.png',
  'k8s': '/providers/kubernetes.svg',
  'kilo-gateway': '/providers/kilo-gateway.png',
  'kilocode': '/providers/kilocode.png',
  'kilogateway': '/providers/kilogateway.png',
  'kimchi': '/providers/kimchi.png',
  'kimi': '/providers/kimi.png',
  'kimi-coding': '/providers/kimi.png',
  'kiro': '/providers/kiro.png',
  'kubernetes': '/providers/kubernetes.svg',
  'lambda': '/providers/lambda.png',
  'lambdalabs': '/providers/lambda.png',
  'lepton': '/providers/lepton.svg',
  'lingyi': '/providers/lingyi.svg',
  'linkup': '/providers/linkup.png',
  'llm7': '/providers/llm7.png',
  'local-device': '/providers/generic.svg',
  'longcat': '/providers/longcat.png',
  'mapbox': '/providers/mapbox.svg',
  'meta': '/providers/meta.svg',
  'midtrans': '/providers/midtrans.svg',
  'milvus': '/providers/milvus.svg',
  'mimo': '/providers/mimo.svg',
  'mimo-free': '/providers/mimo-free.png',
  'mimofree': '/providers/mimofree.png',
  'minimax': '/providers/minimax.png',
  'minimax-cn': '/providers/minimax-cn.png',
  'mistral': '/providers/mistral.png',
  'mmf': '/providers/mmf.png',
  'modal': '/providers/modal.png',
  'mongo': '/providers/mongodb.svg',
  'mongodb': '/providers/mongodb.svg',
  'moonshot': '/providers/moonshot.svg',
  'morph': '/providers/morph.png',
  'mysql': '/providers/mysql.svg',
  'nanobanana': '/providers/nanobanana.png',
  'nebius': '/providers/nebius.png',
  'neon': '/providers/neon.svg',
  'novita': '/providers/novita.png',
  'nvidia': '/providers/nvidia.png',
  'oa': '/providers/openai.svg',
  'oai-cc': '/providers/openai.svg',
  'oai-r': '/providers/openai.svg',
  'ollama': '/providers/ollama.png',
  'ollama-local': '/providers/ollama-local.png',
  'openagentic': '/providers/anthropic.png',
  'openai': '/providers/openai.svg',
  'openclaw': '/providers/openclaw.png',
  'opencode': '/providers/opencode.png',
  'opencode-go': '/providers/opencode-go.png',
  'opendesign': '/providers/opendesign.png',
  'openrouter': '/providers/openrouter.png',
  'perplexity': '/providers/perplexity.png',
  'perplexity-agent': '/providers/perplexity-agent.png',
  'perplexity-web': '/providers/perplexity-web.png',
  'pinecone': '/providers/pinecone.svg',
  'playht': '/providers/playht.png',
  'poolside': '/providers/poolside.png',
  'postgres': '/providers/postgresql.svg',
  'postgresql': '/providers/postgresql.svg',
  'posthog': '/providers/posthog.svg',
  'predibase': '/providers/predibase.png',
  'prisma': '/providers/prisma.svg',
  'qdrant': '/providers/qdrant.svg',
  'qoder': '/providers/qoder.png',
  'qwen': '/providers/qwen.png',
  'recraft': '/providers/recraft.png',
  'redis': '/providers/redis.svg',
  'reka': '/providers/reka.png',
  'replicate': '/providers/replicate.svg',
  'resend': '/providers/resend.svg',
  'roo': '/providers/roo.png',
  'runpod': '/providers/runpod.png',
  'runwayml': '/providers/runwayml.png',
  'sambanova': '/providers/sambanova.png',
  'sdwebui': '/providers/sdwebui.png',
  'searchapi': '/providers/searchapi.png',
  'searxng': '/providers/searxng.png',
  'selfhosted-embedding': '/providers/generic.svg',
  'selfhosted-stt': '/providers/generic.svg',
  'selfhosted-tts': '/providers/generic.svg',
  'sendgrid': '/providers/sendgrid.svg',
  'sensenova': '/providers/sensenova.svg',
  'sentry': '/providers/sentry.svg',
  'serpapi': '/providers/serpapi.svg',
  'serper': '/providers/serper.png',
  'siliconflow': '/providers/siliconflow.png',
  'slack': '/providers/slack.svg',
  'sqlite': '/providers/sqlite.svg',
  'stability-ai': '/providers/stability-ai.png',
  'stepfun': '/providers/stepfun.svg',
  'stripe': '/providers/stripe.svg',
  'supabase': '/providers/supabase.svg',
  'tabitoken': '/providers/openai.svg',
  'tavily': '/providers/tavily.svg',
  'tavily-search': '/providers/tavily.svg',
  'tavily-search-api': '/providers/tavily.svg',
  'tencent': '/providers/tencent.png',
  'together': '/providers/together.png',
  'tokenrouter': '/providers/tokenrouter.png',
  'topaz': '/providers/topaz.png',
  'tortoise': '/providers/tortoise.png',
  'trae': '/providers/trae.png',
  'twilio': '/providers/twilio.svg',
  'umami': '/providers/umami.svg',
  'upstage': '/providers/upstage.svg',
  'upstash': '/providers/upstash.svg',
  'venice': '/providers/venice.png',
  'vercel': '/providers/vercel.svg',
  'vercel-ai-gateway': '/providers/vercel.svg',
  'vertex': '/providers/vertex.png',
  'vertex-partner': '/providers/vertex.png',
  'virustotal': '/providers/virustotal.svg',
  'volcengine-ark': '/providers/volcengine-ark.png',
  'voyage-ai': '/providers/voyage-ai.png',
  'weaviate': '/providers/weaviate.svg',
  'webhook': '/providers/webhook.svg',
  'webhooks': '/providers/webhook.svg',
  'generic-webhook': '/providers/webhook.svg',
  'generic-web': '/providers/webhook.svg',
  'windsurf': '/providers/windsurf.png',
  'workbuddy': '/providers/workbuddy.png',
  'writer': '/providers/writer.png',
  'xai': '/providers/xai.png',
  'xiaomi-mimo': '/providers/xiaomi-mimo.png',
  'xiaomi-tokenplan': '/providers/xiaomi-tokenplan.png',
  'xquik': '/providers/xquik.png',
  'yi': '/providers/01-ai.svg',
  'youcom': '/providers/youcom.png',
  'zans': '/providers/zans.svg',
  'zanslab': '/providers/zans.svg',
  'zanslab-id': '/providers/zans.svg',
  'node_zanslab_id': '/providers/zans.svg',
  'zed': '/providers/zed.png',
  'zeroone': '/providers/01-ai.svg',
  'zhipu': '/providers/zhipu.svg',
}

function resolveIcon(providerId: string, name: string): string {
  const id = providerId.toLowerCase().trim()
  if (PROVIDER_ICON_MAP[id]) return PROVIDER_ICON_MAP[id]

  const cleaned = id
    .replace(/^node_/, '')
    .replace(/[-_](api|v\d+|cn|intl|free|web|local|cloud|id|global|search)$/i, '')
    .trim()

  if (PROVIDER_ICON_MAP[cleaned]) return PROVIDER_ICON_MAP[cleaned]

  const nameLower = name.toLowerCase().trim()
  const nameCleaned = nameLower.split(/[\s-_]+/)[0]

  if (nameCleaned && PROVIDER_ICON_MAP[nameCleaned]) {
    return PROVIDER_ICON_MAP[nameCleaned]
  }

  for (const [key, icon] of Object.entries(PROVIDER_ICON_MAP)) {
    if (key.length >= 3 && (id.includes(key) || nameLower.includes(key))) {
      return icon
    }
  }

  return '/providers/generic.svg'
}

let customIconCache: Record<string, string> = {}
let isFetchingCustomIcons = false
const customIconListeners = new Set<() => void>()

export function setCustomIconInMemory(providerId: string, url: string) {
  if (url) {
    customIconCache[providerId] = url
  } else {
    delete customIconCache[providerId]
  }
  customIconListeners.forEach((fn) => fn())
}

export function ProviderLogo({ providerId, name = '', customIcon, size = 'md', className = '' }: ProviderLogoProps) {
  const [hasError, setHasError] = useState(false)
  const [, setVersion] = useState(0)

  useEffect(() => {
    const listener = () => setVersion((v) => v + 1)
    customIconListeners.add(listener)
    if (!isFetchingCustomIcons && Object.keys(customIconCache).length === 0) {
      isFetchingCustomIcons = true
      fetch('/api/providers/custom-icons')
        .then((r) => r.json())
        .then((data: Record<string, { icon_url?: string }>) => {
          if (data && typeof data === 'object') {
            for (const [pId, item] of Object.entries(data)) {
              if (item && item.icon_url) {
                customIconCache[pId] = item.icon_url
              }
            }
            customIconListeners.forEach((fn) => fn())
          }
        })
        .catch(() => {})
    }
    return () => {
      customIconListeners.delete(listener)
    }
  }, [])

  const iconSrc = customIcon || customIconCache[providerId] || resolveIcon(providerId, name)

  const sizeClasses = {
    sm: 'w-5 h-5 text-[10px]',
    md: 'w-7 h-7 text-[11px]',
    lg: 'w-9 h-9 text-[13px]',
    xl: 'w-11 h-11 text-[15px]',
  }

  const containerSizes = {
    sm: 'w-6 h-6',
    md: 'w-8 h-8',
    lg: 'w-10 h-10',
    xl: 'w-12 h-12',
  }

  const initials = (name || providerId)
    .split(/[\s-_]+/)
    .slice(0, 2)
    .map((word) => word[0]?.toUpperCase() || '')
    .join('') || 'AI'

  if (hasError) {
    return (
      <div
        className={`${containerSizes[size]} rounded-[8px] bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center font-mono font-bold text-indigo-400 select-none shrink-0 ${className}`}
      >
        {initials}
      </div>
    )
  }

  return (
    <div
      className={`${containerSizes[size]} rounded-[8px] bg-white/[0.04] border border-white/10 flex items-center justify-center p-[2.5px] shrink-0 shadow-xs ${className}`}
    >
      <img
        src={iconSrc}
        alt={name || providerId}
        onError={() => setHasError(true)}
        className={`${sizeClasses[size]} object-contain`}
        loading="eager"
      />
    </div>
  )
}
