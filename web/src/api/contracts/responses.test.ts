import { describe, expect, it } from 'vitest'
import { isResponseCompatible } from './responses'

describe('rendering snapshot compatibility', () => {
  it('accepts fresh catalog contents without requiring a previous snapshot shape', () => {
    expect(isResponseCompatible('/api/v1/catalog', 'GET', {
      combos: [],
      addons: [],
      nodes: [],
      revision: 5,
    })).toBe(true)
  })

  it('rejects malformed cached catalog data before it reaches Vue state', () => {
    expect(isResponseCompatible('/api/v1/catalog', 'GET', {
      combos: null,
      addons: [],
      nodes: [],
    })).toBe(false)
  })

  it('does not claim a contract for an unknown route', () => {
    expect(isResponseCompatible('/api/v1/future-resource', 'GET', { any: 'shape' })).toBe(true)
  })
})
