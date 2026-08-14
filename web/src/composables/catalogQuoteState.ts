import { computed, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { PurchaseQuote } from '@/api/types'

interface QuoteSelection {
  comboId: string
  squadProductIds: string[]
  couponGrantId?: string
}

interface CatalogQuoteOptions {
  selection: () => QuoteSelection | null
  fingerprint: () => string
  onError: (caught: unknown) => void
}

export function useCatalogQuote(options: CatalogQuoteOptions) {
  const quote = shallowRef<PurchaseQuote | null>(null)
  const quoteFingerprint = shallowRef<string | null>(null)
  const quoting = shallowRef(false)
  let requestVersion = 0

  const quoteUsable = computed(() => quote.value !== null && quoteFingerprint.value === options.fingerprint())

  function clearQuote(): void {
    quote.value = null
    quoteFingerprint.value = null
  }

  async function refreshQuote(): Promise<boolean> {
    const selection = options.selection()
    const fingerprint = options.fingerprint()
    const version = ++requestVersion
    if (!selection) {
      clearQuote()
      quoting.value = false
      return false
    }
    quoting.value = true
    try {
      const response = await api.quotePurchase(selection.comboId, selection.squadProductIds, selection.couponGrantId)
      if (version !== requestVersion || options.fingerprint() !== fingerprint) return false
      quote.value = response
      quoteFingerprint.value = fingerprint
      return true
    } catch (caught) {
      if (version === requestVersion && options.fingerprint() === fingerprint) {
        clearQuote()
        options.onError(caught)
      }
      return false
    } finally {
      if (version === requestVersion) quoting.value = false
    }
  }

  return { quote: readonly(quote), quoting: readonly(quoting), quoteUsable, clearQuote, refreshQuote }
}
