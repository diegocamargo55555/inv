import { create } from 'zustand'

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
}

interface PrivacyState {
  hideValues: boolean
  toggleHideValues: () => void
}

export const usePrivacyStore = create<PrivacyState>((set) => ({
  hideValues: storage.get('inv_hide_values') === 'true',
  toggleHideValues: () =>
    set((state) => {
      const next = !state.hideValues
      storage.set('inv_hide_values', String(next))
      return { hideValues: next }
    }),
}))
