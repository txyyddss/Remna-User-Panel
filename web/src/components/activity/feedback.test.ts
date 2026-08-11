import { describe, expect, it } from 'vitest'

import { activityNotification, isSuccessfulBet } from './feedback'

describe('activity result feedback', () => {
  it('maps wins and completions to success while losses use warning feedback', () => {
    expect(activityNotification({ outcome: 'win' })).toBe('success')
    expect(activityNotification({ outcome: 'complete' })).toBe('success')
    expect(activityNotification({ outcome: 'loss' })).toBe('warning')
  })

  it.each([
    [{ kind: 'bet', outcome: 'win' }, true],
    [{ kind: 'bet', outcome: 'loss' }, false],
    [{ kind: 'check_in', outcome: 'complete' }, false],
    [{ kind: 'draw', outcome: 'complete' }, false],
    [null, false],
  ] as const)('classifies %o as fireworks=%s', (result, expected) => {
    expect(isSuccessfulBet(result)).toBe(expected)
  })
})
