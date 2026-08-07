import { computed, onScopeDispose, shallowRef } from 'vue'

export const onboardingMessages = [
  'Hi, how are you?',
  'Welcome to TX Carpool',
  'Just take you several seconds to complete',
] as const

export interface IntroSequenceOptions {
  duration?: number
  onComplete: () => void
}

export function useIntroSequence(options: IntroSequenceOptions) {
  const { duration = 900, onComplete } = options
  const index = shallowRef(0)
  const active = shallowRef(true)
  let timer: ReturnType<typeof setTimeout> | undefined

  const message = computed(() => onboardingMessages[index.value] ?? onboardingMessages[0])
  const progress = computed(() => ((index.value + 1) / onboardingMessages.length) * 100)

  function stopTimer(): void {
    if (timer !== undefined) clearTimeout(timer)
    timer = undefined
  }

  function finish(): void {
    if (!active.value) return
    stopTimer()
    active.value = false
    onComplete()
  }

  function schedule(): void {
    stopTimer()
    timer = setTimeout(() => {
      if (index.value >= onboardingMessages.length - 1) {
        finish()
        return
      }
      index.value += 1
      schedule()
    }, duration)
  }

  function start(): void {
    if (!active.value) return
    schedule()
  }

  function skip(): void {
    finish()
  }

  onScopeDispose(stopTimer)

  return { index, active, message, progress, start, skip }
}
