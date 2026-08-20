import { onMounted, readonly, shallowRef } from 'vue'

import type { AdminAffiliateView, AffiliateTier, CouponDefinition } from '@/api/features'
import { featuresApi } from '@/api/features'
import { localizedError } from '@/i18n'

export function useAdminAffiliates() {
  const configuration = shallowRef<AdminAffiliateView | null>(null)
  const coupons = shallowRef<CouponDefinition[]>([])
  const tiers = shallowRef<AffiliateTier[]>([])
  const loading = shallowRef(true)
  const saving = shallowRef(false)
  const error = shallowRef<string | null>(null)

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const [config, couponPage] = await Promise.all([featuresApi.getAdminAffiliates(), featuresApi.getAdminCoupons()])
      configuration.value = config
      tiers.value = structuredClone(config.tiers)
      coupons.value = couponPage.items.filter((coupon) => coupon.active && coupon.kind.startsWith('purchase_'))
    } catch (caught) { error.value = localizedError(caught, 'adminAffiliates.loadFailed') }
    finally { loading.value = false }
  }

  async function save(): Promise<void> {
    if (!configuration.value || saving.value) return
    saving.value = true
    error.value = null
    try {
      const saved = await featuresApi.saveAdminAffiliates(configuration.value.version, tiers.value)
      configuration.value = saved
      tiers.value = structuredClone(saved.tiers)
    } catch (caught) { error.value = localizedError(caught, 'adminAffiliates.saveFailed') }
    finally { saving.value = false }
  }

  function replace(next: AffiliateTier[]): void { tiers.value = next }
  onMounted(() => void load())
  return { configuration: readonly(configuration), coupons: readonly(coupons), tiers: readonly(tiers), loading: readonly(loading), saving: readonly(saving), error: readonly(error), load, save, replace }
}
