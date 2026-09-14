import { Component, ErrorInfo, ReactNode } from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null, errorInfo: null }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({ error, errorInfo })
    if (import.meta.env.DEV) {
      console.error('[ErrorBoundary] Caught error:', error, errorInfo)
    }
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null, errorInfo: null })
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback

      const message = this.state.error?.message ?? 'Unknown error'
      const stack =
        import.meta.env.DEV && this.state.errorInfo?.componentStack
          ? this.state.errorInfo.componentStack.trim().slice(0, 600)
          : null

      return (
        <div className="min-h-[240px] flex items-center justify-center p-6">
          <div className="w-full max-w-lg rounded-[8px] border border-[var(--status-danger)]/30 bg-[var(--bg-card)] p-5 space-y-3">
            <div className="flex items-center gap-2.5">
              <AlertTriangle className="w-4 h-4 text-[var(--status-danger)] shrink-0" />
              <span className="text-[13px] font-semibold text-[var(--text-primary)]">
                Component Error
              </span>
            </div>
            <p className="text-[12px] text-[var(--text-secondary)] leading-relaxed font-mono break-all">
              {message}
            </p>
            {stack && (
              <pre className="text-[10.5px] text-[var(--text-muted)] bg-[var(--bg-surface)] rounded-[6px] p-3 overflow-x-auto max-h-32 leading-relaxed">
                {stack}
              </pre>
            )}
            <div className="flex items-center gap-2 pt-1">
              <button
                type="button"
                onClick={this.handleReset}
                className="flex items-center gap-1.5 px-3 py-1.5 text-[11.5px] font-medium rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-strong)] text-[var(--text-primary)] hover:bg-[var(--bg-surface)] transition-colors"
              >
                <RefreshCw className="w-3.5 h-3.5" />
                Retry
              </button>
              <button
                type="button"
                onClick={() => window.location.reload()}
                className="px-3 py-1.5 text-[11.5px] font-medium rounded-[5px] text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
              >
                Reload page
              </button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
