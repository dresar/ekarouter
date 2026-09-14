import { useState, ReactNode } from 'react'
import { Button } from './Button.tsx'

export interface InlineConfirmProps {
  trigger: ReactNode
  confirmText?: string
  cancelText?: string
  onConfirm: () => void | Promise<void>
  isLoading?: boolean
}

export function InlineConfirm({
  trigger,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onConfirm,
  isLoading = false,
}: InlineConfirmProps) {
  const [isConfirming, setIsConfirming] = useState(false)

  if (!isConfirming) {
    return <span onClick={() => setIsConfirming(true)}>{trigger}</span>
  }

  return (
    <div className="inline-flex items-center gap-1.5" onClick={(e) => e.stopPropagation()}>
      <Button
        variant="danger"
        size="compact"
        isLoading={isLoading}
        onClick={async () => {
          await onConfirm()
          setIsConfirming(false)
        }}
      >
        {confirmText}
      </Button>
      <Button
        variant="ghost"
        size="compact"
        disabled={isLoading}
        onClick={() => setIsConfirming(false)}
      >
        {cancelText}
      </Button>
    </div>
  )
}
