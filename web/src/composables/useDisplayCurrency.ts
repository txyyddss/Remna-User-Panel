import { computed, readonly, shallowRef } from 'vue'

import { displayCurrencyApi } from '@/api/displayCurrency'
import type { DisplayCurrencyPreference } from '@/api/types'
import { t } from '@/i18n'
import { displayCurrencyState, resetDisplayCurrency, setDisplayCurrency } from '@/utils/displayCurrency'

const saving = shallowRef(false)
const loading = shallowRef(false)
const error = shallowRef<string | null>(null)

export function useDisplayCurrency() {
  const currencies = computed<DisplayCurrencyPreference['currency'][]>(() => ['TXB', 'CNY', 'USD'])

  async function refresh(): Promise<void> {
    if (loading.value) return
    loading.value = true
    try {
      setDisplayCurrency(await displayCurrencyApi.get())
      error.value = null
    } catch {
      if (displayCurrencyState.currency.value !== 'TXB') error.value = t('app.currencyRateUnavailable')
    } finally {
      loading.value = false
    }
  }

  async function select(next: DisplayCurrencyPreference['currency']): Promise<boolean> {
    if (saving.value || next === displayCurrencyState.currency.value) return false
    saving.value = true
    try {
      setDisplayCurrency(await displayCurrencyApi.update(next))
      error.value = null
      return true
    } catch {
      error.value = t('app.currencyRateUnavailable')
      return false
    } finally {
      saving.value = false
    }
  }

  function hydrate(currency: DisplayCurrencyPreference['currency'] | undefined): void {
    resetDisplayCurrency(currency ?? 'TXB')
    error.value = null
  }

  return {
    currency: displayCurrencyState.currency,
    rates: displayCurrencyState.rates,
    currencies,
    saving: readonly(saving),
    loading: readonly(loading),
    error: readonly(error),
    refresh,
    select,
    hydrate,
  }
}
