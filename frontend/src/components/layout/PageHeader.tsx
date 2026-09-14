import { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { ChevronRight } from 'lucide-react'

export interface BreadcrumbItem {
  label: string
  to?: string
}

export interface PageHeaderProps {
  title: string
  description?: string
  breadcrumbs?: BreadcrumbItem[]
  actions?: ReactNode
  metadata?: ReactNode
}

export function PageHeader({ title, description, breadcrumbs, actions, metadata }: PageHeaderProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3 pb-4 mb-4 border-b border-[var(--border-subtle)]">
      <div>
        {breadcrumbs && breadcrumbs.length > 0 && (
          <nav className="flex items-center gap-1.5 text-[11px] text-[var(--text-muted)] mb-1">
            {breadcrumbs.map((crumb, idx) => (
              <span key={idx} className="flex items-center gap-1.5">
                {idx > 0 && <ChevronRight className="w-3 h-3 text-[var(--text-muted)]" />}
                {crumb.to ? (
                  <Link
                    to={crumb.to}
                    className="text-[var(--brand-text)] hover:underline transition-colors"
                  >
                    {crumb.label}
                  </Link>
                ) : (
                  <span className="text-[var(--text-secondary)] font-medium">
                    {crumb.label}
                  </span>
                )}
              </span>
            ))}
          </nav>
        )}
        <h1 className="text-[18px] font-semibold text-[var(--text-primary)] tracking-tight">
          {title}
        </h1>
        {description && (
          <p className="text-[12px] text-[var(--text-secondary)] mt-0.5 max-w-3xl leading-normal">
            {description}
          </p>
        )}
      </div>

      <div className="flex flex-col items-end gap-2 shrink-0">
        {actions && <div className="flex items-center gap-2">{actions}</div>}
        {metadata && (
          <div className="text-[11px] text-[var(--text-muted)] font-mono flex items-center gap-1.5">
            {metadata}
          </div>
        )}
      </div>
    </div>
  )
}
