import { computed, onMounted, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { Dashboard } from '@/api/types'
import { notifyHaptic } from '@/utils/telegram'

export function useDashboard() {
  const dashboard = shallowRef<Dashboard | null>(null)
  const loading = shallowRef(true)
  const refreshing = shallowRef(false)
  const revoking = shallowRef(false)
  const error = shallowRef<string | null>(null)

  const hasEntitlement = computed(() => dashboard.value?.activePurchase != null)
  const usageRatio = computed(() => {
    if (!dashboard.value?.statistics) return 0
    const used = Number(dashboard.value.statistics.usedTrafficBytes)
    const limit = Number(dashboard.value.statistics.trafficLimitBytes)
    if (!Number.isFinite(used) || !Number.isFinite(limit) || limit <= 0) return 0
    return Math.min(1, used / limit)
  })

  async function load(options: { quiet?: boolean } = {}): Promise<void> {
    if (options.quiet) refreshing.value = true
    else loading.value = true
    error.value = null
    try {
      dashboard.value = await api.getDashboard()
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'Dashboard data is unavailable.'
    } finally {
      loading.value = false
      refreshing.value = false
    }
  }

  async function revokeSubscription(): Promise<boolean> {
    revoking.value = true
    error.value = null
    try {
      const response = await api.revokeSubscription()
      if (dashboard.value) {
        dashboard.value = {
          ...dashboard.value,
          subscriptionUrl: response.subscriptionUrl,
        }
      }
      notifyHaptic('success')
      return true
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'The subscription link could not be revoked.'
      notifyHaptic('error')
      return false
    } finally {
      revoking.value = false
    }
  }

  onMounted(() => void load())

  return {
    dashboard: readonly(dashboard),
    loading: readonly(loading),
    refreshing: readonly(refreshing),
    revoking: readonly(revoking),
    error: readonly(error),
    hasEntitlement,
    usageRatio,
    load,
    revokeSubscription,
  }
}
