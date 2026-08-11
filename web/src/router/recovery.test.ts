import { afterEach, describe, expect, it } from 'vitest'

import { beginRouteRecovery, completeRouteRecovery, isRouteChunkError, pendingRouteRecoveryPath } from './recovery'

describe('route recovery', () => {
  afterEach(() => sessionStorage.removeItem('txc-route-recovery'))

  it('allows one reload attempt for a failed route chunk', () => {
    expect(beginRouteRecovery('/catalog')).toBe(true)
    expect(pendingRouteRecoveryPath()).toBe('/catalog')
    expect(beginRouteRecovery('/catalog')).toBe(false)
  })

  it('clears recovery after the route succeeds', () => {
    expect(beginRouteRecovery('/activity')).toBe(true)
    completeRouteRecovery('/activity')
    expect(pendingRouteRecoveryPath()).toBeUndefined()
  })

  it('recognizes browser module loading failures only', () => {
    expect(isRouteChunkError(new TypeError('Failed to fetch dynamically imported module'))).toBe(true)
    expect(isRouteChunkError(new Error('API request failed'))).toBe(false)
  })
})
