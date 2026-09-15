import { useState, FormEvent } from 'react'
import { AlertCircle, Cloud, UploadCloud, Terminal, ExternalLink } from 'lucide-react'
import { Button } from '../../components/ui/Button.tsx'
import { api } from '../../api/client.ts'

export type RelayPlatform = 'cloudflare' | 'vercel' | 'deno'

interface DeployRelayModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
  defaultPlatform?: RelayPlatform
}

export function DeployRelayModal({
  isOpen,
  onClose,
  onSuccess,
  defaultPlatform = 'cloudflare',
}: DeployRelayModalProps) {
  const [platform, setPlatform] = useState<RelayPlatform>(defaultPlatform)
  const [accountId, setAccountId] = useState('')
  const [apiToken, setApiToken] = useState('')
  const [workerName, setWorkerName] = useState('cloudflare-relay')
  const [projectId, setProjectId] = useState('')
  const [isDeploying, setIsDeploying] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!isOpen) return null

  const handleDeploy = async (e: FormEvent) => {
    e.preventDefault()
    if (!apiToken.trim()) {
      setError('API Token is required to deploy relay')
      return
    }

    setIsDeploying(true)
    setError(null)

    try {
      await api.post('/api/proxies/deploy-relay', {
        platform,
        account_id: accountId.trim(),
        api_token: apiToken.trim(),
        worker_name: workerName.trim(),
        project_id: projectId.trim(),
      })
      onSuccess()
      onClose()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to deploy relay'
      setError(msg)
    } finally {
      setIsDeploying(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-150">
      <div
        className="w-full max-w-[480px] max-h-[90vh] flex flex-col bg-[#1a1c23] border border-[#2e323e] rounded-[12px] shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-2 px-4 py-3 bg-[#15171d] border-b border-[#262934] select-none shrink-0">
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              onClick={onClose}
              className="w-3 h-3 rounded-full bg-[#ff5f56] hover:brightness-110 transition-all border border-[#e0443e]/40"
              title="Close"
            />
            <span className="w-3 h-3 rounded-full bg-[#ffbd2e] border border-[#dea123]/40" />
            <span className="w-3 h-3 rounded-full bg-[#27c93f] border border-[#1aab29]/40" />
          </div>
          <span className="text-[13.5px] font-semibold text-[#f3f4f6] ml-2 capitalize">
            Deploy {platform} Relay
          </span>
        </div>

        <div className="flex items-center border-b border-[#262934] bg-[#171920] px-3 gap-1 shrink-0">
          <button
            type="button"
            onClick={() => {
              setPlatform('cloudflare')
              setWorkerName('cloudflare-relay')
            }}
            className={`flex items-center gap-2 px-3 py-2 text-[12px] font-medium border-b-2 transition-colors cursor-pointer ${
              platform === 'cloudflare'
                ? 'border-[#ff6940] text-[#f3f4f6]'
                : 'border-transparent text-[#9ca3af] hover:text-[#e5e7eb]'
            }`}
          >
            <Cloud className="w-3.5 h-3.5 text-[#f6821f]" />
            <span>Cloudflare</span>
          </button>
          <button
            type="button"
            onClick={() => {
              setPlatform('vercel')
              setWorkerName('vercel-relay')
            }}
            className={`flex items-center gap-2 px-3 py-2 text-[12px] font-medium border-b-2 transition-colors cursor-pointer ${
              platform === 'vercel'
                ? 'border-[#ff6940] text-[#f3f4f6]'
                : 'border-transparent text-[#9ca3af] hover:text-[#e5e7eb]'
            }`}
          >
            <UploadCloud className="w-3.5 h-3.5 text-[#0070f3]" />
            <span>Vercel</span>
          </button>
          <button
            type="button"
            onClick={() => {
              setPlatform('deno')
              setWorkerName('deno-relay')
            }}
            className={`flex items-center gap-2 px-3 py-2 text-[12px] font-medium border-b-2 transition-colors cursor-pointer ${
              platform === 'deno'
                ? 'border-[#ff6940] text-[#f3f4f6]'
                : 'border-transparent text-[#9ca3af] hover:text-[#e5e7eb]'
            }`}
          >
            <Terminal className="w-3.5 h-3.5 text-[#10b981]" />
            <span>Deno</span>
          </button>
        </div>

        <form onSubmit={handleDeploy} className="p-5 space-y-4 overflow-y-auto flex-1 leading-relaxed">
          {error && (
            <div className="p-2.5 rounded-[6px] bg-[var(--status-danger)]/15 border border-[var(--status-danger)]/30 flex items-center gap-2 text-[12px] text-[var(--status-danger)]">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {platform === 'cloudflare' && (
            <div className="p-3.5 rounded-[8px] bg-[#221c19] border border-[#442c1e] text-[11.5px] space-y-2 text-[#d1d5db]">
              <div className="font-bold text-[#f97316] text-[12px]">
                What is Cloudflare Relay?
              </div>
              <p>
                Deploys a Cloudflare Worker as a proxy relay. All AI provider requests will be forwarded through Cloudflare's global edge network.
              </p>
              <ul className="list-disc pl-4 space-y-1 text-[#9ca3af]">
                <li>High performance global routing and IP masking via Cloudflare Workers</li>
                <li>Free tier: 100,000 requests per day</li>
                <li>Requires Cloudflare Account ID and a Workers API Token (Edit Workers permission)</li>
              </ul>
              <div className="pt-2 border-t border-[#442c1e] space-y-1">
                <div className="font-semibold text-[#f3f4f6]">
                  How to generate your API Token:
                </div>
                <ol className="list-decimal pl-4 space-y-0.5 text-[#9ca3af]">
                  <li>Go to <span className="text-[#f3f4f6]">My Profile → API Tokens → Create Token</span></li>
                  <li>Scroll down to <span className="text-[#f3f4f6]">Custom Token</span> and click <span className="text-[#f3f4f6]">Get started</span></li>
                  <li>Under <span className="text-[#f3f4f6]">Permissions</span>: Account | Workers Scripts | Edit</li>
                  <li>Under <span className="text-[#f3f4f6]">Account Resources</span>: Include | Account | Your Account Name</li>
                  <li>Click <span className="text-[#f3f4f6]">Continue to summary → Create Token</span></li>
                </ol>
              </div>
            </div>
          )}

          {platform === 'vercel' && (
            <div className="p-3.5 rounded-[8px] bg-[#17202e] border border-[#1e3450] text-[11.5px] space-y-2 text-[#d1d5db]">
              <div className="font-bold text-[#38bdf8] text-[12px]">
                What is Vercel Relay?
              </div>
              <p>
                Deploys a serverless Edge Relay on Vercel's global network. Forward requests to upstream LLMs with zero server overhead.
              </p>
              <ul className="list-disc pl-4 space-y-1 text-[#9ca3af]">
                <li>Low latency serverless execution with automatic HTTPS edge termination</li>
                <li>Requires Vercel Personal Access Token or Project Token</li>
              </ul>
            </div>
          )}

          {platform === 'deno' && (
            <div className="p-3.5 rounded-[8px] bg-[#16271e] border border-[#1d432e] text-[11.5px] space-y-2 text-[#d1d5db]">
              <div className="font-bold text-[#34d399] text-[12px]">
                What is Deno Relay?
              </div>
              <p>
                Deploys a lightweight Deno Deploy worker using standard Web Streams and high-efficiency fetch proxying.
              </p>
              <ul className="list-disc pl-4 space-y-1 text-[#9ca3af]">
                <li>Free tier: 100,000 requests per day</li>
                <li>Requires Deno Deploy Access Token and Project ID</li>
              </ul>
            </div>
          )}

          {platform === 'cloudflare' && (
            <div className="space-y-1">
              <label className="block text-[12px] font-semibold text-[#e5e7eb]">
                Account ID
              </label>
              <input
                type="text"
                required
                value={accountId}
                onChange={(e) => setAccountId(e.target.value)}
                placeholder="your-cloudflare-account-id"
                className="w-full px-3 py-2 text-[12.5px] font-mono rounded-[6px] bg-[#232630] border border-[#333745] text-[#f3f4f6] placeholder-[#6b7280] focus:outline-none focus:border-[#4b5563]"
              />
              <p className="text-[11px] text-[#9ca3af]">
                Found on the right side of the Cloudflare dashboard overview page.
              </p>
            </div>
          )}

          <div className="space-y-1">
            <div className="flex items-center justify-between">
              <label className="block text-[12px] font-semibold text-[#e5e7eb]">
                API Token *
              </label>
              <a
                href={
                  platform === 'cloudflare'
                    ? 'https://dash.cloudflare.com/profile/api-tokens'
                    : platform === 'vercel'
                    ? 'https://vercel.com/account/tokens'
                    : 'https://dash.deno.com/account#tokens'
                }
                target="_blank"
                rel="noreferrer"
                className="text-[11px] text-[#ff6940] hover:underline inline-flex items-center gap-1"
              >
                <span>Get token</span>
                <ExternalLink className="w-3 h-3" />
              </a>
            </div>
            <input
              type="password"
              required
              autoComplete="new-password"
              value={apiToken}
              onChange={(e) => setApiToken(e.target.value)}
              placeholder={`your-${platform}-api-token`}
              className="w-full px-3 py-2 text-[12.5px] font-mono rounded-[6px] bg-[#232630] border border-[#333745] text-[#f3f4f6] placeholder-[#6b7280] focus:outline-none focus:border-[#4b5563]"
            />
            <p className="text-[11px] text-[#9ca3af]">
              {platform === 'cloudflare' && 'Requires "Workers Scripts: Edit" permission.'}
              {platform === 'vercel' && 'Requires Vercel token with deployment permission.'}
              {platform === 'deno' && 'Requires Deno Deploy personal access token.'}
            </p>
          </div>

          <div className="space-y-1">
            <label className="block text-[12px] font-semibold text-[#e5e7eb]">
              {platform === 'deno' ? 'Project ID' : 'Worker / Project Name'}
            </label>
            <input
              type="text"
              value={platform === 'deno' ? projectId : workerName}
              onChange={(e) =>
                platform === 'deno' ? setProjectId(e.target.value) : setWorkerName(e.target.value)
              }
              placeholder={platform === 'cloudflare' ? 'cloudflare-relay' : `${platform}-relay`}
              className="w-full px-3 py-2 text-[12.5px] font-mono rounded-[6px] bg-[#232630] border border-[#333745] text-[#f3f4f6] placeholder-[#6b7280] focus:outline-none focus:border-[#4b5563]"
            />
            <p className="text-[11px] text-[#9ca3af]">
              Unique name for your {platform} relay. Leave empty for auto-generated name.
            </p>
          </div>

          <div className="pt-3 flex items-center justify-end gap-2 border-t border-[#262934] shrink-0">
            <Button
              type="button"
              variant="ghost"
              size="compact"
              onClick={onClose}
              className="text-[#9ca3af] hover:text-[#f3f4f6]"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="compact"
              isLoading={isDeploying}
              className="bg-[#2d313d] hover:bg-[#383d4c] text-[#f3f4f6] px-5 font-semibold"
            >
              Deploy Relay
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
