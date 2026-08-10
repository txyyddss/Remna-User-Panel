import { onMounted, readonly, shallowRef } from 'vue'

import type { CouponGrant } from '@/api/features'
import { featuresApi } from '@/api/features'
import { localizedError, t } from '@/i18n'
import { notifyHaptic } from '@/utils/telegram'

export function useCoupons() {
  const grants = shallowRef<CouponGrant[]>([])
  const loading = shallowRef(true)
  const redeeming = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const message = shallowRef<string | null>(null)
  const redemptionKeys = new Map<string, string>()

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try { grants.value = (await featuresApi.getCouponWallet()).items } catch (caught) { error.value = localizedError(caught, 'errors.couponWallet') } finally { loading.value = false }
  }

  async function redeem(code: string): Promise<boolean> {
    if (redeeming.value || !code.trim()) return false
    redeeming.value = true
    error.value = null
    message.value = null
    try {
      const canonicalCode = code.trim().toUpperCase()
      const key = redemptionKeys.get(canonicalCode) ?? globalThis.crypto.randomUUID()
      redemptionKeys.set(canonicalCode, key)
      const result = await featuresApi.redeemCoupon(canonicalCode, key)
      redemptionKeys.delete(canonicalCode)
      if (result.grant) grants.value = [result.grant, ...grants.value.filter((item) => item.id !== result.grant?.id)]
      message.value = result.grant
        ? t('coupons.addedToWallet', { name: result.coupon.name })
        : t('coupons.appliedToBalance', { name: result.coupon.name })
      notifyHaptic('success')
      return true
    } catch (caught) {
      error.value = localizedError(caught, 'errors.couponRedeem')
      notifyHaptic('error')
      return false
    } finally {
      redeeming.value = false
    }
  }

  onMounted(() => void load())

  return { grants: readonly(grants), loading: readonly(loading), redeeming: readonly(redeeming), error: readonly(error), message: readonly(message), load, redeem }
}
