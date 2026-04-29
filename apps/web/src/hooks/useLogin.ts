import { useMutation } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useAuth } from './useAuth'
import type { LoginInput, LoginResult } from '@/types/auth'

export function useLogin() {
  const { login } = useAuth()

  return useMutation({
    mutationFn: (input: LoginInput) => api.post<LoginResult>('/auth/login', input),
    onSuccess: (data) => {
      login(data.token, data.admin)
    },
  })
}
