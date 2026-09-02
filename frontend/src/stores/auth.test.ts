import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore } from './auth'
import { User } from '../types'

describe('Auth Store', () => {
  beforeEach(() => {
    useAuthStore.getState().logout()
  })

  it('initializes with unauthenticated state', () => {
    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
    expect(state.accessToken).toBeNull()
  })

  it('sets authentication data correctly on setAuth', () => {
    const mockUser: User = {
      id: '123e4567-e89b-12d3-a456-426614174000',
      name: 'Lucas Silva',
      email: 'lucas@example.com',
      base_currency: 'BRL',
    }

    useAuthStore.getState().setAuth(mockUser, 'access-token-123', 'refresh-token-456')

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.user?.email).toBe('lucas@example.com')
    expect(state.accessToken).toBe('access-token-123')
    expect(state.refreshToken).toBe('refresh-token-456')
  })

  it('clears state on logout', () => {
    const mockUser: User = {
      id: '123',
      name: 'Test',
      email: 'test@example.com',
      base_currency: 'BRL',
    }
    useAuthStore.getState().setAuth(mockUser, 'acc', 'ref')
    expect(useAuthStore.getState().isAuthenticated).toBe(true)

    useAuthStore.getState().logout()

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
    expect(state.accessToken).toBeNull()
  })
})
