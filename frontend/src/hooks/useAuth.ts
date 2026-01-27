import { useEffect } from 'react'
import { useAuthStore } from '@/stores/auth-store'

export function useAuth() {
  const { state, user, error, initialize, login, register, logout, clearError } = useAuthStore()

  useEffect(() => {
    if (state === 'loading') {
      initialize()
    }
  }, [state, initialize])

  return {
    // State
    isLoading: state === 'loading',
    isAuthenticated: state === 'authenticated',
    needsRegistration: state === 'no-user',
    needsLogin: state === 'unauthenticated',
    user,
    error,

    // Actions
    login,
    register,
    logout,
    clearError,
  }
}
