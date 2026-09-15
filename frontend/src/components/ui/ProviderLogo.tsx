import { useState } from 'react'

interface ProviderLogoProps {
  providerId: string
  name?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
}

const PROVIDER_ICON_MAP: Record<string, string> = {
  'agentrouter': '/providers/agentrouter.png',
  'alicode-intl': '/providers/alicode-intl.png',
  'alicode': '/providers/alicode.png',
  'alims-intl': '/providers/alims-intl.png',
  'alitp-intl': '/providers/alitp-intl.png',
  'amp': '/providers/amp.png',
  'anthropic-m': '/providers/anthropic-m.png',
  'anthropic': '/providers/anthropic.png',
  'claude': '/providers/claude.png',
  'antigravity': '/providers/antigravity.png',
  'airforce': '/providers/airforce.png',
  'api-airforce': '/providers/airforce.png',
  'assemblyai': '/providers/assemblyai.png',
  'aws-polly': '/providers/aws-polly.png',
  'azure': '/providers/azure.png',
  'baidu': '/providers/baidu.png',
  'bazaarlink': '/providers/bazaarlink.png',
  'black-forest-labs': '/providers/black-forest-labs.png',
  'blackbox': '/providers/blackbox.png',
  'bluesminds': '/providers/bluesminds.png',
  'brave-search': '/providers/brave-search.png',
  'byteplus': '/providers/byteplus.png',
  'cartesia': '/providers/cartesia.png',
  'cerebras': '/providers/cerebras.png',
  'chutes': '/providers/chutes.png',
  'cloudflare-ai': '/providers/cloudflare-ai.png',
  'cloudflare': '/providers/cloudflare-ai.png',
  'codebuddy-cn': '/providers/codebuddy-cn.png',
  'codebuddy-intl': '/providers/codebuddy-intl.png',
  'codebuddy': '/providers/codebuddy-intl.png',
  'codex': '/providers/codex.png',
  'cohere': '/providers/cohere.png',
  'comfyui': '/providers/comfyui.png',
  'commandcode': '/providers/commandcode.png',
  'continue': '/providers/continue.png',
  'copilot': '/providers/copilot.png',
  'coqui': '/providers/coqui.png',
  'cursor': '/providers/cursor.png',
  'custom': '/providers/local-device.png',
  'deepgram': '/providers/deepgram.png',
  'deepseek-tui': '/providers/deepseek-tui.png',
  'deepseek': '/providers/deepseek.png',
  'devin-cli': '/providers/devin-cli.png',
  'devin': '/providers/devin-cli.png',
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
  'gemini-cli': '/providers/gemini-cli.png',
  'gemini': '/providers/gemini.png',
  'google': '/providers/gemini.png',
  'github': '/providers/github.png',
  'gitlab': '/providers/gitlab.png',
  'glm-cn': '/providers/glm-cn.png',
  'glm': '/providers/glm.png',
  'google-pse': '/providers/google-pse.png',
  'google-tts': '/providers/google-tts.png',
  'grok-cli': '/providers/grok-cli.png',
  'grok-web': '/providers/grok-web.png',
  'grok': '/providers/xai.png',
  'xai': '/providers/xai.png',
  'groq': '/providers/groq.png',
  'hermes': '/providers/hermes.png',
  'huggingface': '/providers/huggingface.png',
  'hyperbolic': '/providers/hyperbolic.png',
  'iflow': '/providers/iflow.png',
  'inworld': '/providers/inworld.png',
  'jcode': '/providers/jcode.png',
  'jina-ai': '/providers/jina-ai.png',
  'jina-reader': '/providers/jina-reader.png',
  'kilo-gateway': '/providers/kilo-gateway.png',
  'kilogateway': '/providers/kilo-gateway.png',
  'kilocode': '/providers/kilocode.png',
  'kimchi': '/providers/kimchi.png',
  'kimi-coding': '/providers/kimi-coding.png',
  'kimi': '/providers/kimi.png',
  'kiro': '/providers/kiro.png',
  'linkup': '/providers/linkup.png',
  'llm7': '/providers/llm7.png',
  'local-device': '/providers/local-device.png',
  'longcat': '/providers/longcat.png',
  'mimo-free': '/providers/mimo-free.png',
  'mimofree': '/providers/mimo-free.png',
  'mimo': '/providers/mimo-free.png',
  'xiaomi-mimo': '/providers/xiaomi-mimo.png',
  'minimax-cn': '/providers/minimax-cn.png',
  'minimax': '/providers/minimax.png',
  'mistral': '/providers/mistral.png',
  'mmf': '/providers/mmf.png',
  'morph': '/providers/morph.png',
  'nanobanana': '/providers/nanobanana.png',
  'nebius': '/providers/nebius.png',
  'novita': '/providers/novita.png',
  'nvidia': '/providers/nvidia.png',
  'oai-cc': '/providers/oai-cc.png',
  'oai-r': '/providers/oai-r.png',
  'ollama-local': '/providers/ollama-local.png',
  'ollama': '/providers/ollama.png',
  'openai': '/providers/openai.png',
  'openclaw': '/providers/openclaw.png',
  'opencode-go': '/providers/opencode-go.png',
  'opencode': '/providers/opencode.png',
  'opendesign': '/providers/opendesign.png',
  'openrouter': '/providers/openrouter.png',
  'perplexity-agent': '/providers/perplexity-agent.png',
  'perplexity-web': '/providers/perplexity-web.png',
  'perplexity': '/providers/perplexity.png',
  'playht': '/providers/playht.png',
  'poolside': '/providers/poolside.png',
  'qoder': '/providers/qoder.png',
  'qwen': '/providers/qwen.png',
  'recraft': '/providers/recraft.png',
  'reka': '/providers/reka.png',
  'roo': '/providers/roo.png',
  'runwayml': '/providers/runwayml.png',
  'sambanova': '/providers/sambanova.png',
  'sdwebui': '/providers/sdwebui.png',
  'searchapi': '/providers/searchapi.png',
  'searxng': '/providers/searxng.png',
  'selfhosted-embedding': '/providers/selfhosted-embedding.png',
  'selfhosted-stt': '/providers/selfhosted-stt.png',
  'selfhosted-tts': '/providers/selfhosted-tts.png',
  'serper': '/providers/serper.png',
  'siliconflow': '/providers/siliconflow.png',
  'stability-ai': '/providers/stability-ai.png',
  'tavily': '/providers/tavily.png',
  'tencent': '/providers/tencent.png',
  'together': '/providers/together.png',
  'tokenrouter': '/providers/tokenrouter.png',
  'topaz': '/providers/topaz.png',
  'tortoise': '/providers/tortoise.png',
  'trae': '/providers/trae.png',
  'venice': '/providers/venice.png',
  'vercel-ai-gateway': '/providers/vercel-ai-gateway.png',
  'vercel': '/providers/vercel.png',
  'vertex-partner': '/providers/vertex-partner.png',
  'vertex': '/providers/vertex.png',
  'volcengine-ark': '/providers/volcengine-ark.png',
  'voyage-ai': '/providers/voyage-ai.png',
  'windsurf': '/providers/windsurf.png',
  'workbuddy': '/providers/workbuddy.png',
  'xiaomi-tokenplan': '/providers/xiaomi-tokenplan.png',
  'xquik': '/providers/xquik.png',
  'youcom': '/providers/youcom.png',
  'zed': '/providers/zed.png',
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
