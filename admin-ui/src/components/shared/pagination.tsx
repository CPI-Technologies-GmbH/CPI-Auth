import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-react'

interface PaginationProps {
  hasMore: boolean
  hasPrevious: boolean
  onNext: () => void
  onPrevious: () => void
  loading?: boolean
}

function Pagination({ hasMore, hasPrevious, onNext, onPrevious, loading }: PaginationProps) {
  return (
    <div className="flex items-center justify-end gap-2 mt-4">
      <Button
        variant="outline"
        size="sm"
        onClick={onPrevious}
        disabled={!hasPrevious || loading}
      >
        <ChevronLeft className="h-4 w-4 mr-1" />
        Previous
      </Button>
      <Button
        variant="outline"
        size="sm"
        onClick={onNext}
        disabled={!hasMore || loading}
      >
        Next
        <ChevronRight className="h-4 w-4 ml-1" />
      </Button>
    </div>
  )
}

export { Pagination }
