import { useMemo } from 'react'
import { KeyRound, ShieldCheck } from 'lucide-react'
import { SearchableSelect, SearchableOption } from './SearchableSelect.tsx'
import { ProviderLogo } from './ProviderLogo.tsx'
import { Account } from '../../types/api.ts'

interface AccountSelectProps {
  value: string
  onChange: (value: string) => void
  providerId: string
  providerName?: string
  accounts: Account[]
  placeholder?: string
  disabled?: boolean
  className?: string
}

export function AccountSelect({
  value,
  onChange,
  providerId,
  providerName,
  accounts,
  placeholder = 'Default Provider Credential',
  disabled = false,
  className = '',
}: AccountSelectProps) {
  const options: SearchableOption[] = useMemo(() => {
    const list: SearchableOption[] = [
      {
        value: '',
        label: 'Default Provider Credential',
        sublabel: '(Auto / Env)',
        icon: providerId ? (
          <ProviderLogo providerId={providerId} name={providerName} size="sm" />
        ) : (
          <KeyRound className="w-3.5 h-3.5 text-[#f97316]" />
        ),
      },
    ]

    accounts.forEach((acc) => {
      list.push({
        value: acc.id,
        label: acc.name,
        sublabel: `Tier ${acc.priority}`,
        icon: (
          <div className="relative">
            <ProviderLogo providerId={providerId} name={providerName} size="sm" />
            <ShieldCheck className="w-2.5 h-2.5 text-emerald-400 absolute -bottom-1 -right-1" />
          </div>
        ),
      })
    })

    return list
  }, [accounts, providerId, providerName])

  return (
    <SearchableSelect
      options={options}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      searchPlaceholder="Search account keys..."
      disabled={disabled}
      className={className}
    />
  )
}
