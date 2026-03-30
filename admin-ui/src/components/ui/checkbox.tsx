import { cn } from '@/lib/utils'
import { Check, Minus } from 'lucide-react'

interface CheckboxProps {
  checked: boolean | 'indeterminate'
  onCheckedChange: (checked: boolean) => void
  disabled?: boolean
  className?: string
  id?: string
}

function Checkbox({ checked, onCheckedChange, disabled, className, id }: CheckboxProps) {
  return (
    <button
      id={id}
      role="checkbox"
      aria-checked={checked === 'indeterminate' ? 'mixed' : checked}
      disabled={disabled}
      onClick={() => onCheckedChange(checked === true ? false : true)}
      className={cn(
        'peer h-4 w-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 cursor-pointer flex items-center justify-center',
        (checked === true || checked === 'indeterminate') && 'bg-primary text-primary-foreground',
        className
      )}
    >
      {checked === true && <Check className="h-3 w-3" />}
      {checked === 'indeterminate' && <Minus className="h-3 w-3" />}
    </button>
  )
}

export { Checkbox }
