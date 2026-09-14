import { Link } from 'react-router-dom'
import { AlertCircle, ArrowLeft } from 'lucide-react'
import { Button } from '../../components/ui/Button.tsx'

export function NotFoundPage() {
  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center text-center p-6">
      <div className="w-12 h-12 rounded-full bg-[var(--status-danger)]/10 text-[var(--status-danger)] flex items-center justify-center mb-4">
        <AlertCircle className="w-6 h-6" />
      </div>
      <h1 className="text-[28px] font-bold font-mono text-[var(--text-primary)]">404</h1>
      <h2 className="text-[16px] font-semibold text-[var(--text-secondary)] mt-1">
        Page Not Found
      </h2>
      <p className="text-[12.5px] text-[var(--text-muted)] mt-1 max-w-sm">
        The console route you requested does not exist or has been relocated.
      </p>
      <div className="mt-6">
        <Link to="/overview">
          <Button variant="primary" size="standard" leftIcon={<ArrowLeft className="w-4 h-4" />}>
            Return to Command Center
          </Button>
        </Link>
      </div>
    </div>
  )
}
