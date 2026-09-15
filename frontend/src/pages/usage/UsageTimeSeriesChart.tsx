import { useState } from 'react'
import { TimeSeriesPoint } from '../../types/api.ts'
import { formatCompactNumber } from '../../utils/formatters.ts'

interface UsageTimeSeriesChartProps {
  data: TimeSeriesPoint[]
}

export function UsageTimeSeriesChart({ data }: UsageTimeSeriesChartProps) {
  const [metricMode, setMetricMode] = useState<'tokens' | 'cost'>('tokens')
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)

  const defaultPoints: TimeSeriesPoint[] = [
    { timestamp: '1', label: '00:00', input_tokens: 2400000, output_tokens: 8000, total_tokens: 2408000, cost: 0.72 },
    { timestamp: '2', label: '03:00', input_tokens: 400000, output_tokens: 2000, total_tokens: 402000, cost: 0.12 },
    { timestamp: '3', label: '06:00', input_tokens: 50000, output_tokens: 500, total_tokens: 50500, cost: 0.02 },
    { timestamp: '4', label: '09:00', input_tokens: 120000, output_tokens: 1200, total_tokens: 121200, cost: 0.04 },
    { timestamp: '5', label: '12:00', input_tokens: 350000, output_tokens: 3000, total_tokens: 353000, cost: 0.11 },
    { timestamp: '6', label: '15:00', input_tokens: 980000, output_tokens: 4500, total_tokens: 984500, cost: 0.29 },
    { timestamp: '7', label: '18:00', input_tokens: 420000, output_tokens: 2100, total_tokens: 422100, cost: 0.13 },
    { timestamp: '8', label: '21:00', input_tokens: 110000, output_tokens: 800, total_tokens: 110800, cost: 0.03 },
  ]

  const chartData = data && data.length > 0 ? data : defaultPoints

  const values = chartData.map((d) => (metricMode === 'tokens' ? d.total_tokens : d.cost))
  const rawMax = Math.max(...values, metricMode === 'tokens' ? 2400000 : 2.0)
  const maxValue = rawMax === 0 ? (metricMode === 'tokens' ? 1000 : 1.0) : rawMax

  const width = 800
  const height = 180
  const padLeft = 60
  const padRight = 20
  const padTop = 15
  const padBottom = 25
  const plotW = width - padLeft - padRight
  const plotH = height - padTop - padBottom

  const coords = chartData.map((d, i) => {
    const val = metricMode === 'tokens' ? d.total_tokens : d.cost
    const x = padLeft + (i / Math.max(chartData.length - 1, 1)) * plotW
    const y = padTop + plotH - (val / maxValue) * plotH
    return { x, y, data: d, val }
  })

  const linePath = coords.reduce((acc, pt, i, arr) => {
    if (i === 0) return `M ${pt.x} ${pt.y}`
    const prev = arr[i - 1]
    const cx1 = prev.x + (pt.x - prev.x) / 2
    const cy1 = prev.y
    const cx2 = prev.x + (pt.x - prev.x) / 2
    const cy2 = pt.y
    return `${acc} C ${cx1} ${cy1}, ${cx2} ${cy2}, ${pt.x} ${pt.y}`
  }, '')

  const lastCoord = coords[coords.length - 1]
  const firstCoord = coords[0]
  const areaPath = `${linePath} L ${lastCoord?.x ?? 0} ${padTop + plotH} L ${firstCoord?.x ?? 0} ${padTop + plotH} Z`

  const yTicks = [
    { ratio: 1.0, label: metricMode === 'tokens' ? formatCompactNumber(maxValue) : `$${maxValue.toFixed(2)}` },
    { ratio: 0.75, label: metricMode === 'tokens' ? formatCompactNumber(maxValue * 0.75) : `$${(maxValue * 0.75).toFixed(2)}` },
    { ratio: 0.5, label: metricMode === 'tokens' ? formatCompactNumber(maxValue * 0.5) : `$${(maxValue * 0.5).toFixed(2)}` },
    { ratio: 0.25, label: metricMode === 'tokens' ? formatCompactNumber(maxValue * 0.25) : `$${(maxValue * 0.25).toFixed(2)}` },
    { ratio: 0.0, label: metricMode === 'tokens' ? '0' : '$0.00' },
  ]

  return (
    <div className="rounded-lg border border-[#232634] bg-[#14161d] p-4 space-y-3">
      <div className="flex items-center justify-between">
        <div className="inline-flex bg-[#1a1c25] p-0.5 rounded-[6px] border border-[#282a38]">
          <button
            type="button"
            onClick={() => setMetricMode('tokens')}
            className={`px-3 py-1 text-xs font-semibold rounded-[5px] transition-all ${
              metricMode === 'tokens'
                ? 'bg-[#ff6940] text-white shadow-sm'
                : 'text-zinc-400 hover:text-white'
            }`}
          >
            Tokens
          </button>
          <button
            type="button"
            onClick={() => setMetricMode('cost')}
            className={`px-3 py-1 text-xs font-semibold rounded-[5px] transition-all ${
              metricMode === 'cost'
                ? 'bg-[#ff6940] text-white shadow-sm'
                : 'text-zinc-400 hover:text-white'
            }`}
          >
            Cost
          </button>
        </div>

        {hoveredIndex !== null && coords[hoveredIndex] && (
          <div className="flex items-center gap-3 text-xs font-mono bg-[#1c1f2b] px-2.5 py-1 rounded border border-[#2d3040]">
            <span className="text-zinc-400">{coords[hoveredIndex].data.label}:</span>
            <span className="text-[#ff6940] font-semibold">
              {metricMode === 'tokens'
                ? `${coords[hoveredIndex].data.total_tokens.toLocaleString()} tokens`
                : `$${coords[hoveredIndex].data.cost.toFixed(4)}`}
            </span>
          </div>
        )}
      </div>

      <div className="relative w-full overflow-hidden">
        <svg
          viewBox={`0 0 ${width} ${height}`}
          className="w-full h-auto overflow-visible"
          preserveAspectRatio="none"
        >
          <defs>
            <linearGradient id="areaGradIndigo" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#6366f1" stopOpacity="0.45" />
              <stop offset="100%" stopColor="#6366f1" stopOpacity="0.0" />
            </linearGradient>
            <linearGradient id="areaGradGold" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#f59e0b" stopOpacity="0.45" />
              <stop offset="100%" stopColor="#f59e0b" stopOpacity="0.0" />
            </linearGradient>
            <filter id="lineGlow" x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation="3" result="blur" />
              <feComposite in="SourceGraphic" in2="blur" operator="over" />
            </filter>
          </defs>

          {yTicks.map((t, idx) => {
            const yPos = padTop + plotH * (1 - t.ratio)
            return (
              <g key={idx}>
                <line
                  x1={padLeft}
                  y1={yPos}
                  x2={width - padRight}
                  y2={yPos}
                  stroke="#232634"
                  strokeWidth="1"
                  strokeDasharray={t.ratio === 0 ? undefined : '2 3'}
                />
                <text
                  x={padLeft - 8}
                  y={yPos + 3.5}
                  textAnchor="end"
                  fill="#71717a"
                  fontSize="10"
                  fontFamily="monospace"
                >
                  {t.label}
                </text>
              </g>
            )
          })}

          <path
            d={areaPath}
            fill={metricMode === 'tokens' ? 'url(#areaGradIndigo)' : 'url(#areaGradGold)'}
          />

          <path
            d={linePath}
            fill="none"
            stroke={metricMode === 'tokens' ? '#818cf8' : '#fbbf24'}
            strokeWidth="2.5"
            strokeLinecap="round"
            filter="url(#lineGlow)"
          />

          {coords.map((pt, idx) => (
            <g
              key={idx}
              className="cursor-pointer"
              onMouseEnter={() => setHoveredIndex(idx)}
              onMouseLeave={() => setHoveredIndex(null)}
            >
              <circle
                cx={pt.x}
                cy={pt.y}
                r={hoveredIndex === idx ? 5 : 3}
                fill="#12141c"
                stroke={metricMode === 'tokens' ? '#818cf8' : '#fbbf24'}
                strokeWidth="2"
                className="transition-all"
              />
              <text
                x={pt.x}
                y={height - 6}
                textAnchor="middle"
                fill="#71717a"
                fontSize="10"
                fontFamily="monospace"
              >
                {pt.data.label}
              </text>
            </g>
          ))}
        </svg>
      </div>
    </div>
  )
}
