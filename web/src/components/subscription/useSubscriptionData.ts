import { computed, onMounted, onScopeDispose, readonly, shallowRef } from 'vue'
import { localizedError } from '@/i18n'
import { createLatestRequest } from '@/utils/latestRequest'

// Live member data is never restored from persistent rendering snapshots.
export function useSubscriptionData<T>(fetcher: (signal: AbortSignal) => Promise<T>) {
  const value = shallowRef<T | null>(null)
  const loading = shallowRef(true)
  const refreshing = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const latest = createLatestRequest()
  let controller: AbortController | undefined
  let timer: ReturnType<typeof setTimeout> | undefined
  let disposed = false

  function schedule(): void {
    if (timer) clearTimeout(timer)
    if (!disposed && !document.hidden) timer = setTimeout(() => void refresh(), 60_000)
  }

  async function refresh(clear = false): Promise<void> {
    if (disposed) return
    if (timer) clearTimeout(timer)
    controller?.abort()
    controller = new AbortController()
    const signal = controller.signal
    const token = latest.begin()
    if (clear) value.value = null
    loading.value = value.value == null
    refreshing.value = true
    error.value = null
    try {
      const result = await fetcher(signal)
      if (latest.isCurrent(token) && !disposed) value.value = result
    } catch (caught) {
      if (latest.isCurrent(token) && !signal.aborted && !disposed) error.value = localizedError(caught, 'subscription.loadFailed')
    } finally {
      if (latest.isCurrent(token) && !disposed) {
        loading.value = false
        refreshing.value = false
        schedule()
      }
    }
  }

  function visibilityChanged(): void {
    if (document.hidden) { if (timer) clearTimeout(timer) }
    else void refresh()
  }

  onMounted(() => {
    document.addEventListener('visibilitychange', visibilityChanged)
    void refresh()
  })
  onScopeDispose(() => {
    disposed = true
    latest.dispose()
    controller?.abort()
    if (timer) clearTimeout(timer)
    document.removeEventListener('visibilitychange', visibilityChanged)
    value.value = null
  })
  return { value: readonly(value), loading: readonly(loading), refreshing: readonly(refreshing), error: readonly(error), hasData: computed(() => value.value != null), refresh }
}
