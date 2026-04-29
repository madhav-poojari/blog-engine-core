import { Container } from '@/components/ui/Container'
import { Button } from '@/components/ui/Button'
import { Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'

export function NotFoundPage() {
  return (
    <Container size="sm">
      <div className="text-center py-24">
        <p className="text-8xl font-serif font-bold text-ink-muted/20 mb-4">404</p>
        <h1 className="text-2xl font-serif text-ink mb-3">Page not found</h1>
        <p className="text-ink-muted font-mono text-sm mb-8 max-w-md mx-auto">
          The page you're looking for has been lost in the mail.
          Perhaps it was never written.
        </p>
        <Link to="/">
          <Button variant="secondary">
            <ArrowLeft size={16} className="mr-2" />
            Go home
          </Button>
        </Link>
      </div>
    </Container>
  )
}
