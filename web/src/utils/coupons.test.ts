import { describe, expect, it } from 'vitest'

import type { CouponGrant } from '@/api/features'
import { isCouponGrantAvailable } from './coupons'

function grant(overrides: Partial<CouponGrant> = {}): CouponGrant {
  return {
    id: 'grant-1', sourceType: 'admin', sourceId: 'source-1', status: 'active', useCount: 0,
    createdAt: '2026-08-18T00:00:00Z', consumedAt: null,
    coupon: {
      id: 'coupon-1', code: 'RIDE', name: 'Ride', kind: 'purchase_once', discountMode: 'fixed',
      valueMinorOrBps: '100', eligibleComboIds: [], eligibleSquadIds: [], expiresAt: null,
      globalUseLimit: null, perUserUseLimit: null, active: true, usageCount: 0,
      createdAt: '2026-08-18T00:00:00Z', updatedAt: '2026-08-18T00:00:00Z',
    },
    ...overrides,
  }
}

describe('isCouponGrantAvailable', () => {
  it('hides inactive, expired, and exhausted grants', () => {
    const now = Date.parse('2026-08-18T12:00:00Z')
    expect(isCouponGrantAvailable(grant(), now)).toBe(true)
    expect(isCouponGrantAvailable(grant({ status: 'expired' }), now)).toBe(false)
    expect(isCouponGrantAvailable(grant({ useCount: 1, coupon: { ...grant().coupon, perUserUseLimit: 1 } }), now)).toBe(false)
    expect(isCouponGrantAvailable(grant({ coupon: { ...grant().coupon, globalUseLimit: 4, usageCount: 4 } }), now)).toBe(false)
    expect(isCouponGrantAvailable(grant({ coupon: { ...grant().coupon, expiresAt: '2026-08-18T11:59:59Z' } }), now)).toBe(false)
  })
})
