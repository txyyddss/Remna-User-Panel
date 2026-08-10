import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useIntroSequence } from './useIntroSequence'

const messages = [
  { id: 'one', text: 'One', durationMs: 900 },
  { id: 'two', text: 'Two', durationMs: 900 },
  { id: 'three', text: 'Three', durationMs: 900 },
]

describe('onboarding intro sequence', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows each supplied message for 900ms and completes', () => {
    vi.useFakeTimers()
    const complete = vi.fn()
    const scope = effectScope()
    const sequence = scope.run(() => useIntroSequence({ messages, onComplete: complete }))!

    sequence.start()
    expect(sequence.message.value).toBe(messages[0].text)
    vi.advanceTimersByTime(900)
    expect(sequence.message.value).toBe(messages[1].text)
    vi.advanceTimersByTime(900)
    expect(sequence.message.value).toBe(messages[2].text)
    vi.advanceTimersByTime(900)
    expect(complete).toHaveBeenCalledOnce()
    scope.stop()
  })

  it('lets the user skip without a later duplicate completion', () => {
    vi.useFakeTimers()
    const complete = vi.fn()
    const scope = effectScope()
    const sequence = scope.run(() => useIntroSequence({ messages, onComplete: complete }))!
    sequence.start()
    sequence.skip()
    vi.runAllTimers()
    expect(complete).toHaveBeenCalledOnce()
    scope.stop()
  })
})
