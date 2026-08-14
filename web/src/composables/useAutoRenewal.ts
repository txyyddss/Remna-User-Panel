import { computed, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { AutoRenewal } from '@/api/types'
import { localizedError } from '@/i18n'
import { notifyHaptic } from '@/utils/telegram'

export function useAutoRenewal(purchaseID: () => string, initialEnabled: () => boolean) {
  const renewal = shallowRef<AutoRenewal | null>(null)
  const loading = shallowRef(false)
  const updating = shallowRef(false)
  const error = shallowRef<string | null>(null)
  let requestVersion = 0

  const enabled = computed(() => renewal.value?.enabled ?? initialEnabled())
  const canEnable = computed(() => renewal.value?.canEnable ?? false)
  const ineligibleReason = computed(() => renewal.value?.ineligibleReason ?? null)

  function reset(): void {
    requestVersion += 1
    renewal.value = null
    error.value = null
  }

  async function load(): Promise<void> {
    const purchaseId = purchaseID()
    const version = ++requestVersion
    loading.value = true
    error.value = null
    try {
      const response = await api.getAutoRenewal(purchaseId)
      if (version === requestVersion && purchaseID() === purchaseId) renewal.value = response
    } catch (caught) {
      if (version === requestVersion) error.value = localizedError(caught, 'errors.autoRenewalFailed')
    } finally {
      if (version === requestVersion) loading.value = false
    }
  }

  async function setEnabled(next: boolean): Promise<boolean> {
    if (updating.value || (next && !canEnable.value)) return false
    updating.value = true
    error.value = null
    try {
      renewal.value = await api.setAutoRenewal(purchaseID(), next)
      notifyHaptic('success')
      return true
    } catch (caught) {
      error.value = localizedError(caught, 'errors.autoRenewalFailed')
      notifyHaptic('error')
      return false
    } finally {
      updating.value = false
    }
  }

  return {
    renewal: readonly(renewal),
    loading: readonly(loading),
    updating: readonly(updating),
    error: readonly(error),
    enabled,
    canEnable,
    ineligibleReason,
    reset,
    load,
    setEnabled,
  }
}
