export const PROVIDER_API_KEY_URLS: Record<string, string> = {
  gemini: 'https://aistudio.google.com/app/apikey',
  'gemini-cli': 'https://aistudio.google.com/app/apikey',
  'gemini-agy': 'https://antigravity.google',
  antigravity: 'https://antigravity.google',
  openai: 'https://platform.openai.com/api-keys',
  codex: 'https://platform.openai.com/api-keys',
  anthropic: 'https://console.anthropic.com/settings/keys',
  claude: 'https://console.anthropic.com/settings/keys',
  groq: 'https://console.groq.com/keys',
  openrouter: 'https://openrouter.ai/keys',
  mistral: 'https://console.mistral.ai/api-keys',
  deepseek: 'https://platform.deepseek.com/api_keys',
  cohere: 'https://dashboard.cohere.com/api-keys',
  together: 'https://api.together.ai/settings/api-keys',
  'together-ai': 'https://api.together.ai/settings/api-keys',
  cerebras: 'https://cloud.cerebras.ai/platform/',
  perplexity: 'https://www.perplexity.ai/settings/api',
  github: 'https://github.com/settings/tokens',
  'github-copilot': 'https://github.com/settings/tokens',
  huggingface: 'https://huggingface.co/settings/tokens',
  cloudflare: 'https://dash.cloudflare.com/',
  'cloudflare-ai': 'https://dash.cloudflare.com/',
  xai: 'https://console.x.ai/',
  'grok-cli': 'https://console.x.ai/',
  siliconflow: 'https://cloud.siliconflow.cn/',
  kimi: 'https://platform.moonshot.cn/console/api-keys',
  moonshot: 'https://platform.moonshot.cn/console/api-keys',
  ollama: 'https://ollama.com/download',
  chutes: 'https://chutes.ai/',
  hyperbolic: 'https://app.hyperbolic.xyz/',
  'kilo-gateway': 'https://kilogateway.com/',
  kilocode: 'https://kilocode.ai/',
  byteplus: 'https://console.byteplus.com/',
  'codebuddy-intl': 'https://codebuddy.ai/',
  'codebuddy-cn': 'https://codebuddy.cn/',
  opencode: 'https://opencode.ai/',
  'opencode-go': 'https://opencode.ai/',
  qoder: 'https://qoder.ai/',
  cursor: 'https://www.cursor.com/',
  cline: 'https://cline.bot/',
  clinepass: 'https://cline.bot/',
  poolside: 'https://poolside.ai/',
  bazaarlink: 'https://bazaarlink.com/',
  'api-airforce': 'https://api.airforce/',
  venice: 'https://venice.ai/',
  tokenrouter: 'https://tokenrouter.io/',
  'vercel-ai-gateway': 'https://vercel.com/docs/ai/ai-gateway',
  nvidia: 'https://build.nvidia.com/',
  devin: 'https://devin.ai/',
  'devin-cli': 'https://devin.ai/',
  mimofree: 'https://ai.mi.com/',
  'mimo-free': 'https://ai.mi.com/',
  'xiaomi-tokenplan': 'https://ai.mi.com/',
  'xiaomi-mimo': 'https://ai.mi.com/',
  zed: 'https://zed.dev/',
  windsurf: 'https://codeium.com/windsurf',
  trae: 'https://www.trae.ai/',
  kiro: 'https://kiro.ai/',
  vertex: 'https://console.cloud.google.com/vertex-ai',
}

export function getProviderApiKeyUrl(providerId?: string, fallbackDocUrl?: string): string {
  if (!providerId) return fallbackDocUrl || 'https://google.com'
  const normalized = providerId.toLowerCase().trim()
  if (PROVIDER_API_KEY_URLS[normalized]) {
    return PROVIDER_API_KEY_URLS[normalized]
  }
  for (const [key, url] of Object.entries(PROVIDER_API_KEY_URLS)) {
    if (normalized.includes(key) || key.includes(normalized)) {
      return url
    }
  }
  if (fallbackDocUrl) return fallbackDocUrl
  return `https://www.google.com/search?q=${encodeURIComponent(providerId + ' api key console')}`
}
