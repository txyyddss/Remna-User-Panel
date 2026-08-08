import { computed, getCurrentInstance, onMounted, readonly, shallowRef, watch } from 'vue'

import { api, ApiError } from '@/api/client'
import { featuresApi } from '@/api/features'
import type { CouponGrant } from '@/api/features'
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
  const couponGrants = shallowRef<CouponGrant[]>([])
  const selectedCouponGrantId = shallowRef<string | null>(null)
  const checkoutOpen = shallowRef(false)
  const needsBalance = shallowRef(false)
  const purchaseIdempotencyKey = shallowRef<string | null>(null)

  const visibleCombos = computed(() => catalog.value?.combos.filter((combo) => combo.active) ?? [])
  const selectedCombo = computed<Combo | undefined>(() =>
    visibleCombos.value.find((combo) => combo.id === selectedComboId.value),
  )
  const includedSquadIds = computed(() => selectedCombo.value?.includedSquads.map((squad) => squad.id) ?? [])
  const visibleSquads = computed(() => {
    const squads = [
      ...(selectedCombo.value?.includedSquads ?? []),
      ...(catalog.value?.addons.filter((squad) => squad.visible && squad.upstreamPresent) ?? []),
    ]
    return squads.filter((squad, index) => squads.findIndex((candidate) => candidate.id === squad.id) === index)
  })
  const selectedSquads = computed<SquadProduct[]>(() =>
    visibleSquads.value.filter((squad) => selectedSquadIds.value.includes(squad.id)),
  )
  const eligibleCoupons = computed(() => couponGrants.value.filter((grant) => {
    const coupon = grant.coupon
    if (!selectedCombo.value || grant.status !== 'active' || !coupon.active) return false
    if (coupon.perUserUseLimit !== null && grant.useCount >= coupon.perUserUseLimit) return false
    if (coupon.expiresAt && new Date(coupon.expiresAt).getTime() <= Date.now()) return false
    const unrestricted = coupon.eligibleComboIds.length === 0 && coupon.eligibleSquadIds.length === 0
    const comboEligible = coupon.eligibleComboIds.includes(selectedCombo.value.id)
    const squadEligible = selectedSquadIds.value.some((id) => coupon.eligibleSquadIds.includes(id))
    if (!unrestricted && !comboEligible && !squadEligible) return false
    return coupon.kind === 'purchase_once' || coupon.kind === 'purchase_recurring'
  }))
  const purchaseFingerprint = computed(() => JSON.stringify([
    selectedComboId.value,
    [...selectedSquadIds.value].sort(),
    selectedCouponGrantId.value,
  ]))

  watch(purchaseFingerprint, () => {
    purchaseIdempotencyKey.value = null
  })

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const [catalogResponse, balanceResponse, couponResponse] = await Promise.all([
        api.getCatalog(),
        api.getBalance(),
        featuresApi.getCouponWallet().catch(() => ({ items: [] })),
      ])
      catalog.value = catalogResponse
      balance.value = balanceResponse.balance
      couponGrants.value = couponResponse.items
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
    const included = visibleCombos.value.find((combo) => combo.id === id)?.includedSquads.map((squad) => squad.id) ?? []
    selectedSquadIds.value = selectedSquadIds.value.filter((squadId) => !included.includes(squadId))
    if (!eligibleCoupons.value.some((coupon) => coupon.id === selectedCouponGrantId.value)) selectedCouponGrantId.value = null
    purchase.value = null
  }

  function toggleSquad(id: string): void {
    if (includedSquadIds.value.includes(id)) return
    selectedSquadIds.value = selectedSquadIds.value.includes(id)
      ? selectedSquadIds.value.filter((value) => value !== id)
      : [...selectedSquadIds.value, id]
    if (!eligibleCoupons.value.some((coupon) => coupon.id === selectedCouponGrantId.value)) selectedCouponGrantId.value = null
  }

  function reviewPurchase(): void {
    if (!selectedCombo.value) return
    error.value = null
    needsBalance.value = false
    purchaseIdempotencyKey.value ??= globalThis.crypto.randomUUID()
    checkoutOpen.value = true
  }

  async function confirmPurchase(): Promise<boolean> {
    if (!selectedCombo.value) return false
    purchasing.value = true
    error.value = null
    needsBalance.value = false
    try {
      purchaseIdempotencyKey.value ??= globalThis.crypto.randomUUID()
      purchase.value = await api.createPurchase(
        selectedCombo.value.id,
        selectedSquadIds.value,
        selectedCouponGrantId.value ?? undefined,
        purchaseIdempotencyKey.value,
      )
      notifyHaptic('success')
      checkoutOpen.value = false
      purchaseIdempotencyKey.value = null
      selectedSquadIds.value = []
      selectedCouponGrantId.value = null
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

  if (getCurrentInstance()) onMounted(() => void load())

  return {
    catalog: readonly(catalog),
    balance: readonly(balance),
    loading: readonly(loading),
    purchasing: readonly(purchasing),
    error: readonly(error),
    purchase: readonly(purchase),
    selectedComboId,
    selectedSquadIds: readonly(selectedSquadIds),
    selectedCouponGrantId,
    checkoutOpen,
    needsBalance: readonly(needsBalance),
    visibleCombos,
    visibleSquads,
    selectedCombo,
    selectedSquads,
    includedSquadIds,
    eligibleCoupons,
    load,
    selectCombo,
    toggleSquad,
    reviewPurchase,
    confirmPurchase,
  }
}
