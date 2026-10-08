import { computed, onMounted, onScopeDispose, shallowRef } from 'vue'
import { connectivityApi, type ConnectivitySnapshot } from '@/api/connectivity'
import { connectivityError } from './presentation'

export function useConnectivityMonitor() {
  const snapshot = shallowRef<ConnectivitySnapshot | null>(null)
  const loading = shallowRef(true)
  const checking = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const active = computed(() => snapshot.value?.run?.status === 'running')
  let disposed = false
  let generation = 0
  let timer: ReturnType<typeof setTimeout> | undefined
  let controller: AbortController | undefined
  let checkController: AbortController | undefined

  async function refresh(): Promise<void> {
    if (disposed) return
    if (timer) clearTimeout(timer)
    controller?.abort()
    controller = new AbortController()
    const current = ++generation
    try {
      const result = await connectivityApi.snapshot(controller.signal)
      if (disposed || current !== generation) return
      snapshot.value = result
      error.value = null
    } catch (caught) {
      if (!disposed && current === generation && !controller.signal.aborted) error.value = connectivityError(caught, 'hostConnectivity.loadFailed')
    } finally {
      if (!disposed && current === generation) {
        loading.value = false
        timer = setTimeout(() => void refresh(), active.value ? 2000 : 15000)
      }
    }
  }

  async function check(): Promise<void> {
    if (disposed || checking.value) return
    checking.value = true
    error.value = null
    checkController = new AbortController()
    try {
      const run = await connectivityApi.check(checkController.signal)
      if (disposed) return
      if (snapshot.value) snapshot.value = { ...snapshot.value, run }
      await refresh()
    } catch (caught) {
      if (!disposed && !checkController.signal.aborted) error.value = connectivityError(caught, 'hostConnectivity.checkFailed')
    } finally {
      if (!disposed) checking.value = false
    }
  }

  onMounted(() => void refresh())
  onScopeDispose(() => {
    disposed = true
    generation++
    controller?.abort()
    checkController?.abort()
    if (timer) clearTimeout(timer)
  })
  return { snapshot, loading, checking, error, active, refresh, check }
}
