import { computed, onMounted, readonly, shallowReadonly, shallowRef } from 'vue'

import { ApiError } from '@/api/http'
import { memberOperationsApi } from '@/api/memberOperations'
import type { IPBlock } from '@/api/types'
import { localizedError } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { notifyHaptic } from '@/utils/telegram'
import { operationIsActive, useOperationReceipt } from './useOperationReceipt'

export type ConnectionBlockAction = 'block' | 'unblock'

export function useConnectionBlocks() {
  const items = shallowRef<readonly IPBlock[]>([])
  const loading = shallowRef(true)
  const loadError = shallowRef<string | null>(null)
  const mutationError = shallowRef<string | null>(null)
  const busyAction = shallowRef<string | null>(null)
  const action = shallowRef<ConnectionBlockAction>('block')
  const keys = new Map<string, string>()
  const operation = useOperationReceipt()

  const error = computed(() => mutationError.value ?? operation.error.value)
  const mutationActive = computed(() => busyAction.value !== null || operationIsActive(operation.receipt.value?.status))

  async function load(): Promise<void> {
    loading.value = true
    loadError.value = null
    try {
      items.value = (await memberOperationsApi.listIPBlocks()).items
    } catch (caught) {
      loadError.value = localizedError(caught, 'connections.errors.blocksFailed')
    } finally {
      loading.value = false
    }
  }

  async function mutate(kind: ConnectionBlockAction, target: string, task: (key: string) => Promise<Parameters<typeof operation.track>[0]>): Promise<boolean> {
    if (mutationActive.value) return false
    const keyTarget = `${kind}:${target}`
    const key = keys.get(keyTarget) ?? createUuid()
    keys.set(keyTarget, key)
    busyAction.value = keyTarget
    mutationError.value = null
    action.value = kind
    try {
      const receipt = await task(key)
      keys.delete(keyTarget)
      operation.track(receipt)
      await load()
      notifyHaptic('success')
      return true
    } catch (caught) {
      if (caught instanceof ApiError) keys.delete(keyTarget)
      mutationError.value = localizedError(caught, kind === 'block' ? 'connections.errors.blockFailed' : 'connections.errors.unblockFailed')
      notifyHaptic('error')
      return false
    } finally {
      busyAction.value = null
    }
  }

  function block(handle: string): Promise<boolean> {
    return mutate('block', handle, (key) => memberOperationsApi.dropConnection(handle, key))
  }

  function unblock(blockId: string): Promise<boolean> {
    return mutate('unblock', blockId, (key) => memberOperationsApi.unblockIP(blockId, key))
  }

  function resetOperation(nextAction: ConnectionBlockAction): void {
    mutationError.value = null
    action.value = nextAction
    operation.reset()
  }

  onMounted(() => void load())

  return {
    items: shallowReadonly(items), loading: readonly(loading), loadError: readonly(loadError),
    receipt: operation.receipt, polling: operation.polling, checking: operation.checking,
    terminal: operation.terminal, action: readonly(action), busyAction: readonly(busyAction),
    mutationActive, error, load, block, unblock, refreshOperation: operation.refresh, resetOperation,
  }
}
