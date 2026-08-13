import { shallowRef } from 'vue'

import { api } from '@/api/client'
import type { RolloverProjection } from '@/api/types'
import { localizedError } from '@/i18n'

export function useRolloverDetail() {
  const detail = shallowRef<RolloverProjection | null>(null)
  const loading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const purchaseId = shallowRef<string | null>(null)
  let requestVersion = 0

  async function load(id: string): Promise<void> {
    const version = ++requestVersion
    purchaseId.value = id
    loading.value = true
    error.value = null
    try {
      const response = await api.getPurchaseRollover(id)
      if (version === requestVersion) detail.value = response
    } catch (caught) {
      if (version === requestVersion) {
        detail.value = null
        error.value = localizedError(caught, 'errors.rolloverFailed')
      }
    } finally {
      if (version === requestVersion) loading.value = false
    }
  }

  function reset(): void {
    requestVersion += 1
    detail.value = null
    loading.value = false
    error.value = null
    purchaseId.value = null
  }

  async function retry(): Promise<void> {
    if (purchaseId.value) await load(purchaseId.value)
  }

  return { detail, loading, error, load, retry, reset }
}
