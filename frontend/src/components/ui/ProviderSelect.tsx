import { useMemo } from 'react'
import { SearchableSelect, SearchableOption } from './SearchableSelect.tsx'
import { ProviderLogo } from './ProviderLogo.tsx'

export interface ProviderOption {
  id: string
  name: string
  kind?: string
  category?: string
}

interface ProviderSelectProps {
  value: string
  onChange: (value: string) => void
  providers: ProviderOption[]
  placeholder?: string
  disabled?: boolean
  className?: string
}

export function ProviderSelect({
  value,
  onChange,
  providers,
  placeholder = 'Select Provider...',
  disabled = false,
  className = '',
}: ProviderSelectProps) {
  const options: SearchableOption[] = useMemo(() => {
    return providers.map((p) => {
      const categoryName =
        p.category === 'oauth'
          ? 'OAuth Providers'
          : p.category === 'free_tier'
          ? 'Free Tier Providers'
          : p.kind === 'custom'
          ? 'Custom Backends'
          : 'Platform Providers'

      return {
        value: p.id,
        label: p.name,
        sublabel: p.kind ? `(${p.kind})` : `(${p.id})`,
        icon: <ProviderLogo providerId={p.id} name={p.name} size="sm" />,
        category: categoryName,
      }
    })
  }, [providers])

  return (
    <SearchableSelect
      options={options}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      searchPlaceholder="Search provider or kind..."
      disabled={disabled}
      className={className}
    />
  )
}
