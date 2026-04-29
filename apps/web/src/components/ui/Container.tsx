import { cn } from '@/lib/utils'
import type { HTMLAttributes } from 'react'

interface ContainerProps extends HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg'
}

const sizeStyles = {
  sm: 'max-w-2xl',
  md: 'max-w-4xl',
  lg: 'max-w-6xl',
}

export function Container({ size = 'md', className, children, ...props }: ContainerProps) {
  return (
    <div className={cn('mx-auto w-full px-6', sizeStyles[size], className)} {...props}>
      {children}
    </div>
  )
}
