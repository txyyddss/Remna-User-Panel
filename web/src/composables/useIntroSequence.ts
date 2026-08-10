import { computed, onScopeDispose, shallowRef, type MaybeRefOrGetter, toValue } from 'vue'

import type { OnboardingWelcomeMessage } from '@/api/features'

export interface IntroSequenceOptions {
  messages: MaybeRefOrGetter<readonly OnboardingWelcomeMessage[]>
  onComplete: () => void
}

export function useIntroSequence(options: IntroSequenceOptions) {
  const { onComplete } = options
  const index = shallowRef(0)
  const active = shallowRef(true)
  let timer: ReturnType<typeof setTimeout> | undefined

  const messages = computed(() => toValue(options.messages))
  const message = computed(() => messages.value[index.value]?.text ?? '')
  const progress = computed(() => messages.value.length ? ((index.value + 1) / messages.value.length) * 100 : 0)

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
      if (index.value >= messages.value.length - 1) {
        finish()
        return
      }
      index.value += 1
      schedule()
    }, messages.value[index.value]?.durationMs ?? 1800)
  }

  function start(): void {
    if (!active.value || messages.value.length === 0) return
    schedule()
  }

  function skip(): void {
    finish()
  }

  onScopeDispose(stopTimer)

  return { index, active, message, progress, start, skip }
}
