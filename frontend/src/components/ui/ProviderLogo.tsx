import { useState } from 'react'

interface ProviderLogoProps {
  providerId: string
  name?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
}

const PROVIDER_ICON_MAP: Record<string, string> = {
  '01-ai': '/providers/01-ai.svg',
  'abuseipdb': '/providers/abuseipdb.png',
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
  'assemblyai': '/providers/assemblyai.png',
  'aws-polly': '/providers/aws-polly.png',
  'azure': '/providers/azure.png',
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
  'cloudflare-ai': '/providers/cloudflare-ai.png',
  'codebuddy-cn': '/providers/codebuddy-cn.png',
  'codebuddy-intl': '/providers/codebuddy-intl.png',
  'codex': '/providers/codex.png',
  'cohere': '/providers/cohere.png',
  'comfyui': '/providers/comfyui.png',
  'commandcode': '/providers/commandcode.png',
  'continue': '/providers/continue.png',
  'copilot': '/providers/copilot.png',
  'coqui': '/providers/coqui.png',
  'cursor': '/providers/cursor.png',
  'custom': '/providers/custom.png',
  'deepgram': '/providers/deepgram.png',
  'deepinfra': '/providers/deepinfra.svg',
  'deepseek': '/providers/deepseek.png',
  'deepseek-tui': '/providers/deepseek-tui.png',
  'devin': '/providers/devin.png',
  'devin-cli': '/providers/devin-cli.png',
  'discord': '/providers/discord.svg',
  'docker': '/providers/docker.svg',
  'doubao': '/providers/doubao.svg',
  'drizzle': '/providers/drizzle.svg',
  'droid': '/providers/droid.png',
  'edge-tts': '/providers/edge-tts.png',
  'edgetts': '/providers/edgetts.png',
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
  'github': '/providers/github.png',
  'github-copilot': '/providers/copilot.png',
  'gitlab': '/providers/gitlab.png',
  'glm': '/providers/glm.png',
  'glm-cn': '/providers/glm-cn.png',
  'google': '/providers/gemini.png',
  'google-pse': '/providers/google-pse.png',
  'google-tts': '/providers/google-tts.png',
  'grok': '/providers/xai.png',
  'grok-cli': '/providers/grok-cli.png',
  'grok-web': '/providers/grok-web.png',
  'groq': '/providers/groq.png',
  'hermes': '/providers/hermes.png',
  'huggingface': '/providers/huggingface.png',
  'hunyuan': '/providers/hunyuan.svg',
  'hyperbolic': '/providers/hyperbolic.png',
  'iflow': '/providers/iflow.png',
  'inworld': '/providers/inworld.png',
  'ipinfo': '/providers/ipinfo.png',
  'jcode': '/providers/jcode.png',
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
  'local-device': '/providers/local-device.png',
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
  'oai-cc': '/providers/oai-cc.png',
  'oai-r': '/providers/oai-r.png',
  'ollama': '/providers/ollama.png',
  'ollama-local': '/providers/ollama-local.png',
  'openai': '/providers/openai.png',
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
  'selfhosted-embedding': '/providers/selfhosted-embedding.png',
  'selfhosted-stt': '/providers/selfhosted-stt.png',
  'selfhosted-tts': '/providers/selfhosted-tts.png',
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
  'tavily': '/providers/tavily.png',
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
  'vercel': '/providers/vercel.png',
  'vercel-ai-gateway': '/providers/vercel-ai-gateway.png',
  'vertex': '/providers/vertex.png',
  'vertex-partner': '/providers/vertex-partner.png',
  'virustotal': '/providers/virustotal.svg',
  'volcengine-ark': '/providers/volcengine-ark.png',
  'voyage-ai': '/providers/voyage-ai.png',
  'weaviate': '/providers/weaviate.svg',
  'webhook': '/providers/webhook.svg',
  'webhooks': '/providers/webhook.svg',
  'windsurf': '/providers/windsurf.png',
  'workbuddy': '/providers/workbuddy.png',
  'writer': '/providers/writer.png',
  'xai': '/providers/xai.png',
  'xiaomi-mimo': '/providers/xiaomi-mimo.png',
  'xiaomi-tokenplan': '/providers/xiaomi-tokenplan.png',
  'xquik': '/providers/xquik.png',
  'yi': '/providers/01-ai.svg',
  'youcom': '/providers/youcom.png',
  'zed': '/providers/zed.png',
  'zeroone': '/providers/01-ai.svg',
  'zhipu': '/providers/zhipu.svg',
}

function resolveIcon(providerId: string, name: string): string {
  const id = providerId.toLowerCase().trim()
  if (PROVIDER_ICON_MAP[id]) return PROVIDER_ICON_MAP[id]

  const cleaned = id.replace(/[-_](api|v\d+|cn|intl|free|web|local|cloud)$/i, '')
  if (PROVIDER_ICON_MAP[cleaned]) return PROVIDER_ICON_MAP[cleaned]

  for (const key of Object.keys(PROVIDER_ICON_MAP)) {
    if (id.startsWith(key) || id.includes(key) || key.includes(id)) {
      return PROVIDER_ICON_MAP[key]
    }
  }

  const nameLower = name.toLowerCase()
  for (const key of Object.keys(PROVIDER_ICON_MAP)) {
    if (nameLower.includes(key) || key.includes(nameLower.split(' ')[0])) {
      return PROVIDER_ICON_MAP[key]
    }
  }

  return '/providers/local-device.png'
}

export function ProviderLogo({ providerId, name = '', size = 'md', className = '' }: ProviderLogoProps) {
  const [hasError, setHasError] = useState(false)

  const iconSrc = resolveIcon(providerId, name)

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
        className={`${containerSizes[size]} rounded-[6px] bg-[var(--bg-panel)] border border-[var(--border-strong)] flex items-center justify-center font-mono font-bold text-[var(--text-secondary)] select-none shrink-0 ${className}`}
      >
        {initials}
      </div>
    )
  }

  return (
    <div
      className={`${containerSizes[size]} rounded-[7px] bg-white/5 border border-[var(--border-subtle)] flex items-center justify-center p-[3px] shrink-0 ${className}`}
    >
      <img
        src={iconSrc}
        alt={name || providerId}
        onError={() => setHasError(true)}
        className={`${sizeClasses[size]} object-contain`}
        loading="lazy"
      />
    </div>
  )
}
