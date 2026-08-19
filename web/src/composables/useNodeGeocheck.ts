import { computed, getCurrentInstance, onUnmounted, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { StatisticsNode, StatisticsNodeGeocheck } from '@/api/types'
import { localizedError } from '@/i18n'

export function useNodeGeocheck() {
  const selectedNode = shallowRef<StatisticsNode | null>(null)
  const result = shallowRef<StatisticsNodeGeocheck | null>(null)
  const loading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  let requestVersion = 0

  const isOpen = computed({
    get: () => selectedNode.value !== null,
    set: (open: boolean) => { if (!open) close() },
  })

  async function show(node: StatisticsNode): Promise<void> {
    const version = ++requestVersion
    selectedNode.value = node
    result.value = null
    error.value = null
    loading.value = true
    try {
      const response = await api.getNodeGeocheck(node.uuid)
      if (version === requestVersion) result.value = response
    } catch (caught) {
      if (version === requestVersion) error.value = localizedError(caught, 'statistics.geocheck.unavailable')
    } finally {
      if (version === requestVersion) loading.value = false
    }
  }

  function close(): void {
    requestVersion += 1
    selectedNode.value = null
    result.value = null
    error.value = null
    loading.value = false
  }

  if (getCurrentInstance()) onUnmounted(() => { requestVersion += 1 })

  return { selectedNode: readonly(selectedNode), result: readonly(result), loading: readonly(loading), error: readonly(error), isOpen, show, close }
}
