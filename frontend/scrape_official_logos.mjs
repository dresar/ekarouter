import fs from 'node:fs'
import path from 'node:path'
import https from 'node:https'

const targetDir = path.resolve('public/providers')

const DOMAINS = {
  'xai.svg': 'https://x.ai',
  'together.svg': 'https://together.ai',
  'fireworks.svg': 'https://fireworks.ai',
  'ai21.svg': 'https://www.ai21.com',
  'deepinfra.svg': 'https://deepinfra.com',
  'sambanova.svg': 'https://sambanova.ai',
  'lepton.svg': 'https://lepton.ai',
  'azure.svg': 'https://azure.microsoft.com',
  'bedrock.svg': 'https://aws.amazon.com',
  'moonshot.svg': 'https://www.moonshot.cn',
  'zhipu.svg': 'https://www.zhipuai.cn',
  'minimax.svg': 'https://www.minimaxi.com',
  'stepfun.svg': 'https://www.stepfun.com',
  'mimo.svg': 'https://mimo.xiaomi.com',
  'cohere.svg': 'https://cohere.com'
}

// Also check jsdelivr / unpkg @lobehub/icons or official Simple Icons alternative slugs
const CDN_FALLBACKS = {
  'xai.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/xai.svg',
  'together.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/together.svg',
  'fireworks.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/fireworks.svg',
  'ai21.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/ai21.svg',
  'deepinfra.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/deepinfra.svg',
  'sambanova.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/sambanova.svg',
  'lepton.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/lepton.svg',
  'azure.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/azure.svg',
  'bedrock.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/bedrock.svg',
  'moonshot.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/moonshot.svg',
  'zhipu.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/zhipu.svg',
  'minimax.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/minimax.svg',
  'stepfun.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/stepfun.svg',
  'mimo.svg': 'https://cdn.jsdelivr.net/npm/@lobehub/icons-static-svg@latest/icons/xiaomi.svg'
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
  console.log('Downloading genuine official vector logos from CDN/official brand repos...')
  for (const [filename, cdnUrl] of Object.entries(CDN_FALLBACKS)) {
    try {
      const content = await fetchUrl(cdnUrl)
      if (content && (content.includes('<svg') || content.includes('<?xml'))) {
        const outPath = path.join(targetDir, filename)
        fs.writeFileSync(outPath, content.trim() + '\n', 'utf8')
        console.log(`[OK] Saved genuine official vector ${filename} from ${cdnUrl}`)
        continue
      }
    } catch (e) {
      console.log(`CDN fallback failed for ${filename}: ${e.message}`)
    }
  }
}

run()
