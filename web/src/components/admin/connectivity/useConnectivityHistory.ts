import { onScopeDispose, shallowRef, watch } from 'vue'
import { connectivityApi, type ConnectivityAttempt } from '@/api/connectivity'
import { connectivityError } from './presentation'

export function useConnectivityHistory() {
  const hostUuid = shallowRef('')
  const items = shallowRef<ConnectivityAttempt[]>([])
  const nextCursor = shallowRef<string | null>(null)
  const loading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  let generation = 0
  let controller: AbortController | undefined
  let disposed = false

  async function load(more = false): Promise<void> {
    if (disposed || (more && (loading.value || !nextCursor.value))) return
    controller?.abort()
    controller = new AbortController()
    const current = ++generation
    const cursor = more ? nextCursor.value ?? undefined : undefined
    if (!more) { items.value = []; nextCursor.value = null }
    loading.value = true
    error.value = null
    try {
      const page = await connectivityApi.history(hostUuid.value, cursor, controller.signal)
      if (disposed || current !== generation) return
      const merged = more ? [...items.value, ...page.items] : page.items
      items.value = merged.filter((item, index) => merged.findIndex(other => other.id === item.id) === index)
      nextCursor.value = page.nextCursor
    } catch (caught) {
      if (!disposed && current === generation && !controller.signal.aborted) error.value = connectivityError(caught, 'hostConnectivity.historyFailed')
    } finally {
      if (!disposed && current === generation) loading.value = false
    }
  }

  watch(hostUuid, () => void load(), { immediate: true })
  onScopeDispose(() => { disposed = true; generation++; controller?.abort() })
  return { hostUuid, items, nextCursor, loading, error, load }
}
