import { ButtonHTMLAttributes, forwardRef, ReactNode } from 'react'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
  size?: 'compact' | 'standard'
  isLoading?: boolean
  leftIcon?: ReactNode
  rightIcon?: ReactNode
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      children,
      variant = 'primary',
      size = 'standard',
      isLoading = false,
      leftIcon,
      rightIcon,
      disabled,
      className = '',
      ...props
    },
    ref
  ) => {
    const baseStyles =
      'inline-flex items-center justify-center font-medium transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 disabled:opacity-50 disabled:cursor-not-allowed select-none'

    const sizeStyles =
      size === 'compact'
        ? 'h-[32px] px-3 text-[12px] rounded-[6px] gap-1.5'
        : 'h-[36px] px-4 text-[13px] rounded-[6px] gap-2'

    const variantStyles = {
      primary:
        'bg-[#23252d] text-[#f3f4f6] border border-[#393d4a] hover:bg-[#2d303a] hover:border-[#4b5062] focus-visible:outline-[#4b5062] active:bg-[#1c1e24] shadow-xs cursor-pointer',
      secondary:
        'bg-[var(--bg-panel)] text-[var(--text-primary)] border border-[var(--border-strong)] hover:bg-[var(--border-strong)] focus-visible:outline-[var(--border-strong)] cursor-pointer',
      danger:
        'bg-transparent text-[var(--status-danger)] border border-[var(--status-danger)]/40 hover:bg-[var(--status-danger)]/10 focus-visible:outline-[var(--status-danger)] cursor-pointer',
      ghost:
        'bg-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--border-strong)]/30 focus-visible:outline-[var(--border-strong)] cursor-pointer',
    }[variant]

    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        className={`${baseStyles} ${sizeStyles} ${variantStyles} ${className}`}
        {...props}
      >
        {isLoading ? (
          <svg
            className="animate-spin -ml-0.5 h-3.5 w-3.5 text-current"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        ) : (
          leftIcon
        )}
        <span>{children}</span>
        {!isLoading && rightIcon}
      </button>
    )
  }
)

Button.displayName = 'Button'
