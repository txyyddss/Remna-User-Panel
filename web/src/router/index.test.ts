import { describe, expect, it } from 'vitest'

import router from './index'

describe('member routes', () => {
  it('registers the Community route without adding primary navigation state', () => {
    const route = router.getRoutes().find((entry) => entry.name === 'community')

    expect(route?.path).toBe('/community')
  })
})
