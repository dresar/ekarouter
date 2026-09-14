export interface LoadingSkeletonProps {
  className?: string
  count?: number
}

export function LoadingSkeleton({ className = 'h-5 w-full', count = 1 }: LoadingSkeletonProps) {
  return (
    <div className="space-y-2">
      {Array.from({ length: count }).map((_, index) => (
        <div
          key={index}
          className={`animate-pulse rounded-[4px] bg-[var(--border-strong)]/40 ${className}`}
        />
      ))}
    </div>
  )
}
