import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { CouponGrant } from '@/api/features'

const { discardCouponWalletGrant, getCouponWallet, notifyHaptic, redeemCoupon } = vi.hoisted(() => ({
  discardCouponWalletGrant: vi.fn(),
  getCouponWallet: vi.fn(),
  notifyHaptic: vi.fn(),
  redeemCoupon: vi.fn(),
}))

vi.mock('@/api/features', () => ({
  featuresApi: { discardCouponWalletGrant, getCouponWallet, redeemCoupon },
}))
vi.mock('@/i18n', () => ({
  localizedError: () => 'discard failed',
  t: (key: string, variables?: Record<string, string | number>) => variables?.name ? `${key}:${variables.name}` : key,
}))
vi.mock('@/utils/browserCompatibility', () => ({ createUuid: () => 'coupon-key' }))
vi.mock('@/utils/telegram', () => ({ notifyHaptic }))

import { useCoupons } from './useCoupons'

const grant: CouponGrant = {
  id: 'grant-1',
  sourceType: 'coupon',
  sourceId: 'coupon-1',
  status: 'active',
  useCount: 0,
  createdAt: '2026-08-11T00:00:00Z',
  consumedAt: null,
  coupon: {
    id: 'coupon-1',
    code: 'RIDE',
    name: 'Ride credit',
    kind: 'purchase_once',
    valueMinorOrBps: '100',
    eligibleComboIds: [],
    eligibleSquadIds: [],
    expiresAt: null,
    globalUseLimit: null,
    perUserUseLimit: 1,
    active: true,
    usageCount: 0,
    createdAt: '2026-08-11T00:00:00Z',
    updatedAt: '2026-08-11T00:00:00Z',
  },
}

describe('coupon wallet discard', () => {
  afterEach(() => vi.clearAllMocks())

  it('keeps a grant visible until the discard endpoint confirms it', async () => {
    let resolveDiscard!: () => void
    discardCouponWalletGrant.mockReturnValue(new Promise<void>((resolve) => { resolveDiscard = resolve }))
    getCouponWallet.mockResolvedValue({ items: [grant] })
    const scope = effectScope()
    const state = scope.run(() => useCoupons())!
    await state.load()

    const discard = state.discard(grant.id)
    expect(state.grants.value).toEqual([grant])
    resolveDiscard()

    await expect(discard).resolves.toBe(true)
    expect(state.grants.value).toEqual([])
    expect(discardCouponWalletGrant).toHaveBeenCalledWith(grant.id)
    scope.stop()
  })

  it('retains the grant and exposes an error when discard fails', async () => {
    getCouponWallet.mockResolvedValue({ items: [grant] })
    discardCouponWalletGrant.mockRejectedValue(new Error('offline'))
    const scope = effectScope()
    const state = scope.run(() => useCoupons())!
    await state.load()

    await expect(state.discard(grant.id)).resolves.toBe(false)
    expect(state.grants.value).toEqual([grant])
    expect(state.error.value).toBe('discard failed')
    expect(notifyHaptic).toHaveBeenCalledWith('error')
    scope.stop()
  })
})
