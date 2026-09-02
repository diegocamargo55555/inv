import { create } from 'zustand'
import { User } from '../types'

const storage = {
  get: (key: string) => {
    try {
      return typeof window !== 'undefined' ? localStorage.getItem(key) : null
    } catch {
      return null
    }
  },
  set: (key: string, val: string) => {
    try {
      if (typeof window !== 'undefined') localStorage.setItem(key, val)
    } catch {}
  },
  remove: (key: string) => {
    try {
      if (typeof window !== 'undefined') localStorage.removeItem(key)
    } catch {}
  },
}

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  setAuth: (user: User, accessToken: string, refreshToken: string) => void
  setTokens: (accessToken: string, refreshToken: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: JSON.parse(storage.get('inv_user') || 'null'),
  accessToken: storage.get('inv_access_token'),
  refreshToken: storage.get('inv_refresh_token'),
  isAuthenticated: !!storage.get('inv_access_token'),

  setAuth: (user, accessToken, refreshToken) => {
    storage.set('inv_user', JSON.stringify(user))
    storage.set('inv_access_token', accessToken)
    storage.set('inv_refresh_token', refreshToken)
    set({ user, accessToken, refreshToken, isAuthenticated: true })
  },

  setTokens: (accessToken, refreshToken) => {
    storage.set('inv_access_token', accessToken)
    storage.set('inv_refresh_token', refreshToken)
    set({ accessToken, refreshToken, isAuthenticated: true })
  },

  logout: () => {
    storage.remove('inv_user')
    storage.remove('inv_access_token')
    storage.remove('inv_refresh_token')
    set({ user: null, accessToken: null, refreshToken: null, isAuthenticated: false })
  },
}))
