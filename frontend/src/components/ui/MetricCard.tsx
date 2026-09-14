import { ReactNode } from 'react'

export interface MetricCardProps {
  label: string
  value: string | number
  subtext?: string
  icon?: ReactNode
  trend?: {
    value: string
    isPositive?: boolean
  }
}

export function MetricCard({ label, value, subtext, icon, trend }: MetricCardProps) {
  return (
    <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-3 flex flex-col justify-between h-[88px] transition-all hover:border-[var(--brand-primary)]/40 hover:bg-[var(--bg-panel)]/30">
      <div className="flex items-center justify-between">
        <span className="text-[10.5px] font-semibold tracking-wider uppercase text-[var(--text-muted)] truncate">
          {label}
        </span>
        {icon && <span className="text-[var(--brand-text)] opacity-80">{icon}</span>}
      </div>
      <div className="flex items-baseline justify-between mt-1">
        <span className="text-[20px] font-bold tracking-tight text-[var(--text-primary)] font-mono">
          {value}
        </span>
        {trend && (
          <span
            className={`text-[10.5px] font-medium font-mono ${
              trend.isPositive ? 'text-[var(--status-success)]' : 'text-[var(--status-danger)]'
            }`}
          >
            {trend.value}
          </span>
        )}
        {subtext && !trend && (
          <span className="text-[10.5px] text-[var(--text-muted)] truncate font-mono">{subtext}</span>
        )}
      </div>
    </div>
  )
}
