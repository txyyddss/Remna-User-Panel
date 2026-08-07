import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { onboardingMessages, useIntroSequence } from './useIntroSequence'

describe('onboarding intro sequence', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows each supplied message for 900ms and completes', () => {
    vi.useFakeTimers()
    const complete = vi.fn()
    const scope = effectScope()
    const sequence = scope.run(() => useIntroSequence({ duration: 900, onComplete: complete }))!

    sequence.start()
    expect(sequence.message.value).toBe(onboardingMessages[0])
    vi.advanceTimersByTime(900)
    expect(sequence.message.value).toBe(onboardingMessages[1])
    vi.advanceTimersByTime(900)
    expect(sequence.message.value).toBe(onboardingMessages[2])
    vi.advanceTimersByTime(900)
    expect(complete).toHaveBeenCalledOnce()
    scope.stop()
  })

  it('lets the user skip without a later duplicate completion', () => {
    vi.useFakeTimers()
    const complete = vi.fn()
    const scope = effectScope()
    const sequence = scope.run(() => useIntroSequence({ onComplete: complete }))!
    sequence.start()
    sequence.skip()
    vi.runAllTimers()
    expect(complete).toHaveBeenCalledOnce()
    scope.stop()
  })
})
