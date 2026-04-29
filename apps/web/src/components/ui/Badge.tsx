import { cn } from '@/lib/utils'

interface BadgeProps {
  children: React.ReactNode
  variant?: 'default' | 'private' | 'public'
  className?: string
}

const variantStyles = {
  default: 'bg-cream-dark text-ink-light border-border',
  private: 'bg-accent/10 text-accent-dark border-accent/30',
  public: 'bg-green-50 text-green-800 border-green-200',
}

export function Badge({ children, variant = 'default', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center px-2.5 py-0.5 text-xs font-mono tracking-wider uppercase border',
        variantStyles[variant],
        className
      )}
    >
      {children}
    </span>
  )
}
