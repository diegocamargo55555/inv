import { describe, it, expect, beforeEach } from 'vitest'
import { usePrivacyStore } from './privacy'

describe('Privacy Store', () => {
  beforeEach(() => {
    if (usePrivacyStore.getState().hideValues) {
      usePrivacyStore.getState().toggleHideValues()
    }
  })

  it('initializes with hideValues as false by default', () => {
    const state = usePrivacyStore.getState()
    expect(state.hideValues).toBe(false)
  })

  it('toggles hideValues state correctly', () => {
    usePrivacyStore.getState().toggleHideValues()
    expect(usePrivacyStore.getState().hideValues).toBe(true)

    usePrivacyStore.getState().toggleHideValues()
    expect(usePrivacyStore.getState().hideValues).toBe(false)
  })
})
