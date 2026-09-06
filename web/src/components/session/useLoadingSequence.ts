import { computed, onScopeDispose, ref, watch } from 'vue'

import type { SessionStatus } from '@/stores/session'

// Matches the word replacement's delay + duration + a short settled hold in session-01.css.
const INTRO_DURATION_MS = 1900

export function useLoadingSequence(status: () => SessionStatus) {
  const motion = window.matchMedia('(prefers-reduced-motion: reduce)')
  const reducedMotion = ref(motion.matches)
  const introComplete = ref(true)
  const loading = computed(() => status() === 'idle' || status() === 'loading')
  let timer: ReturnType<typeof setTimeout> | undefined

  function clearTimer(): void {
    clearTimeout(timer)
    timer = undefined
  }

  function syncMotion(): void {
    reducedMotion.value = motion.matches
    if (motion.matches) {
      clearTimer()
      introComplete.value = true
    }
  }

  watch(loading, (active) => {
    if (!active) return
    clearTimer()
    introComplete.value = reducedMotion.value
    if (!introComplete.value) {
      timer = setTimeout(() => { introComplete.value = true }, INTRO_DURATION_MS)
    }
  }, { immediate: true })

  const showing = computed(() => loading.value || (
    status() === 'ready' && !introComplete.value && !reducedMotion.value
  ))

  motion.addEventListener('change', syncMotion)
  onScopeDispose(() => {
    clearTimer()
    motion.removeEventListener('change', syncMotion)
  })

  return { loading, showing }
}
