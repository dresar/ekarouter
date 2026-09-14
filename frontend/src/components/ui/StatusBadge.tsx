import { HTMLAttributes } from 'react'

export type StatusVariant =
  | 'active'
  | 'healthy'
  | 'success'
  | 'cooling_down'
  | 'warning'
  | 'degraded'
  | 'disabled'
  | 'paused'
  | 'neutral'
  | 'error'
  | 'revoked'
  | 'unhealthy'

export interface StatusBadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: StatusVariant
  pulse?: boolean
}

export function StatusBadge({ variant = 'neutral', pulse = false, children, className = '', ...props }: StatusBadgeProps) {
  const variantStyles: Record<StatusVariant, { bg: string; dot: string; text: string }> = {
    active: {
      bg: 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/25',
      dot: 'bg-emerald-400',
      text: 'text-emerald-400',
    },
    healthy: {
      bg: 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/25',
      dot: 'bg-emerald-400',
      text: 'text-emerald-400',
    },
    success: {
      bg: 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/25',
      dot: 'bg-emerald-400',
      text: 'text-emerald-400',
    },
    cooling_down: {
      bg: 'bg-amber-500/10 text-amber-400 border border-amber-500/25',
      dot: 'bg-amber-400',
      text: 'text-amber-400',
    },
    warning: {
      bg: 'bg-amber-500/10 text-amber-400 border border-amber-500/25',
      dot: 'bg-amber-400',
      text: 'text-amber-400',
    },
    degraded: {
      bg: 'bg-amber-500/10 text-amber-400 border border-amber-500/25',
      dot: 'bg-amber-400',
      text: 'text-amber-400',
    },
    disabled: {
      bg: 'bg-slate-500/10 text-slate-400 border border-slate-500/25',
      dot: 'bg-slate-400',
      text: 'text-slate-400',
    },
    paused: {
      bg: 'bg-slate-500/10 text-slate-400 border border-slate-500/25',
      dot: 'bg-slate-400',
      text: 'text-slate-400',
    },
    neutral: {
      bg: 'bg-slate-500/10 text-slate-400 border border-slate-500/25',
      dot: 'bg-slate-400',
      text: 'text-slate-400',
    },
    error: {
      bg: 'bg-rose-500/10 text-rose-400 border border-rose-500/25',
      dot: 'bg-rose-400',
      text: 'text-rose-400',
    },
    revoked: {
      bg: 'bg-rose-500/10 text-rose-400 border border-rose-500/25',
      dot: 'bg-rose-400',
      text: 'text-rose-400',
    },
    unhealthy: {
      bg: 'bg-rose-500/10 text-rose-400 border border-rose-500/25',
      dot: 'bg-rose-400',
      text: 'text-rose-400',
    },
  }

  const current = variantStyles[variant] || variantStyles.neutral

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 text-[11px] font-medium rounded-[4px] select-none ${current.bg} ${className}`}
      {...props}
    >
      <span className="relative flex h-1.5 w-1.5">
        {pulse && (
          <span
            className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${current.dot}`}
          />
        )}
        <span className={`relative inline-flex rounded-full h-1.5 w-1.5 ${current.dot}`} />
      </span>
      <span>{children}</span>
    </span>
  )
}
