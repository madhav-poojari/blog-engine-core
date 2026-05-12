import { Container } from '@/components/ui/Container'
import { Button } from '@/components/ui/Button'
import { Link } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'

export function HomePage() {
  return (
    <Container size="sm">
      <div className="py-16 md:py-24">
        {/* Hero */}
        <div className="mb-16 text-center">
          <p className="text-sm font-mono text-ink-muted tracking-[0.3em] uppercase mb-4">
            Welcome to
          </p>
          <h1 className="text-5xl md:text-7xl font-serif font-bold text-ink mb-6 leading-tight">
            Madhava's musings<br />
            <span className="text-accent">Blog</span>
          </h1>
          <div className="w-24 border-t-2 border-accent mx-auto mb-6" />
          <p className="text-ink-light text-lg leading-relaxed max-w-lg mx-auto">
            i write about my thoughts , learning, travels, ideas and anything that can be typed 
          </p>
        </div>

        {/* About section */}
        <div className="border-t border-border pt-12 mb-12">
          <h2 className="text-2xl !font-serif text-ink mb-6 text-center">About Me</h2>
          <div className="space-y-4 text-ink-light leading-relaxed">
            <p>
              Im a software Developer , 
              Teacher and a human who makes mistakes 
            </p>
            <p>
              By day, I work with <strong className="text-ink">Go</strong>,{' '}
              <strong className="text-ink">React</strong>, and{' '}
              <strong className="text-ink">PostgreSQL</strong>. By night, 
               i teach chess , learn how to teach and write about anything that crosses my mind.

            </p>
            <p>
              This blog is my typewriter — a place for ideas still warm
              from the press.
            </p>
          </div>
        </div>

        {/* CTA */}
        <div className="text-center">
          <Link to="/blog">
            <Button variant="primary" size="lg">
              Read the blog
              <ArrowRight size={18} className="ml-2" />
            </Button>
          </Link>
        </div>
      </div>
    </Container>
  )
}
