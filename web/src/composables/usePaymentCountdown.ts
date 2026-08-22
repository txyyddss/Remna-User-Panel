import { computed, onScopeDispose, readonly, shallowRef, toValue, watch, type MaybeRefOrGetter } from 'vue'

export function usePaymentCountdown(expiresAt: MaybeRefOrGetter<string>) {
  const now = shallowRef(Date.now())
  let timer: ReturnType<typeof setInterval> | undefined

  const expiration = computed(() => Date.parse(toValue(expiresAt)))
  const remainingSeconds = computed(() => {
    if (!Number.isFinite(expiration.value)) return 0
    return Math.max(0, Math.ceil((expiration.value - now.value) / 1000))
  })
  const expired = computed(() => remainingSeconds.value === 0)
  const countdown = computed(() => {
    const minutes = Math.floor(remainingSeconds.value / 60)
    const seconds = remainingSeconds.value % 60
    return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  })

  function stop(): void {
    if (timer !== undefined) clearInterval(timer)
    timer = undefined
  }

  function start(): void {
    stop()
    now.value = Date.now()
    if (!expired.value) timer = setInterval(() => {
      now.value = Date.now()
      if (expired.value) stop()
    }, 1000)
  }

  watch(() => toValue(expiresAt), start, { immediate: true })
  onScopeDispose(stop)
  return { remainingSeconds: readonly(remainingSeconds), expired: readonly(expired), countdown: readonly(countdown) }
}
