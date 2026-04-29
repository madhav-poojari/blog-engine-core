import { Container } from '@/components/ui/Container'
import { socialLinks } from '@/data/socialLinks'

export function Footer() {
  return (
    <footer className="border-t border-border mt-auto">
      <Container size="lg">
        <div className="py-10 flex flex-col items-center gap-6">
          {/* Social links */}
          <div className="flex items-center gap-5">
            {socialLinks.map(({ name, url, icon: Icon }) => (
              <a
                key={name}
                href={url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-ink-muted hover:text-accent transition-colors duration-150"
                aria-label={name}
              >
                <Icon size={20} strokeWidth={1.5} />
              </a>
            ))}
          </div>

          {/* Divider line */}
          <div className="w-16 border-t border-border" />

          {/* Copyright */}
          <p className="text-xs text-ink-muted font-mono tracking-wider">
            &copy; {new Date().getFullYear()} &middot; Built with a typewriter &amp; some code
          </p>
        </div>
      </Container>
    </footer>
  )
}
