import { AlertTriangle, RefreshCw } from 'lucide-react'
import { Button } from './Button.tsx'

export interface ErrorBannerProps {
  message: string
  onRetry?: () => void
  isRetrying?: boolean
  className?: string
}

export function ErrorBanner({
  message,
  onRetry,
  isRetrying = false,
  className = '',
}: ErrorBannerProps) {
  return (
    <div
      className={`flex items-center justify-between p-3 rounded-[6px] border border-[var(--status-danger)]/30 bg-[var(--status-danger)]/10 text-[var(--status-danger)] text-[13px] ${className}`}
    >
      <div className="flex items-center gap-2">
        <AlertTriangle className="w-4 h-4 shrink-0" />
        <span>{message}</span>
      </div>
      {onRetry && (
        <Button
          variant="danger"
          size="compact"
          isLoading={isRetrying}
          onClick={onRetry}
          leftIcon={<RefreshCw className="w-3.5 h-3.5" />}
        >
          Retry
        </Button>
      )}
    </div>
  )
}
