import { cn } from '@/lib/utils'

interface CodeEditorProps {
  value: string
  onChange?: (value: string) => void
  language?: string
  readOnly?: boolean
  className?: string
  height?: string
  placeholder?: string
}

function CodeEditor({
  value,
  onChange,
  language = 'javascript',
  readOnly,
  className,
  height = '300px',
  placeholder,
}: CodeEditorProps) {
  return (
    <div className={cn('relative w-full rounded-md border border-input overflow-hidden', className)}>
      <div className="flex items-center justify-between px-3 py-1.5 bg-muted/50 border-b text-xs text-muted-foreground">
        <span>{language}</span>
      </div>
      <textarea
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        readOnly={readOnly}
        placeholder={placeholder}
        spellCheck={false}
        className="w-full bg-background px-4 py-3 text-sm font-mono leading-relaxed resize-y focus:outline-none text-foreground placeholder:text-muted-foreground"
        style={{ height, minHeight: '150px' }}
      />
    </div>
  )
}

export { CodeEditor }
