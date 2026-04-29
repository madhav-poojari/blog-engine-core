import { cn } from '@/lib/utils'
import type { HTMLAttributes } from 'react'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  hover?: boolean
}

export function Card({ hover = false, className, children, ...props }: CardProps) {
  return (
    <div
      className={cn(
        'border border-border bg-cream p-6',
        hover && 'transition-all duration-200 hover:border-ink hover:shadow-sm cursor-pointer',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
