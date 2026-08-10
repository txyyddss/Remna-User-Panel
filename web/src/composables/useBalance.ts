import { onMounted, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { FeaturePaymentMethod } from '@/api/features'
import type { LedgerEntry, Money } from '@/api/types'
import { localizedError } from '@/i18n'

export function useBalance() {
  const balance = shallowRef<Money | null>(null)
  const methods = shallowRef<FeaturePaymentMethod[]>([])
  const ledger = shallowRef<LedgerEntry[]>([])
  const loading = shallowRef(true)
  const error = shallowRef<string | null>(null)

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const [balanceResponse, ledgerResponse] = await Promise.all([
        api.getBalance(),
        api.getLedger(),
      ])
      balance.value = balanceResponse.balance
      methods.value = balanceResponse.paymentMethods
      ledger.value = ledgerResponse.items
    } catch (caught) {
      error.value = localizedError(caught, 'errors.balanceUnavailable')
    } finally {
      loading.value = false
    }
  }

  onMounted(() => void load())

  return {
    balance: readonly(balance),
    methods: readonly(methods),
    ledger: readonly(ledger),
    loading: readonly(loading),
    error: readonly(error),
    load,
  }
}
