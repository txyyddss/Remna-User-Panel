import type { CouponGrant } from '@/api/features'

export function isCouponGrantAvailable(grant: CouponGrant, now = Date.now()): boolean {
  if (grant.status !== 'active' || !grant.coupon.active) return false
  if (grant.coupon.perUserUseLimit !== null && grant.useCount >= grant.coupon.perUserUseLimit) return false
  if (grant.coupon.globalUseLimit !== null && grant.coupon.usageCount >= grant.coupon.globalUseLimit) return false
  if (!grant.coupon.expiresAt) return true
  const expiresAt = Date.parse(grant.coupon.expiresAt)
  return Number.isFinite(expiresAt) && expiresAt > now
}
