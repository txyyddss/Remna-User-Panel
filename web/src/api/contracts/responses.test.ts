import { afterEach, describe, expect, it, vi } from 'vitest'
import { request } from '../http'
import { isResponseCompatible } from './responses'

afterEach(() => vi.unstubAllGlobals())

describe('live response compatibility', () => {
  it('accepts fresh catalog contents without requiring a previous snapshot shape', () => {
    expect(isResponseCompatible('/api/v1/catalog', 'GET', { combos: [], addons: [], nodes: [], revision: 5 })).toBe(true)
    expect(isResponseCompatible('/api/v1/catalog', 'GET', { combos: null, addons: [], nodes: [] })).toBe(false)
  })

  it('rejects malformed success data before a loader can assign it', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ combos: null, addons: [], nodes: [] }), {
      status: 200, headers: { 'Content-Type': 'application/json' },
    })))
    await expect(request('/api/v1/catalog')).rejects.toMatchObject({ code: 'API_RESPONSE_INCOMPATIBLE', status: 502 })
  })
})
