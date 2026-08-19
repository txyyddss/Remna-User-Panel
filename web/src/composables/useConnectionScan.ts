import { computed, onMounted, onUnmounted, readonly, shallowRef } from 'vue'

import { memberOperationsApi } from '@/api/memberOperations'
import type { ConnectionScan } from '@/api/types'
import { localizedError } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'

export function useConnectionScan(intervalMilliseconds = 3000) {
  const scan = shallowRef<ConnectionScan | null>(null)
  const loading = shallowRef(true)
  const polling = shallowRef(false)
  const error = shallowRef<string | null>(null)
  let requestKey: string | undefined
  let timer: number | undefined
  let version = 0

  const completed = computed(() => scan.value?.isCompleted === true)
  const failed = computed(() => scan.value?.isFailed === true)
  const nodes = computed(() => scan.value?.nodes ?? [])
  const progressPercent = computed(() => scan.value?.progressPercent ?? 0)

  function clearTimer(): void {
    if (timer !== undefined) window.clearTimeout(timer)
    timer = undefined
  }

  function schedule(expectedVersion: number): void {
    clearTimer()
    timer = window.setTimeout(() => void poll(expectedVersion), intervalMilliseconds)
  }

  function accept(next: ConnectionScan, expectedVersion: number): void {
    scan.value = next
    polling.value = !next.isCompleted && !next.isFailed
    if (polling.value) schedule(expectedVersion)
  }

  async function poll(expectedVersion = version): Promise<void> {
    const scanId = scan.value?.id
    if (!scanId || expectedVersion !== version) return
    try {
      const next = await memberOperationsApi.pollConnections(scanId)
      if (expectedVersion !== version) return
      error.value = null
      accept(next, expectedVersion)
    } catch (caught) {
      if (expectedVersion !== version) return
      polling.value = false
      error.value = localizedError(caught, 'connections.errors.pollFailed')
    }
  }

  async function start(): Promise<void> {
    const expectedVersion = ++version
    clearTimer()
    loading.value = true
    polling.value = false
    error.value = null
    try {
      requestKey ??= createUuid()
      const next = await memberOperationsApi.requestConnections(requestKey)
      if (expectedVersion !== version) return
      requestKey = undefined
      accept(next, expectedVersion)
    } catch (caught) {
      if (expectedVersion === version) error.value = localizedError(caught, 'connections.errors.scanFailed')
    } finally {
      if (expectedVersion === version) loading.value = false
    }
  }

  function retry(): Promise<void> {
    if (scan.value && !scan.value.isCompleted && !scan.value.isFailed) {
      polling.value = true
      error.value = null
      return poll(version)
    }
    return start()
  }

  function restart(): Promise<void> {
    if (!scan.value) return start()
    if (!scan.value.isCompleted && !scan.value.isFailed) return retry()
    requestKey = undefined
    scan.value = null
    return start()
  }

  onMounted(() => void start())
  onUnmounted(() => {
    version += 1
    clearTimer()
  })

  return {
    scan: readonly(scan),
    nodes,
    progressPercent,
    loading: readonly(loading),
    polling: readonly(polling),
    error: readonly(error),
    completed,
    failed,
    retry,
    restart,
  }
}
