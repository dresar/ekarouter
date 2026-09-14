import { ReactNode } from 'react'

export interface Column<T> {
  key: string
  title: string
  render?: (item: T) => ReactNode
  className?: string
  width?: string
}

export interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  isLoading?: boolean
  emptyMessage?: string
  keyExtractor: (item: T) => string | number
  onRowClick?: (item: T) => void
}

export function DataTable<T>({
  columns,
  data,
  isLoading = false,
  emptyMessage = 'No data available',
  keyExtractor,
  onRowClick,
}: DataTableProps<T>) {
  const rows: T[] = Array.isArray(data) ? data : []

  return (
    <div className="w-full overflow-x-auto rounded-[8px] border border-[var(--border-subtle)] bg-[var(--bg-card)]">
      <table className="w-full text-left border-collapse text-[12.5px]">
        <thead>
          <tr className="border-b border-[var(--border-subtle)] bg-[var(--bg-panel)]/60">
            {columns.map((col) => (
              <th
                key={col.key}
                style={{ width: col.width }}
                className={`py-2 px-3.5 text-[10.5px] font-semibold uppercase tracking-wider text-[var(--text-muted)] select-none ${
                  col.className || ''
                }`}
              >
                {col.title}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-[var(--border-subtle)]">
          {isLoading ? (
            Array.from({ length: 5 }).map((_, rowIndex) => (
              <tr key={`skeleton-${rowIndex}`} className="animate-pulse">
                {columns.map((col) => (
                  <td key={`skeleton-col-${col.key}`} className="py-2.5 px-3.5">
                    <div className="h-3.5 bg-[var(--border-strong)]/40 rounded-[3px] w-3/4" />
                  </td>
                ))}
              </tr>
            ))
          ) : rows.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="py-12 text-center text-[var(--text-muted)] text-[12.5px]">
                {emptyMessage}
              </td>
            </tr>
          ) : (
            rows.map((item) => {
              const rowKey = keyExtractor(item)
              return (
                <tr
                  key={rowKey}
                  onClick={() => onRowClick && onRowClick(item)}
                  className={`transition-colors hover:bg-[var(--bg-panel)]/40 ${
                    onRowClick ? 'cursor-pointer' : ''
                  }`}
                >
                  {columns.map((col) => {
                    const content = col.render
                      ? col.render(item)
                      : (item as Record<string, unknown>)[col.key] as ReactNode
                    return (
                      <td key={`${rowKey}-${col.key}`} className={`py-2 px-3.5 ${col.className || ''}`}>
                        {content}
                      </td>
                    )
                  })}
                </tr>
              )
            })
          )}
        </tbody>
      </table>
    </div>
  )
}
