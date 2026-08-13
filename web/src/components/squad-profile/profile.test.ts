import { describe, expect, it } from 'vitest'

import { isCountryCode, normalizeCountryCode } from './profile'

describe('squad profile display helpers', () => {
  it('normalizes and validates ISO alpha-2 country codes', () => {
    expect(normalizeCountryCode(' sg ')).toBe('SG')
    expect(isCountryCode('SG')).toBe(true)
    expect(isCountryCode('ZZ')).toBe(false)
  })
})
