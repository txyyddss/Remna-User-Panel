import { computed, readonly, shallowRef } from 'vue'

import { ApiError, api } from '@/api/client'
import type { Catalog, Purchase, SquadProduct } from '@/api/types'
import { localizedError, t } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { notifyHaptic } from '@/utils/telegram'

interface SquadAdditionOptions {
  purchaseId: () => string
  activeSquadUuids: () => readonly string[]
}

export function useSquadAddition(options: SquadAdditionOptions) {
  const catalog = shallowRef<Catalog | null>(null)
  const selectedSquadIds = shallowRef<string[]>([])
  const quote = shallowRef<import('@/api/types').PurchaseAddonQuote | null>(null)
  const quoteFingerprint = shallowRef<string | null>(null)
  const purchase = shallowRef<Purchase | null>(null)
  const loading = shallowRef(false)
  const quoting = shallowRef(false)
  const purchasing = shallowRef(false)
  const needsBalance = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const activeSquadSet = computed(() => new Set(options.activeSquadUuids()))
  const visibleSquads = computed(() => (catalog.value?.addons ?? []).filter((squad) =>
    squad.visible && squad.upstreamPresent && !activeSquadSet.value.has(squad.remnaSquadUuid),
  ))
  const selectedSquads = computed(() => visibleSquads.value.filter((squad) => selectedSquadIds.value.includes(squad.id)))
  const activationSquads = computed(() => selectedSquads.value.filter((squad) => squad.activationRequired))
  const fingerprint = computed(() => JSON.stringify([...selectedSquadIds.value].sort()))
  const quoteUsable = computed(() => quote.value !== null && quoteFingerprint.value === fingerprint.value)

  function clearQuote(): void {
    quote.value = null
    quoteFingerprint.value = null
  }

  function reset(): void {
    selectedSquadIds.value = []
    purchase.value = null
    error.value = null
    needsBalance.value = false
    clearQuote()
  }

  function isUnavailable(squad: SquadProduct): boolean {
    return squad.stockRemaining === 0 && !squad.stockHeldByCurrentUser
  }

  function toggleSquad(id: string): void {
    const squad = visibleSquads.value.find((candidate) => candidate.id === id)
    if (!squad || isUnavailable(squad)) return
    selectedSquadIds.value = selectedSquadIds.value.includes(id)
      ? selectedSquadIds.value.filter((value) => value !== id)
      : [...selectedSquadIds.value, id]
    clearQuote()
    error.value = null
  }

  async function load(): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      catalog.value = await api.getCatalog()
      selectedSquadIds.value = selectedSquadIds.value.filter((id) => visibleSquads.value.some((squad) => squad.id === id))
      return true
    } catch (caught) {
      error.value = localizedError(caught, 'errors.catalogUnavailable')
      notifyHaptic('error')
      return false
    } finally {
      loading.value = false
    }
  }

  async function refreshQuote(): Promise<boolean> {
    if (!selectedSquadIds.value.length) return false
    const selectedFingerprint = fingerprint.value
    quoting.value = true
    error.value = null
    needsBalance.value = false
    try {
      const response = await api.quotePurchaseAddons(options.purchaseId(), selectedSquadIds.value)
      if (selectedFingerprint !== fingerprint.value) return false
      quote.value = response
      quoteFingerprint.value = selectedFingerprint
      return true
    } catch (caught) {
      if (selectedFingerprint === fingerprint.value) error.value = localizedError(caught, 'errors.quoteFailed')
      notifyHaptic('error')
      return false
    } finally {
      quoting.value = false
    }
  }

  async function confirmPurchase(squadActivationCodes: Record<string, string>): Promise<boolean> {
    if (purchasing.value || !selectedSquadIds.value.length) return false
    if (!quoteUsable.value && !(await refreshQuote())) return false
    purchasing.value = true
    error.value = null
    needsBalance.value = false
    try {
      purchase.value = await api.addPurchaseAddons(options.purchaseId(), selectedSquadIds.value, createUuid(), squadActivationCodes)
      notifyHaptic('success')
      return true
    } catch (caught) {
      if (caught instanceof ApiError && caught.code === 'INSUFFICIENT_BALANCE') {
        needsBalance.value = true
        error.value = t('errors.purchaseBalance')
      } else {
        error.value = localizedError(caught, 'errors.purchaseFailed')
      }
      notifyHaptic('error')
      return false
    } finally {
      purchasing.value = false
    }
  }

  return {
    catalog: readonly(catalog), selectedSquadIds: readonly(selectedSquadIds), quote: readonly(quote), purchase: readonly(purchase),
    loading: readonly(loading), quoting: readonly(quoting), purchasing: readonly(purchasing), needsBalance: readonly(needsBalance), error: readonly(error),
    visibleSquads, selectedSquads, activationSquads, quoteUsable, reset, toggleSquad, load, refreshQuote, confirmPurchase,
  }
}
