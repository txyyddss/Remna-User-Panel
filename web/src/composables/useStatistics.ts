import { onMounted, onUnmounted, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { StatisticsNodesSnapshot, StatisticsSnapshot } from '@/api/types'
import { localizedError } from '@/i18n'

const nodeRefreshMilliseconds = 10_000

export function useStatistics() {
  const snapshot = shallowRef<StatisticsSnapshot | null>(null)
  const nodeSnapshot = shallowRef<StatisticsNodesSnapshot | null>(null)
  const loading = shallowRef(true)
  const refreshing = shallowRef(false)
  const nodesLoading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const nodesError = shallowRef<string | null>(null)
  let timer: number | undefined
  let disposed = false
  let requestVersion = 0

  async function loadNodes(): Promise<void> {
    if (nodesLoading.value || disposed) return
    nodesLoading.value = true
    nodesError.value = null
    try {
      const response = await api.getStatisticsNodes()
      if (!disposed) nodeSnapshot.value = response
    } catch (caught) {
      if (!disposed) nodesError.value = localizedError(caught, 'statistics.nodesLoadFailed')
    } finally {
      nodesLoading.value = false
    }
  }

  async function load(options: { quiet?: boolean } = {}): Promise<void> {
    const version = ++requestVersion
    if (options.quiet) refreshing.value = true
    else loading.value = true
    error.value = null
    const [statisticsResult, nodesResult] = await Promise.allSettled([
      api.getStatistics(),
      api.getStatisticsNodes(),
    ])
    if (disposed || version !== requestVersion) return
    if (statisticsResult.status === 'fulfilled') snapshot.value = statisticsResult.value
    else error.value = localizedError(statisticsResult.reason, 'statistics.loadFailed')
    if (nodesResult.status === 'fulfilled') {
      nodeSnapshot.value = nodesResult.value
      nodesError.value = null
    } else {
      nodesError.value = localizedError(nodesResult.reason, 'statistics.nodesLoadFailed')
    }
    loading.value = false
    refreshing.value = false
  }

  onMounted(() => {
    void load()
    timer = window.setInterval(() => {
      if (!document.hidden) void loadNodes()
    }, nodeRefreshMilliseconds)
  })

  onUnmounted(() => {
    disposed = true
    requestVersion += 1
    if (timer !== undefined) window.clearInterval(timer)
  })

  return {
    snapshot: readonly(snapshot),
    nodeSnapshot: readonly(nodeSnapshot),
    loading: readonly(loading),
    refreshing: readonly(refreshing),
    nodesLoading: readonly(nodesLoading),
    error: readonly(error),
    nodesError: readonly(nodesError),
    load,
  }
}
