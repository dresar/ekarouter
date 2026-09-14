import { ReactNode } from 'react'

export interface TabItem {
  key: string
  label: string
  icon?: ReactNode
  badge?: string | number
}

export interface TabsProps {
  items: TabItem[]
  activeKey: string
  onChange: (key: string) => void
  className?: string
}

export function Tabs({ items, activeKey, onChange, className = '' }: TabsProps) {
  return (
    <div className={`flex border-b border-[var(--border-strong)] gap-1 overflow-x-auto ${className}`}>
      {items.map((tab) => {
        const isActive = tab.key === activeKey
        return (
          <button
            key={tab.key}
            type="button"
            onClick={() => onChange(tab.key)}
            className={`flex items-center gap-2 px-3.5 py-2 text-[13px] font-medium border-b-2 -mb-px transition-colors whitespace-nowrap select-none ${
              isActive
                ? 'border-[var(--brand-primary)] text-[var(--brand-text)] font-semibold'
                : 'border-transparent text-[var(--text-muted)] hover:text-[var(--text-secondary)] hover:border-[var(--border-strong)]'
            }`}
          >
            {tab.icon && <span className="text-[14px]">{tab.icon}</span>}
            <span>{tab.label}</span>
            {tab.badge !== undefined && (
              <span className="ml-1 px-1.5 py-0.2 bg-[var(--border-strong)] text-[var(--text-secondary)] text-[10px] rounded-full">
                {tab.badge}
              </span>
            )}
          </button>
        )
      })}
    </div>
  )
}
