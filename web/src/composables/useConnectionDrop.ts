import { computed, readonly, shallowRef } from 'vue'

import { memberOperationsApi } from '@/api/memberOperations'
import { localizedError } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { notifyHaptic } from '@/utils/telegram'
import { operationIsActive, useOperationReceipt } from './useOperationReceipt'

export function useConnectionDrop() {
  const busy = shallowRef(false)
  const mutationError = shallowRef<string | null>(null)
  const keys = new Map<string, string>()
  const operation = useOperationReceipt()

  const error = computed(() => mutationError.value ?? operation.error.value)
  const blocksMutations = computed(() => {
    const status = operation.receipt.value?.status
    return operationIsActive(status) || status === 'pending_review' || status === 'partial'
  })

  async function drop(handle: string): Promise<boolean> {
    if (busy.value || blocksMutations.value) return false
    busy.value = true
    mutationError.value = null
    try {
      const key = keys.get(handle) ?? createUuid()
      keys.set(handle, key)
      const receipt = await memberOperationsApi.dropConnection(handle, key)
      keys.delete(handle)
      operation.track(receipt)
      notifyHaptic('success')
      return true
    } catch (caught) {
      mutationError.value = localizedError(caught, 'connections.errors.dropFailed')
      notifyHaptic('error')
      return false
    } finally {
      busy.value = false
    }
  }

  function reset(): void {
    mutationError.value = null
    operation.reset()
  }

  return {
    receipt: operation.receipt,
    polling: operation.polling,
    checking: operation.checking,
    terminal: operation.terminal,
    busy: readonly(busy),
    error,
    blocksMutations,
    drop,
    refresh: operation.refresh,
    reset,
  }
}
