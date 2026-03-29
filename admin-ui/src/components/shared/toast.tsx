import { useUIStore } from '@/stores/ui'
import { cn } from '@/lib/utils'
import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from 'lucide-react'

function Toaster() {
  const { toasts, removeToast } = useUIStore()

  if (toasts.length === 0) return null

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 max-w-sm">
      {toasts.map((toast) => {
        const icons = {
          default: <Info className="h-4 w-4 text-primary" />,
          success: <CheckCircle className="h-4 w-4 text-success" />,
          error: <AlertCircle className="h-4 w-4 text-destructive" />,
          warning: <AlertTriangle className="h-4 w-4 text-warning" />,
        }

        return (
          <div
            key={toast.id}
            className={cn(
              'flex items-start gap-3 rounded-lg border bg-card p-4 shadow-lg animate-in slide-in-from-right-full',
              toast.variant === 'error' && 'border-destructive/50',
              toast.variant === 'success' && 'border-success/50',
              toast.variant === 'warning' && 'border-warning/50'
            )}
          >
            {icons[toast.variant]}
            <div className="flex-1">
              <p className="text-sm font-semibold">{toast.title}</p>
              {toast.description && (
                <p className="mt-1 text-xs text-muted-foreground">{toast.description}</p>
              )}
            </div>
            <button
              onClick={() => removeToast(toast.id)}
              className="text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )
      })}
    </div>
  )
}

export { Toaster }
