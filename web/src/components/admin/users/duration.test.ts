import { describe, expect, it } from 'vitest'

import { durationMinutes, maxDurationAmount, validDurationDraft } from './duration'

describe('bulk extension duration', () => {
  it.each([
    [{ amount: 17, unit: 'minutes' as const }, 17],
    [{ amount: 3, unit: 'hours' as const }, 180],
    [{ amount: 2, unit: 'days' as const }, 2880],
  ])('normalizes %o', (draft, expected) => {
    expect(durationMinutes(draft)).toBe(expected)
  })

  it('derives exact unit maximums', () => {
    expect(maxDurationAmount('minutes')).toBe(5_256_000)
    expect(maxDurationAmount('hours')).toBe(87_600)
    expect(maxDurationAmount('days')).toBe(3_650)
  })

  it('rejects fractional and out-of-range amounts', () => {
    expect(validDurationDraft({ amount: 1.5, unit: 'hours' })).toBe(false)
    expect(validDurationDraft({ amount: 87_601, unit: 'hours' })).toBe(false)
    expect(validDurationDraft({ amount: 87_600, unit: 'hours' })).toBe(true)
  })
})
