import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'

interface JsonEditorProps {
  value: Record<string, unknown> | string
  onChange?: (value: Record<string, unknown>) => void
  readOnly?: boolean
  className?: string
  height?: string
}

function JsonEditor({ value, onChange, readOnly, className, height = '200px' }: JsonEditorProps) {
  const [text, setText] = useState('')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (typeof value === 'string') {
      setText(value)
    } else {
      setText(JSON.stringify(value, null, 2))
    }
  }, [value])

  const handleChange = (newText: string) => {
    setText(newText)
    try {
      const parsed = JSON.parse(newText)
      setError(null)
      onChange?.(parsed)
    } catch (e) {
      setError((e as Error).message)
    }
  }

  return (
    <div className={cn('w-full', className)}>
      <textarea
        value={text}
        onChange={(e) => handleChange(e.target.value)}
        readOnly={readOnly}
        spellCheck={false}
        className={cn(
          'w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 resize-y',
          error && 'border-destructive focus-visible:ring-destructive'
        )}
        style={{ height, minHeight: '100px' }}
      />
      {error && <p className="mt-1 text-xs text-destructive">Invalid JSON: {error}</p>}
    </div>
  )
}

export { JsonEditor }
