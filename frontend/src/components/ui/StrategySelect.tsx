import { useMemo } from 'react'
import { ArrowDownRight, RotateCw, BarChart2 } from 'lucide-react'
import { SearchableSelect, SearchableOption } from './SearchableSelect.tsx'

interface StrategySelectProps {
  value: 'priority' | 'round_robin' | 'least_used'
  onChange: (value: 'priority' | 'round_robin' | 'least_used') => void
  disabled?: boolean
  className?: string
}

export function StrategySelect({
  value,
  onChange,
  disabled = false,
  className = '',
}: StrategySelectProps) {
  const options: SearchableOption[] = useMemo(() => {
    return [
      {
        value: 'priority',
        label: 'Priority Fallback',
        sublabel: 'Tier 1 then Tier 2',
        icon: (
          <div className="w-5 h-5 rounded-[4px] bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center">
            <ArrowDownRight className="w-3.5 h-3.5 text-emerald-400" />
          </div>
        ),
      },
      {
        value: 'round_robin',
        label: 'Round Robin',
        sublabel: 'Distribute evenly',
        icon: (
          <div className="w-5 h-5 rounded-[4px] bg-blue-500/10 border border-blue-500/30 flex items-center justify-center">
            <RotateCw className="w-3.5 h-3.5 text-blue-400" />
          </div>
        ),
      },
      {
        value: 'least_used',
        label: 'Least Used',
        sublabel: 'Balance load by quota',
        icon: (
          <div className="w-5 h-5 rounded-[4px] bg-purple-500/10 border border-purple-500/30 flex items-center justify-center">
            <BarChart2 className="w-3.5 h-3.5 text-purple-400" />
          </div>
        ),
      },
    ]
  }, [])

  return (
    <SearchableSelect
      options={options}
      value={value}
      onChange={(val) => onChange(val as any)}
      placeholder="Select strategy..."
      disabled={disabled}
      className={className}
    />
  )
}
