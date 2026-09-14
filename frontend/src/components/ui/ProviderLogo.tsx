import { useState } from 'react'

interface ProviderLogoProps {
  providerId: string
  name?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
}

const PROVIDER_ICON_MAP: Record<string, string> = {
  openai: '/providers/openai.svg',
  anthropic: '/providers/anthropic.svg',
  claude: '/providers/anthropic.svg',
  google: '/providers/google.svg',
  gemini: '/providers/google.svg',
  'gemini-cli': '/providers/google.svg',
  antigravity: '/providers/google.svg',
  deepseek: '/providers/deepseek.svg',
  qwen: '/providers/qwen.svg',
  xai: '/providers/xai.svg',
  grok: '/providers/xai.svg',
  'grok-cli': '/providers/xai.svg',
  moonshot: '/providers/moonshot.svg',
  kimi: '/providers/moonshot.svg',
  groq: '/providers/groq.svg',
  mistral: '/providers/mistral.svg',
  ollama: '/providers/ollama.svg',
  nvidia: '/providers/nvidia.svg',
  meta: '/providers/meta.svg',
  cohere: '/providers/cohere.svg',
  cloudflare: '/providers/cloudflare.svg',
  'cloudflare-ai': '/providers/cloudflare.svg',
  github: '/providers/github.svg',
  together: '/providers/together.svg',
  openrouter: '/providers/openrouter.svg',
  minimax: '/providers/minimax.svg',
  zhipu: '/providers/zhipu.svg',
  stepfun: '/providers/stepfun.svg',
  mimo: '/providers/mimo.svg',
  'mimo-free': '/providers/mimo.svg',
  'xiaomi-tokenplan': '/providers/mimo.svg',
  cerebras: '/providers/cerebras.svg',
  huggingface: '/providers/huggingface.svg',
  supabase: '/providers/supabase.svg',
  neon: '/providers/neon.svg',
  vercel: '/providers/vercel.svg',
  'vercel-ai-gateway': '/providers/vercel.svg',
}

export function ProviderLogo({ providerId, name = '', size = 'md', className = '' }: ProviderLogoProps) {
  const [hasError, setHasError] = useState(false)

  const normalizedId = providerId.toLowerCase().trim().replace(/[-_]api$|[-_]v\d+$/i, '')
  let iconSrc = PROVIDER_ICON_MAP[normalizedId]

  if (!iconSrc) {
    for (const [key, path] of Object.entries(PROVIDER_ICON_MAP)) {
      if (normalizedId.includes(key) || name.toLowerCase().includes(key)) {
        iconSrc = path
        break
      }
    }
  }

  if (!iconSrc) {
    iconSrc = '/providers/generic.svg'
  }

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
      className={`${containerSizes[size]} rounded-[7px] bg-[var(--bg-panel)]/80 border border-[var(--border-subtle)] flex items-center justify-center p-1 shrink-0 ${className}`}
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
