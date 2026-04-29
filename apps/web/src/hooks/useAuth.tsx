import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react'
import { getToken, setToken, clearToken } from '@/lib/api'
import type { Admin } from '@/types/auth'

interface AuthState {
  admin: Admin | null
  isAuthenticated: boolean
  login: (token: string, admin: Admin) => void
  logout: () => void
}

const AuthContext = createContext<AuthState | null>(null)

const ADMIN_KEY = 'typewriter_blog_admin'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [admin, setAdmin] = useState<Admin | null>(() => {
    try {
      const stored = localStorage.getItem(ADMIN_KEY)
      const token = getToken()
      if (stored && token) return JSON.parse(stored) as Admin
    } catch {
      // ignore
    }
    return null
  })

  const login = useCallback((token: string, adminData: Admin) => {
    setToken(token)
    localStorage.setItem(ADMIN_KEY, JSON.stringify(adminData))
    setAdmin(adminData)
  }, [])

  const logout = useCallback(() => {
    clearToken()
    localStorage.removeItem(ADMIN_KEY)
    setAdmin(null)
  }, [])

  // Check token still exists on mount (could have been cleared by a 401)
  useEffect(() => {
    if (!getToken()) {
      setAdmin(null)
      localStorage.removeItem(ADMIN_KEY)
    }
  }, [])

  return (
    <AuthContext.Provider value={{ admin, isAuthenticated: !!admin, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
