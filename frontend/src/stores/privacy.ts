import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface PrivacyState {
  hideValues: boolean
  toggleHideValues: () => void
}

const safeStorage = createJSONStorage(() => ({
  getItem: (key: string) => (typeof window !== 'undefined' ? localStorage.getItem(key) : null),
  setItem: (key: string, val: string) => {
    if (typeof window !== 'undefined') localStorage.setItem(key, val)
  },
  removeItem: (key: string) => {
    if (typeof window !== 'undefined') localStorage.removeItem(key)
  },
}))

export const usePrivacyStore = create<PrivacyState>()(
  persist(
    (set) => ({
      hideValues: false,
      toggleHideValues: () =>
        set((state) => ({ hideValues: !state.hideValues })),
    }),
    {
      name: 'inv_hide_values',
      storage: safeStorage,
    }
  )
)
