import { useState, FormEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext.tsx'
import { Button } from '../../components/ui/Button.tsx'
import { Lock, User, AlertCircle } from 'lucide-react'

export function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const redirectPath = searchParams.get('redirect') || '/overview'

  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('admin1234')
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const handleUseDefault = () => {
    setUsername('admin')
    setPassword('admin1234')
    setError(null)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsLoading(true)

    try {
      await login(username, password)
      navigate(redirectPath, { replace: true })
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Invalid credentials'
      setError(message)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--bg-app)] bg-grid px-4">
      <div className="w-full max-w-sm">
        <div className="flex flex-col items-center text-center mb-6">
          <img src="/logo.svg" alt="EkaRouter" className="w-14 h-14 object-contain mb-3" />
          <h1 className="text-[20px] font-semibold text-[var(--text-primary)] tracking-tight">
            EkaRouter Console
          </h1>
          <p className="text-[12px] text-[var(--text-muted)] mt-1">
            Universal AI Gateway & Infrastructure Control Plane
          </p>
        </div>

        <div className="bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-[8px] p-6 shadow-xl">
          {error && (
            <div className="mb-4 p-3 rounded-[6px] bg-rose-950/20 border border-rose-600/30 flex items-center gap-2 text-[12px] text-rose-300">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1.5">
                Username
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[var(--text-muted)]">
                  <User className="w-3.5 h-3.5" />
                </span>
                <input
                  type="text"
                  required
                  autoFocus
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="admin"
                  className="w-full pl-8 pr-3 py-1.5 text-[12.5px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                />
              </div>
            </div>

            <div>
              <label className="block text-[11px] font-semibold uppercase tracking-wider text-[var(--text-secondary)] mb-1.5">
                Password
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-[var(--text-muted)]">
                  <Lock className="w-3.5 h-3.5" />
                </span>
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  autoComplete="current-password"
                  className="w-full pl-8 pr-3 py-1.5 text-[12.5px] rounded-[5px] bg-[var(--bg-panel)] border border-[var(--border-subtle)] text-[var(--text-primary)] focus:outline-none focus:border-[var(--brand-primary)]"
                />
              </div>
            </div>

            <Button
              type="submit"
              variant="primary"
              size="standard"
              isLoading={isLoading}
              className="w-full mt-2"
            >
              Sign In to Console
            </Button>
          </form>

          <div className="mt-4 pt-3 border-t border-[var(--border-subtle)] text-center">
            <button
              type="button"
              onClick={handleUseDefault}
              className="text-[11.5px] text-[var(--brand-text)] hover:underline"
            >
              Fill default development credentials
            </button>
          </div>
        </div>

        <div className="mt-6 text-center text-[11px] text-[var(--text-muted)] font-mono">
          EkaRouter Gateway Control Plane &bull; Port 8080
        </div>
      </div>
    </div>
  )
}
