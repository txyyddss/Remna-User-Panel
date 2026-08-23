import { describe, expect, it } from 'vitest'

import { durationParts, multiplierFactor } from './format'

describe('compensation display conversion', () => {
  it('preserves basis-point precision and floors observed seconds', () => {
    expect(multiplierFactor(12_345)).toBe('1.23')
    expect(durationParts(7_259)).toEqual({ hours: 2, minutes: 0 })
  })
})
