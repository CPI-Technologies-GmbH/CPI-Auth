import { Link } from 'react-router-dom'
import { ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

interface BreadcrumbItem {
  label: string
  href?: string
}

interface BreadcrumbsProps {
  items: BreadcrumbItem[]
  className?: string
}

function Breadcrumbs({ items, className }: BreadcrumbsProps) {
  return (
    <nav aria-label="Breadcrumb" className={cn('flex items-center space-x-1 text-sm text-muted-foreground', className)}>
      {items.map((item, index) => (
        <div key={index} className="flex items-center">
          {index > 0 && <ChevronRight className="mx-1 h-4 w-4" />}
          {item.href && index < items.length - 1 ? (
            <Link to={item.href} className="hover:text-foreground transition-colors">
              {item.label}
            </Link>
          ) : (
            <span className={cn(index === items.length - 1 && 'text-foreground font-medium')}>
              {item.label}
            </span>
          )}
        </div>
      ))}
    </nav>
  )
}

export { Breadcrumbs }
export type { BreadcrumbItem }
