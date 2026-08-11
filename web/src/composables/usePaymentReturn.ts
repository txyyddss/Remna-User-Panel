import type { Ref } from 'vue'
import { computed, onScopeDispose, readonly, shallowRef, watch } from 'vue'

import type { FeaturePaymentOrder } from '@/api/features'
import { api } from '@/api/client'

export type PaymentReturnState = 'checking' | 'pending' | 'confirmed' | 'terminal' | 'missing' | 'unavailable'

const terminalStatuses = new Set<FeaturePaymentOrder['status']>(['cancelled', 'expired', 'failed', 'refunded'])

export function usePaymentReturn(orderId: Ref<string>) {
  const state = shallowRef<PaymentReturnState>('checking')
  const order = shallowRef<FeaturePaymentOrder | null>(null)
  let pollTimer: ReturnType<typeof setTimeout> | undefined

  const isConfirmed = computed(() => state.value === 'confirmed')

  function stopPolling(): void {
    if (pollTimer !== undefined) clearTimeout(pollTimer)
    pollTimer = undefined
  }

  function schedulePoll(): void {
    stopPolling()
    pollTimer = setTimeout(() => void refresh(), 2000)
  }

  async function refresh(): Promise<void> {
    stopPolling()
    if (!orderId.value) {
      state.value = 'missing'
      return
    }
    state.value = order.value ? 'pending' : 'checking'
    try {
      const next = await api.getPaymentOrder(orderId.value)
      order.value = next
      if (next.status === 'paid') {
        state.value = 'confirmed'
      } else if (terminalStatuses.has(next.status)) {
        state.value = 'terminal'
      } else {
        state.value = 'pending'
        schedulePoll()
      }
    } catch {
      if (order.value) {
        state.value = 'pending'
        schedulePoll()
      } else {
        state.value = 'unavailable'
      }
    }
  }

  watch(orderId, () => {
    order.value = null
    void refresh()
  }, { immediate: true })
  onScopeDispose(stopPolling)

  return { state: readonly(state), order: readonly(order), isConfirmed, refresh, stopPolling }
}
