import { computed, onMounted, readonly, shallowRef } from 'vue'

import { api, ApiError } from '@/api/client'
import type { Catalog, Combo, Purchase, SquadProduct } from '@/api/types'
import { notifyHaptic } from '@/utils/telegram'

export function useCatalog() {
  const catalog = shallowRef<Catalog | null>(null)
  const balance = shallowRef<import('@/api/types').Money | null>(null)
  const loading = shallowRef(true)
  const purchasing = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const purchase = shallowRef<Purchase | null>(null)
  const selectedComboId = shallowRef<string | null>(null)
  const selectedSquadIds = shallowRef<string[]>([])
  const checkoutOpen = shallowRef(false)
  const needsBalance = shallowRef(false)

  const visibleCombos = computed(() => catalog.value?.combos.filter((combo) => combo.active) ?? [])
  const visibleSquads = computed(() => catalog.value?.addons.filter((squad) => squad.visible && squad.upstreamPresent) ?? [])
  const selectedCombo = computed<Combo | undefined>(() =>
    visibleCombos.value.find((combo) => combo.id === selectedComboId.value),
  )
  const selectedSquads = computed<SquadProduct[]>(() =>
    visibleSquads.value.filter((squad) => selectedSquadIds.value.includes(squad.id)),
  )

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const [catalogResponse, balanceResponse] = await Promise.all([api.getCatalog(), api.getBalance()])
      catalog.value = catalogResponse
      balance.value = balanceResponse.balance
      const preferred = catalog.value.combos.find((combo) => combo.active)
      selectedComboId.value = preferred?.id ?? null
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'The catalog is unavailable.'
    } finally {
      loading.value = false
    }
  }

  function selectCombo(id: string): void {
    selectedComboId.value = id
    purchase.value = null
  }

  function toggleSquad(id: string): void {
    selectedSquadIds.value = selectedSquadIds.value.includes(id)
      ? selectedSquadIds.value.filter((value) => value !== id)
      : [...selectedSquadIds.value, id]
  }

  function reviewPurchase(): void {
    if (!selectedCombo.value) return
    error.value = null
    needsBalance.value = false
    checkoutOpen.value = true
  }

  async function confirmPurchase(): Promise<boolean> {
    if (!selectedCombo.value) return false
    purchasing.value = true
    error.value = null
    needsBalance.value = false
    try {
      purchase.value = await api.createPurchase(selectedCombo.value.id, selectedSquadIds.value)
      notifyHaptic('success')
      checkoutOpen.value = false
      await load()
      return true
    } catch (caught) {
      if (caught instanceof ApiError && caught.code === 'INSUFFICIENT_BALANCE') {
        needsBalance.value = true
        error.value = 'Your TXB balance is too low for this purchase.'
      } else {
        error.value = caught instanceof Error ? caught.message : 'The purchase could not be completed.'
      }
      notifyHaptic('error')
      return false
    } finally {
      purchasing.value = false
    }
  }

  onMounted(() => void load())

  return {
    catalog: readonly(catalog),
    balance: readonly(balance),
    loading: readonly(loading),
    purchasing: readonly(purchasing),
    error: readonly(error),
    purchase: readonly(purchase),
    selectedComboId,
    selectedSquadIds: readonly(selectedSquadIds),
    checkoutOpen,
    needsBalance: readonly(needsBalance),
    visibleCombos,
    visibleSquads,
    selectedCombo,
    selectedSquads,
    load,
    selectCombo,
    toggleSquad,
    reviewPurchase,
    confirmPurchase,
  }
}
