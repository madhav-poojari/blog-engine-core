import { Link, useLocation } from '@tanstack/react-router'
import { Container } from '@/components/ui/Container'
import { Menu, X, PenLine, LogIn, LogOut } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/hooks/useAuth'

const navLinks = [
  { to: '/' as const, label: 'Home' },
  { to: '/blog' as const, label: 'Blog' },
]

export function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const location = useLocation()
  const { isAuthenticated, logout } = useAuth()

  return (
    <header className="sticky top-0 z-50 border-b border-border bg-cream/95 backdrop-blur-sm">
      <Container size="lg">
        <nav className="flex items-center justify-between h-16">
          <Link to="/" className="font-serif text-2xl font-bold text-ink tracking-tight no-underline">
            tb<span className="text-accent">.</span>
          </Link>

          {/* Desktop nav */}
          <div className="hidden md:flex items-center gap-8">
            <ul className="flex items-center gap-8 list-none m-0 p-0">
              {navLinks.map(({ to, label }) => (
                <li key={to}>
                  <Link
                    to={to}
                    className={cn(
                      'font-mono text-sm tracking-wider uppercase no-underline transition-colors duration-150',
                      location.pathname === to
                        ? 'text-ink border-b-2 border-accent pb-0.5'
                        : 'text-ink-muted hover:text-ink'
                    )}
                  >
                    {label}
                  </Link>
                </li>
              ))}
            </ul>

            <div className="w-px h-5 bg-border" />

            {isAuthenticated ? (
              <div className="flex items-center gap-4">
                <Link
                  to="/admin/editor"
                  className="inline-flex items-center gap-1.5 font-mono text-sm text-accent hover:text-accent-dark no-underline transition-colors"
                >
                  <PenLine size={14} />
                  New Post
                </Link>
                <button
                  onClick={logout}
                  className="inline-flex items-center gap-1.5 font-mono text-sm text-ink-muted hover:text-ink cursor-pointer bg-transparent border-none p-0 transition-colors"
                >
                  <LogOut size={14} />
                  Logout
                </button>
              </div>
            ) : (
              <Link
                to="/admin/login"
                className="inline-flex items-center gap-1.5 font-mono text-sm text-ink-muted hover:text-ink no-underline transition-colors"
              >
                <LogIn size={14} />
                Login
              </Link>
            )}
          </div>

          {/* Mobile menu button */}
          <button
            className="md:hidden p-2 text-ink cursor-pointer bg-transparent border-none"
            onClick={() => setMobileOpen(!mobileOpen)}
            aria-label="Toggle menu"
          >
            {mobileOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </nav>

        {/* Mobile nav */}
        {mobileOpen && (
          <div className="md:hidden border-t border-border-light py-4">
            <ul className="flex flex-col gap-4 list-none m-0 p-0">
              {navLinks.map(({ to, label }) => (
                <li key={to}>
                  <Link
                    to={to}
                    className={cn(
                      'font-mono text-sm tracking-wider uppercase no-underline block py-2',
                      location.pathname === to ? 'text-ink font-bold' : 'text-ink-muted'
                    )}
                    onClick={() => setMobileOpen(false)}
                  >
                    {label}
                  </Link>
                </li>
              ))}
              <li className="border-t border-border-light pt-4 mt-2">
                {isAuthenticated ? (
                  <div className="flex flex-col gap-3">
                    <Link
                      to="/admin/editor"
                      className="inline-flex items-center gap-1.5 font-mono text-sm text-accent no-underline"
                      onClick={() => setMobileOpen(false)}
                    >
                      <PenLine size={14} />
                      New Post
                    </Link>
                    <button
                      onClick={() => { logout(); setMobileOpen(false) }}
                      className="inline-flex items-center gap-1.5 font-mono text-sm text-ink-muted cursor-pointer bg-transparent border-none p-0"
                    >
                      <LogOut size={14} />
                      Logout
                    </button>
                  </div>
                ) : (
                  <Link
                    to="/admin/login"
                    className="inline-flex items-center gap-1.5 font-mono text-sm text-ink-muted no-underline"
                    onClick={() => setMobileOpen(false)}
                  >
                    <LogIn size={14} />
                    Login
                  </Link>
                )}
              </li>
            </ul>
          </div>
        )}
      </Container>
    </header>
  )
}
