import { useState, useRef, useEffect, useMemo } from 'react'
import { Search, ChevronDown, Check, X } from 'lucide-react'

export interface SearchableOption {
  value: string
  label: string
  sublabel?: string
  icon?: React.ReactNode
  category?: string
}

export interface SearchableSelectProps {
  options: SearchableOption[]
  value: string
  onChange: (value: string) => void
  placeholder?: string
  searchPlaceholder?: string
  disabled?: boolean
  className?: string
  allowClear?: boolean
  emptyMessage?: string
}

export function SearchableSelect({
  options,
  value,
  onChange,
  placeholder = 'Pilih opsi...',
  searchPlaceholder = 'Ketik untuk mencari...',
  disabled = false,
  className = '',
  allowClear = false,
  emptyMessage = 'Tidak ada hasil yang cocok',
}: SearchableSelectProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [openUpward, setOpenUpward] = useState(false)
  const [query, setQuery] = useState('')
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const selectedOption = useMemo(() => {
    return options.find((opt) => opt.value === value)
  }, [options, value])

  const filteredOptions = useMemo(() => {
    if (!query.trim()) return options
    const q = query.toLowerCase().trim()
    return options.filter(
      (opt) =>
        opt.label.toLowerCase().includes(q) ||
        opt.value.toLowerCase().includes(q) ||
        (opt.sublabel && opt.sublabel.toLowerCase().includes(q)) ||
        (opt.category && opt.category.toLowerCase().includes(q))
    )
  }, [options, query])

  const groupedOptions = useMemo(() => {
    const hasCategories = filteredOptions.some((opt) => !!opt.category)
    if (!hasCategories) return { default: filteredOptions }

    const groups: Record<string, SearchableOption[]> = {}
    filteredOptions.forEach((opt) => {
      const cat = opt.category || 'Lainnya'
      if (!groups[cat]) groups[cat] = []
      groups[cat].push(opt)
    })
    return groups
  }, [filteredOptions])

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      if (containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect()
        const spaceBelow = window.innerHeight - rect.bottom
        setOpenUpward(spaceBelow < 280 && rect.top > spaceBelow)
      }
      document.addEventListener('mousedown', handleClickOutside)
      document.addEventListener('keydown', handleKeyDown)
      setTimeout(() => inputRef.current?.focus(), 50)
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [isOpen])

  const handleSelect = (val: string) => {
    onChange(val)
    setIsOpen(false)
    setQuery('')
  }

  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation()
    onChange('')
    setQuery('')
  }

  return (
    <div ref={containerRef} className={`relative ${className}`}>
      <button
        type="button"
        disabled={disabled}
        onClick={() => {
          if (!disabled) {
            setIsOpen(!isOpen)
            setQuery('')
          }
        }}
        className={`w-full h-9 px-3 text-[13px] rounded-[6px] bg-[#181920] border transition-colors flex items-center justify-between gap-2 text-left cursor-pointer ${
          isOpen
            ? 'border-[#ea580c] ring-1 ring-[#ea580c]/30'
            : 'border-[#2d313e] hover:border-[#424758]'
        } ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
      >
        <div className="flex items-center gap-2 min-w-0 flex-1">
          {selectedOption?.icon && <span className="shrink-0">{selectedOption.icon}</span>}
          <div className="min-w-0 flex-1 truncate">
            {selectedOption ? (
              <span className="text-[#f3f4f6] font-medium">{selectedOption.label}</span>
            ) : (
              <span className="text-[#6b7280]">{placeholder}</span>
            )}
            {selectedOption?.sublabel && (
              <span className="text-[11px] font-mono text-[#8e93a6] ml-2 truncate">
                {selectedOption.sublabel}
              </span>
            )}
          </div>
        </div>

        <div className="flex items-center gap-1 shrink-0">
          {allowClear && selectedOption && (
            <span
              role="button"
              tabIndex={0}
              onClick={handleClear}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  handleClear(e as any)
                }
              }}
              className="p-0.5 text-[#6b7280] hover:text-white rounded hover:bg-[#282b37] cursor-pointer"
            >
              <X className="w-3.5 h-3.5" />
            </span>
          )}
          <ChevronDown
            className={`w-4 h-4 text-[#8e93a6] transition-transform duration-200 ${
              isOpen ? 'rotate-180 text-[#ea580c]' : ''
            }`}
          />
        </div>
      </button>

      {isOpen && (
        <div className={`absolute left-0 ${openUpward ? 'bottom-full mb-1.5' : 'top-full mt-1.5'} z-50 w-full min-w-[280px] rounded-[8px] bg-[#14151c] border border-[#2e3240] shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-100`}>
          <div className="p-2 border-b border-[#242733] bg-[#111217]">
            <div className="relative">
              <Search className="w-3.5 h-3.5 text-[#7c8294] absolute left-2.5 top-1/2 -translate-y-1/2" />
              <input
                ref={inputRef}
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder={searchPlaceholder}
                className="w-full h-8 pl-8 pr-7 text-[12px] rounded-[5px] bg-[#1b1c24] border border-[#2f3342] text-[#f3f4f6] placeholder-[#64687a] focus:outline-none focus:border-[#ea580c] transition-colors"
              />
              {query && (
                <button
                  type="button"
                  onClick={() => setQuery('')}
                  className="absolute right-2 top-1/2 -translate-y-1/2 p-0.5 text-[#7c8294] hover:text-white cursor-pointer"
                >
                  <X className="w-3 h-3" />
                </button>
              )}
            </div>
          </div>

          <div className="max-h-60 overflow-y-auto p-1 space-y-0.5 divide-y divide-[#1e2029] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-[#363a48] [&::-webkit-scrollbar-thumb]:rounded-full">
            {filteredOptions.length === 0 ? (
              <div className="py-6 text-center text-[12px] text-[#787d90]">
                {emptyMessage}
              </div>
            ) : (
              Object.entries(groupedOptions).map(([category, items]) => (
                <div key={category} className="pt-1 first:pt-0">
                  {category !== 'default' && (
                    <div className="px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider text-[#6a6f80]">
                      {category}
                    </div>
                  )}
                  <div className="space-y-0.5">
                    {items.map((opt) => {
                      const isSelected = opt.value === value
                      return (
                        <button
                          key={opt.value}
                          type="button"
                          onClick={() => handleSelect(opt.value)}
                          className={`w-full px-2.5 py-2 text-[12.5px] rounded-[5px] flex items-center justify-between gap-2 text-left transition-colors cursor-pointer ${
                            isSelected
                              ? 'bg-[#2a1d17] text-[#f97316] font-medium'
                              : 'text-[#d1d5db] hover:bg-[#1f212a] hover:text-white'
                          }`}
                        >
                          <div className="flex items-center gap-2.5 min-w-0 flex-1">
                            {opt.icon && <span className="shrink-0">{opt.icon}</span>}
                            <div className="min-w-0 flex-1">
                              <span className="block truncate">{opt.label}</span>
                              {opt.sublabel && (
                                <span className="block text-[11px] font-mono text-[#787d90] truncate">
                                  {opt.sublabel}
                                </span>
                              )}
                            </div>
                          </div>
                          {isSelected && (
                            <Check className="w-4 h-4 text-[#ea580c] shrink-0 ml-2" />
                          )}
                        </button>
                      )
                    })}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
