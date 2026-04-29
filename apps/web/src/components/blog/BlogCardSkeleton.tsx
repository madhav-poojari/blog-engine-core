import { Skeleton } from '@/components/ui/Skeleton'
import { Card } from '@/components/ui/Card'

export function BlogCardSkeleton() {
  return (
    <Card>
      <Skeleton className="w-full h-48 -mx-6 -mt-6 mb-4" />
      <div className="flex gap-2 mb-3">
        <Skeleton className="w-16 h-5" />
        <Skeleton className="w-12 h-5" />
      </div>
      <Skeleton className="w-3/4 h-6 mb-2" />
      <Skeleton className="w-full h-4 mb-1" />
      <Skeleton className="w-2/3 h-4 mb-4" />
      <div className="flex gap-4">
        <Skeleton className="w-24 h-3" />
        <Skeleton className="w-16 h-3" />
      </div>
    </Card>
  )
}
