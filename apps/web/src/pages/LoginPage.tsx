import { useState } from 'react'
import { Container } from '@/components/ui/Container'
import { Button } from '@/components/ui/Button'
import { useLogin } from '@/hooks/useLogin'
import { useNavigate } from '@tanstack/react-router'
import { LogIn } from 'lucide-react'

export function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const loginMutation = useLogin()
  const navigate = useNavigate()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!username.trim() || !password.trim()) return

    loginMutation.mutate(
      { username: username.trim(), password: password.trim() },
      {
        onSuccess: () => {
          navigate({ to: '/admin/editor' })
        },
      },
    )
  }

  return (
    <Container size="sm">
      <div className="max-w-sm mx-auto py-20">
        <div className="text-center mb-10">
          <h1 className="text-3xl font-serif font-bold text-ink mb-2">Admin Login</h1>
          <div className="w-12 border-t-2 border-accent mx-auto" />
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label htmlFor="username" className="block font-mono text-xs text-ink-muted uppercase tracking-wider mb-2">
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
              className="w-full font-mono text-sm text-ink bg-transparent border border-border p-3 outline-none focus:border-accent transition-colors"
            />
          </div>

          <div>
            <label htmlFor="password" className="block font-mono text-xs text-ink-muted uppercase tracking-wider mb-2">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              className="w-full font-mono text-sm text-ink bg-transparent border border-border p-3 outline-none focus:border-accent transition-colors"
            />
          </div>

          {loginMutation.isError && (
            <p className="text-red-700 font-mono text-sm">
              Invalid credentials. Please try again.
            </p>
          )}

          <Button
            type="submit"
            variant="primary"
            size="md"
            className="w-full"
            disabled={!username.trim() || !password.trim() || loginMutation.isPending}
          >
            <LogIn size={16} className="mr-2" />
            {loginMutation.isPending ? 'Logging in...' : 'Log in'}
          </Button>
        </form>
      </div>
    </Container>
  )
}
