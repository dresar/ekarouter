import { RefreshCw, ServerCrash, ArrowLeft } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Button } from '../../components/ui/Button.tsx'

export interface ServerErrorPageProps {
  error?: string
  onRetry?: () => void
}

export function ServerErrorPage({ error = 'Gateway Connection Error', onRetry }: ServerErrorPageProps) {
  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center text-center p-6">
      <div className="w-12 h-12 rounded-full bg-[var(--status-danger)]/10 text-[var(--status-danger)] flex items-center justify-center mb-4">
        <ServerCrash className="w-6 h-6" />
      </div>
      <h1 className="text-[28px] font-bold font-mono text-[var(--status-danger)]">500</h1>
      <h2 className="text-[16px] font-semibold text-[var(--text-secondary)] mt-1">
        Backend Service Unavailable
      </h2>
      <p className="text-[12.5px] text-[var(--text-muted)] mt-1 max-w-sm">
        {error || 'Unable to connect to the EkaRouter Go backend service. Please check if the backend is running.'}
      </p>
      <div className="flex items-center gap-3 mt-6">
        {onRetry && (
          <Button
            variant="secondary"
            size="standard"
            onClick={onRetry}
            leftIcon={<RefreshCw className="w-4 h-4" />}
          >
            Retry Connection
          </Button>
        )}
        <Link to="/overview">
          <Button variant="primary" size="standard" leftIcon={<ArrowLeft className="w-4 h-4" />}>
            Command Center
          </Button>
        </Link>
      </div>
    </div>
  )
}
