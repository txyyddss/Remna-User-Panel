import { describe, expect, it } from 'vitest'

import { onboardingStepAfterMembership } from './onboardingState'

describe('onboardingStepAfterMembership', () => {
  it('keeps a returning user on the revised agreement', () => {
    expect(onboardingStepAfterMembership('agreement', true)).toBe('agreement')
  })

  it('keeps completed users out of username reservation', () => {
    expect(onboardingStepAfterMembership('complete', true)).toBe('complete')
  })

  it('advances a new member to username reservation', () => {
    expect(onboardingStepAfterMembership('membership', true)).toBe('username')
  })
})
