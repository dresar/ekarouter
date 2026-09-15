import { useState, useRef } from 'react'
import { Plus, Minus, Maximize2, Zap } from 'lucide-react'
import { ProviderLogo } from '../../components/ui/ProviderLogo.tsx'

interface NodeItem {
  id: string
  name: string
  providerId: string
  x: number
  y: number
  active?: boolean
  category: 'provider' | 'client' | 'gateway'
}

interface UsageTopologyMapProps {
  activeProviders?: string[]
}

const TOPOLOGY_NODES: NodeItem[] = [
  { id: 'mimo', name: 'MiMo Code Free', providerId: 'mimo', x: 210, y: 70, category: 'client' },
  { id: 'opencode_free', name: 'OpenCode Free', providerId: 'opencode', x: 285, y: 55, category: 'client' },
  { id: 'qoder', name: 'Qoder', providerId: 'qoder', x: 355, y: 50, category: 'client' },
  { id: 'tabtoken', name: 'TabToken', providerId: 'tabtoken', x: 155, y: 125, category: 'client' },
  { id: 'gemini', name: 'Gemini', providerId: 'gemini', x: 140, y: 195, active: true, category: 'provider' },
  { id: 'openrouter', name: 'OpenRouter', providerId: 'openrouter', x: 115, y: 270, category: 'provider' },
  { id: 'groq', name: 'Groq', providerId: 'groq', x: 130, y: 345, category: 'provider' },
  { id: 'ollama', name: 'Ollama Cloud', providerId: 'ollama', x: 170, y: 405, category: 'provider' },
  { id: 'mistral', name: 'Mistral', providerId: 'mistral', x: 220, y: 450, category: 'provider' },
  { id: 'vercel', name: 'Vercel AI Gateway', providerId: 'vercel', x: 285, y: 475, category: 'gateway' },
  { id: 'cohere', name: 'Cohere', providerId: 'cohere', x: 355, y: 490, category: 'provider' },
  { id: 'chutes', name: 'Chutes AI', providerId: 'chutes', x: 425, y: 490, category: 'provider' },
  { id: 'nvidia', name: 'NVIDIA NIM', providerId: 'nvidia', x: 495, y: 475, category: 'provider' },
  { id: 'opencode_go', name: 'OpenCode Go', providerId: 'opencode', x: 565, y: 450, category: 'client' },
  { id: 'github_copilot', name: 'GitHub Copilot', providerId: 'github', x: 505, y: 55, category: 'client' },
  { id: 'cline', name: 'Cline', providerId: 'cline', x: 575, y: 85, category: 'client' },
  { id: 'gemini_cli', name: 'Gemini CLI', providerId: 'gemini', x: 635, y: 125, category: 'client' },
  { id: 'clinepass', name: 'ClinePass', providerId: 'clinepass', x: 650, y: 195, category: 'client' },
  { id: 'codebuddy', name: 'CodeBuddy', providerId: 'codebuddy', x: 670, y: 265, category: 'client' },
  { id: 'grok_cli', name: 'Grok CLI (Grok Build)', providerId: 'grok', x: 680, y: 335, category: 'client' },
  { id: 'kimi', name: 'Kimi', providerId: 'kimi', x: 655, y: 400, category: 'provider' },
  { id: 'kimi_ai', name: 'Kimi AI', providerId: 'kimi', x: 595, y: 440, category: 'provider' },
]

export function UsageTopologyMap({ activeProviders = ['gemini'] }: UsageTopologyMapProps) {
  const [zoom, setZoom] = useState(1)
  const [hoveredNode, setHoveredNode] = useState<string | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const hubX = 420
  const hubY = 270

  const handleZoomIn = () => setZoom((z) => Math.min(1.5, z + 0.1))
  const handleZoomOut = () => setZoom((z) => Math.max(0.7, z - 0.1))
  const handleResetZoom = () => setZoom(1)

  return (
    <div
      ref={containerRef}
      className="relative w-full h-[430px] rounded-lg border border-[#232634] bg-[#12141c] overflow-hidden select-none"
      style={{
        backgroundImage: 'radial-gradient(#242738 1px, transparent 1px)',
        backgroundSize: '20px 20px',
      }}
    >
      <div
        className="w-full h-full relative transition-transform duration-200 ease-out origin-center"
        style={{ transform: `scale(${zoom})` }}
      >
        <svg
          className="absolute inset-0 w-full h-full pointer-events-none"
          viewBox="0 0 840 540"
          preserveAspectRatio="xMidYMid meet"
        >
          <defs>
            <linearGradient id="activeBeam" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#f59e0b" stopOpacity="0.9" />
              <stop offset="100%" stopColor="#ff6940" stopOpacity="0.95" />
            </linearGradient>
            <filter id="glowGold" x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation="3.5" result="blur" />
              <feComposite in="SourceGraphic" in2="blur" operator="over" />
            </filter>
          </defs>

          {TOPOLOGY_NODES.map((node) => {
            const isActive =
              node.active ||
              activeProviders.some(
                (p) => p.toLowerCase() === node.providerId.toLowerCase() || p.toLowerCase() === node.id.toLowerCase()
              )
            const isHovered = hoveredNode === node.id

            const dx = (node.x - hubX) * 0.45
            const dy = (node.y - hubY) * 0.45
            const cx1 = node.x - dx
            const cy1 = node.y - dy * 0.2
            const cx2 = hubX + dx
            const cy2 = hubY + dy * 0.2

            const pathD = `M ${node.x} ${node.y} C ${cx1} ${cy1}, ${cx2} ${cy2}, ${hubX} ${hubY}`

            if (isActive || isHovered) {
              return (
                <g key={node.id}>
                  <path
                    d={pathD}
                    fill="none"
                    stroke="url(#activeBeam)"
                    strokeWidth="2.5"
                    strokeLinecap="round"
                    filter="url(#glowGold)"
                  />
                  <path
                    d={pathD}
                    fill="none"
                    stroke="#ffffff"
                    strokeWidth="1"
                    strokeDasharray="6 18"
                    className="animate-pulse"
                  />
                </g>
              )
            }

            return (
              <path
                key={node.id}
                d={pathD}
                fill="none"
                stroke="#2a2e40"
                strokeWidth="1"
                strokeDasharray="3 4"
                opacity="0.6"
              />
            )
          })}
        </svg>

        <div
          className="absolute z-20 -translate-x-1/2 -translate-y-1/2"
          style={{ left: `${(hubX / 840) * 100}%`, top: `${(hubY / 540) * 100}%` }}
        >
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-[7px] bg-[#1a1c26] border border-[#ff6940]/70 shadow-[0_0_24px_rgba(255,105,64,0.3)]">
            <div className="w-5 h-5 rounded-[4px] bg-[#ff6940] flex items-center justify-center text-white shadow-sm">
              <Zap className="w-3.5 h-3.5 fill-current" />
            </div>
            <span className="font-semibold text-[12px] tracking-wide text-white font-mono">
              9Router
            </span>
          </div>
        </div>

        {TOPOLOGY_NODES.map((node) => {
          const isActive =
            node.active ||
            activeProviders.some(
              (p) => p.toLowerCase() === node.providerId.toLowerCase() || p.toLowerCase() === node.id.toLowerCase()
            )
          const isHovered = hoveredNode === node.id

          return (
            <div
              key={node.id}
              onMouseEnter={() => setHoveredNode(node.id)}
              onMouseLeave={() => setHoveredNode(null)}
              className="absolute z-10 -translate-x-1/2 -translate-y-1/2 cursor-pointer transition-transform duration-150 hover:scale-105"
              style={{
                left: `${(node.x / 840) * 100}%`,
                top: `${(node.y / 540) * 100}%`,
              }}
            >
              <div
                className={`flex items-center gap-1.5 px-2 py-1 rounded-[5px] border backdrop-blur-sm transition-all duration-150 ${
                  isActive
                    ? 'bg-[#1e2230] border-amber-500/70 shadow-[0_0_12px_rgba(245,158,11,0.25)] text-white'
                    : isHovered
                    ? 'bg-[#222533] border-zinc-500 text-white shadow'
                    : 'bg-[#161822]/90 border-[#282b3a] text-zinc-300 hover:text-white'
                }`}
              >
                <div className="w-3.5 h-3.5 shrink-0 flex items-center justify-center">
                  <ProviderLogo providerId={node.providerId} size="sm" />
                </div>
                <span className="font-mono text-[10.5px] whitespace-nowrap leading-none">
                  {node.name}
                </span>
                {isActive && (
                  <span className="w-1.5 h-1.5 rounded-full bg-amber-400 shadow-[0_0_6px_#fbbf24] shrink-0" />
                )}
              </div>
            </div>
          )
        })}
      </div>

      <div className="absolute bottom-3 left-3 z-30 flex flex-col gap-1 bg-[#181a24]/90 backdrop-blur border border-[#2b2e3e] rounded-[6px] p-1 shadow-md">
        <button
          type="button"
          onClick={handleZoomIn}
          className="p-1 rounded-[4px] text-zinc-400 hover:text-white hover:bg-[#252838] transition-colors"
          title="Zoom In"
        >
          <Plus className="w-3.5 h-3.5" />
        </button>
        <button
          type="button"
          onClick={handleZoomOut}
          className="p-1 rounded-[4px] text-zinc-400 hover:text-white hover:bg-[#252838] transition-colors"
          title="Zoom Out"
        >
          <Minus className="w-3.5 h-3.5" />
        </button>
        <button
          type="button"
          onClick={handleResetZoom}
          className="p-1 rounded-[4px] text-zinc-400 hover:text-white hover:bg-[#252838] transition-colors"
          title="Reset View"
        >
          <Maximize2 className="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  )
}
