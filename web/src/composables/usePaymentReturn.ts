import type { Ref } from 'vue'
import { computed, onScopeDispose, readonly, shallowRef, watch } from 'vue'

import type { FeaturePaymentOrder, PaymentReturnProvider } from '@/api/features'
import { api } from '@/api/client'

export type PaymentReturnState = 'checking' | 'pending' | 'confirmed' | 'terminal' | 'missing' | 'unavailable'

const terminalStatuses = new Set<FeaturePaymentOrder['status']>(['cancelled', 'expired', 'failed', 'refunded'])

interface PaymentReturnOptions {
  browserStatus?: boolean
  provider?: Ref<PaymentReturnProvider | null>
  capability?: Ref<string>
}

export function usePaymentReturn(orderId: Ref<string>, options: PaymentReturnOptions = {}) {
  const state = shallowRef<PaymentReturnState>('checking')
  const order = shallowRef<FeaturePaymentOrder | null>(null)
  const orderStatus = shallowRef<FeaturePaymentOrder['status'] | null>(null)
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

  function applyStatus(status: FeaturePaymentOrder['status']): void {
    if (status === 'paid') {
      state.value = 'confirmed'
    } else if (terminalStatuses.has(status)) {
      state.value = 'terminal'
    } else {
      state.value = 'pending'
      schedulePoll()
    }
  }

  async function refresh(): Promise<void> {
    stopPolling()
    if (!orderId.value) {
      state.value = 'missing'
      return
    }
    state.value = order.value ? 'pending' : 'checking'
    try {
      if (options.browserStatus) {
        const provider = options.provider?.value
        const capability = options.capability?.value
        if (!provider || !capability) {
          state.value = 'unavailable'
          return
        }
        const next = await api.getPaymentReturnStatus(provider, orderId.value, capability)
        orderStatus.value = next.status
        applyStatus(next.status)
      } else {
        const next = await api.getPaymentOrder(orderId.value)
        order.value = next
        orderStatus.value = next.status
        applyStatus(next.status)
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

  watch(() => [orderId.value, options.provider?.value ?? null, options.capability?.value ?? ''], () => {
    order.value = null
    orderStatus.value = null
    void refresh()
  }, { immediate: true })
  onScopeDispose(stopPolling)

  return { state: readonly(state), order: readonly(order), orderStatus: readonly(orderStatus), isConfirmed, refresh, stopPolling }
}
