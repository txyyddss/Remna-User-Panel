import type { ShallowRef } from 'vue'
import { api } from '@/api/client'
import { featuresApi, type CouponGrant } from '@/api/features'
import type { Catalog, Money } from '@/api/types'
import { localizedError } from '@/i18n'
import { readCatalogDraft } from '@/composables/catalogDraft'

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
}

export async function loadCatalogData(state: CatalogLoadState): Promise<void> {
  state.loading.value = true
  state.error.value = null
  try {
    const [catalogResponse, balanceResponse, couponResponse] = await Promise.all([
      api.getCatalog(), api.getBalance(), featuresApi.getCouponWallet().catch(() => ({ items: [] })),
    ])
    state.catalog.value = catalogResponse
    state.balance.value = balanceResponse.balance
    state.couponGrants.value = couponResponse.items
    if (!state.draftRestored.value) {
      state.draftRestored.value = true
      const draft = readCatalogDraft(state.userID())
      if (draft) {
        const validCombo = catalogResponse.combos.find((combo) => combo.active && combo.id === draft.comboId)
        if (validCombo) state.selectedComboId.value = validCombo.id
        const validSquads = new Set(catalogResponse.addons.filter((squad) => squad.visible && squad.upstreamPresent).map((squad) => squad.id))
        state.selectedSquadIds.value = (draft.squadIds ?? []).filter((id) => validSquads.has(id))
        state.selectedCouponGrantId.value = couponResponse.items.some((grant) => grant.id === draft.couponGrantId && grant.status === 'active') ? (draft.couponGrantId ?? null) : null
      }
    }
    const preferred = catalogResponse.combos.find((combo) => combo.active)
    if (!state.selectedComboId.value || !catalogResponse.combos.some((combo) => combo.active && combo.id === state.selectedComboId.value)) state.selectedComboId.value = preferred?.id ?? null
  } catch (caught) {
    state.error.value = localizedError(caught, 'errors.catalogUnavailable')
  } finally {
    state.loading.value = false
  }
}
