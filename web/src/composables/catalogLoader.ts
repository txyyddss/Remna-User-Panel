import type { ShallowRef } from 'vue'
import { api } from '@/api/client'
import { restoreCached, restoreItems, restoreRef } from '@/api/cache/restore'
import { featuresApi, type CouponGrant } from '@/api/features'
import type { Catalog, Money } from '@/api/types'
import { localizedError } from '@/i18n'
import { readCatalogDraft } from '@/composables/catalogDraft'
import type { LatestRequest } from '@/utils/latestRequest'

export interface CatalogLoadState {
  loading: ShallowRef<boolean>
  error: ShallowRef<string | null>
  catalog: ShallowRef<Catalog | null>
  balance: ShallowRef<Money | null>
  couponGrants: ShallowRef<CouponGrant[]>
  draftRestored: ShallowRef<boolean>
  selectedComboId: ShallowRef<string | null>
  selectedSquadIds: ShallowRef<string[]>
  selectedCouponGrantId: ShallowRef<string | null>
  userID: () => string | null | undefined
  onError?: (caught: unknown) => boolean
  latest: LatestRequest
}

function usableCatalog(value: Catalog): boolean {
  return Array.isArray(value?.combos) && Array.isArray(value?.addons)
}

function normalizedCatalog(value: Catalog): Catalog {
  return {
    ...value,
    combos: Array.isArray(value.combos) ? value.combos : [],
    addons: Array.isArray(value.addons) ? value.addons : [],
    nodes: Array.isArray(value.nodes) ? value.nodes : [],
  }
}

export async function loadCatalogData(state: CatalogLoadState): Promise<boolean> {
  const token = state.latest.begin()

  state.loading.value = !restoreRef('/api/v1/catalog', state.catalog)
  restoreCached<Awaited<ReturnType<typeof api.getBalance>>>(
    '/api/v1/balance',
    (value) => { state.balance.value = value.balance },
  )
  restoreItems('/api/v1/coupons/wallet', state.couponGrants)

  if (state.catalog.value) {
    state.catalog.value = normalizedCatalog(state.catalog.value)
    restoreSelection(state, state.catalog.value, state.couponGrants.value)
  }

  state.error.value = null

  const settle = async <T>(promise: Promise<T>): Promise<{ ok: true; value: T } | { ok: false; reason: unknown }> => {
    try {
      return { ok: true, value: await promise }
    } catch (reason) {
      return { ok: false, reason }
    }
  }

  const [catalogResult, balanceResult, couponResult] = await Promise.all([
    settle(api.getCatalog()),
    settle(api.getBalance()),
    settle(featuresApi.getCouponWallet()),
  ])

  if (!state.latest.isCurrent(token)) return false

  try {
    if (!catalogResult.ok) throw catalogResult.reason
    if (!usableCatalog(catalogResult.value)) throw new TypeError('CATALOG_RESPONSE_INVALID')

    state.catalog.value = normalizedCatalog(catalogResult.value)

    // Balance is not a prerequisite for browsing/selecting a combo. Checkout
    // is server-priced and already reports INSUFFICIENT_BALANCE on purchase.
    // Keep a restored balance when refresh fails, otherwise update it.
    if (balanceResult.ok) {
      state.balance.value = balanceResult.value.balance
    }

    // Coupon-wallet failures were already non-fatal before this change.
    state.couponGrants.value = couponResult.ok ? couponResult.value.items : []

    restoreSelection(state, state.catalog.value, state.couponGrants.value)
    return true
  } catch (caught) {
    if (!state.onError?.(caught)) {
      state.error.value = localizedError(caught, 'errors.catalogUnavailable')
    }
    return false
  } finally {
    if (state.latest.isCurrent(token)) state.loading.value = false
  }
}

function restoreSelection(state: CatalogLoadState, catalogResponse: Catalog, coupons: CouponGrant[]): void {
  if (!state.draftRestored.value) {
    state.draftRestored.value = true
    const draft = readCatalogDraft(state.userID())

    if (draft) {
      const validCombo = catalogResponse.combos.find((combo) => combo.active && combo.id === draft.comboId)
      if (validCombo) state.selectedComboId.value = validCombo.id

      const validSquads = new Set(
        catalogResponse.addons
          .filter((squad) => squad.visible && squad.upstreamPresent)
          .map((squad) => squad.id),
      )

      state.selectedSquadIds.value = (draft.squadIds ?? []).filter((id) => validSquads.has(id))
      state.selectedCouponGrantId.value = coupons.some(
        (grant) => grant.id === draft.couponGrantId && grant.status === 'active',
      )
        ? (draft.couponGrantId ?? null)
        : null
    }
  }

  const preferred = catalogResponse.combos.find((combo) => combo.active)
  if (
    !state.selectedComboId.value
    || !catalogResponse.combos.some(
      (combo) => combo.active && combo.id === state.selectedComboId.value,
    )
  ) {
    state.selectedComboId.value = preferred?.id ?? null
  }
}
