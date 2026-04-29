import { cn } from '@/lib/utils'
import type { ButtonHTMLAttributes } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
}

const variantStyles = {
  primary:
    'bg-ink text-cream hover:bg-ink-light border border-ink',
  secondary:
    'bg-cream text-ink border border-border hover:border-ink hover:bg-cream-dark',
  ghost:
    'bg-transparent text-ink hover:bg-cream-dark border border-transparent',
}

const sizeStyles = {
  sm: 'px-3 py-1.5 text-sm',
  md: 'px-5 py-2.5 text-base',
  lg: 'px-7 py-3 text-lg',
}

export function Button({
  variant = 'primary',
  size = 'md',
  className,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(
        'font-mono tracking-wide transition-colors duration-150 cursor-pointer inline-flex items-center justify-center',
        variantStyles[variant],
        sizeStyles[size],
        className
      )}
      {...props}
    >
      {children}
    </button>
  )
}
