import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface PrivacyState {
  hideValues: boolean
  toggleHideValues: () => void
}

const storage = createJSONStorage(() => (typeof window !== 'undefined' ? localStorage : {
  getItem: () => null,
  setItem: () => {},
  removeItem: () => {},
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
      storage,
    }
  )
)
