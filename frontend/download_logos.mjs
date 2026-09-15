import fs from 'node:fs'
import path from 'node:path'
import https from 'node:https'

const targetDir = path.resolve('public/providers')
if (!fs.existsSync(targetDir)) {
  fs.mkdirSync(targetDir, { recursive: true })
}

const LOGO_SOURCES = {
  'openai.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/openai.svg',
  'anthropic.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/anthropic.svg',
  'google.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/google.svg',
  'gemini.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/googlegemini.svg',
  'groq.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/groq.svg',
  'cerebras.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cerebras.svg',
  'mistral.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/mistral.svg',
  'deepseek.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/deepseek.svg',
  'cohere.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cohere.svg',
  'openrouter.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/openrouter.svg',
  'together.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/togetherai.svg',
  'perplexity.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/perplexity.svg',
  'fireworks.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/fireworksai.svg',
  'xai.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/xai.svg',
  'nvidia.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/nvidia.svg',
  'ai21.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/ai21labs.svg',
  'replicate.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/replicate.svg',
  'ollama.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/ollama.svg',
  'cloudflare.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cloudflare.svg',
  'bedrock.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/amazonwebservices.svg',
  'azure.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/microsoftazure.svg',
  'vertex.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/googlecloud.svg',
  'github.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/github.svg',
  'deepinfra.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/deepinfra.svg',
  'sambanova.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/sambanova.svg',
  'huggingface.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/huggingface.svg',
  'lepton.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/leptonai.svg',
  'qwen.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/alibabacloud.svg',
  'meta.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/meta.svg',
  'supabase.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/supabase.svg',
  'neon.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/neon.svg',
  'vercel.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/vercel.svg',
  'posthog.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/posthog.svg',
  'sentry.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/sentry.svg',
  'stripe.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/stripe.svg',
  'slack.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/slack.svg',
  'discord.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/discord.svg',
  'twilio.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/twilio.svg',
  'sendgrid.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/sendgrid.svg',
  'resend.svg': 'https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/resend.svg'
}

function fetchUrl(url) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { 'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)' } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return fetchUrl(res.headers.location).then(resolve).catch(reject)
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`Failed ${url}: HTTP ${res.statusCode}`))
      }
      let data = ''
      res.on('data', (chunk) => { data += chunk })
      res.on('end', () => resolve(data))
    }).on('error', reject)
  })
}

async function run() {
  console.log('Downloading genuine official provider SVG logos...')
  let successCount = 0
  for (const [filename, url] of Object.entries(LOGO_SOURCES)) {
    try {
      let svg = await fetchUrl(url)
      // If SVG has no fill, ensure it has fill="currentColor" so it works across dark/light mode
      if (!svg.includes('fill=') && !svg.includes('stroke=')) {
        svg = svg.replace('<path ', '<path fill="currentColor" ')
      }
      const outPath = path.join(targetDir, filename)
      fs.writeFileSync(outPath, svg.trim() + '\n', 'utf8')
      console.log(`[OK] ${filename}`)
      successCount++
    } catch (e) {
      console.error(`[ERR] ${filename}: ${e.message}`)
    }
  }
  console.log(`Completed: ${successCount}/${Object.keys(LOGO_SOURCES).length} genuine official logos saved.`)
}

run()
