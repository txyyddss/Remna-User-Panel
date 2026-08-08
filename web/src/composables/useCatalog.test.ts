import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Catalog, SquadProduct } from '@/api/types'

const { getCatalog, getBalance, getCouponWallet, createPurchase } = vi.hoisted(() => ({
  getCatalog: vi.fn(),
  getBalance: vi.fn(),
  getCouponWallet: vi.fn(),
  createPurchase: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  api: { getCatalog, getBalance, createPurchase },
  ApiError: class ApiError extends Error {},
}))

vi.mock('@/api/features', () => ({
  featuresApi: { getCouponWallet },
}))

import { useCatalog } from './useCatalog'

const money = (minor: string) => ({ currency: 'TXB' as const, minor, display: `${Number(minor) / 100} TXB` })
const squad = (id: string): SquadProduct => ({
  id,
  remnaSquadUuid: `00000000-0000-4000-8000-${id.padStart(12, '0')}`,
  name: `Squad ${id}`,
  description: '',
  price: money('100'),
  visible: true,
  upstreamPresent: true,
  createdAt: '2026-08-08T00:00:00Z',
  updatedAt: '2026-08-08T00:00:00Z',
})

describe('catalog squad selection', () => {
  afterEach(() => vi.clearAllMocks())

  it('disables included squads and prunes newly included paid add-ons after a combo change', async () => {
    const alpha = squad('1')
    const beta = squad('2')
    const catalog = {
      addons: [alpha, beta],
      combos: [
        { id: 'combo-a', name: 'A', description: '', price: money('1000'), validityDays: 30, trafficLimitBytes: '1', resetStrategy: 'MONTH', includedSquads: [alpha], active: true, rolloverMinRemainingBps: 0, rolloverMaxTxbMinor: '0', rolloverMax: money('0'), createdAt: '2026-08-08T00:00:00Z', updatedAt: '2026-08-08T00:00:00Z' },
        { id: 'combo-b', name: 'B', description: '', price: money('1000'), validityDays: 30, trafficLimitBytes: '1', resetStrategy: 'MONTH', includedSquads: [beta], active: true, rolloverMinRemainingBps: 0, rolloverMaxTxbMinor: '0', rolloverMax: money('0'), createdAt: '2026-08-08T00:00:00Z', updatedAt: '2026-08-08T00:00:00Z' },
      ],
    } satisfies Catalog
    getCatalog.mockResolvedValue(catalog)
    getBalance.mockResolvedValue({ balance: money('5000'), paymentMethods: [] })
    getCouponWallet.mockResolvedValue({ items: [] })
    const scope = effectScope()
    const state = scope.run(() => useCatalog())!
    await state.load()

    state.toggleSquad(alpha.id)
    expect(state.selectedSquadIds.value).toEqual([])
    state.toggleSquad(beta.id)
    expect(state.selectedSquadIds.value).toEqual([beta.id])

    state.selectCombo('combo-b')
    expect(state.selectedSquadIds.value).toEqual([])
    expect(state.includedSquadIds.value).toEqual([beta.id])
    scope.stop()
  })

  it('reuses one purchase idempotency key after an ambiguous failed attempt', async () => {
    const catalog = {
      addons: [],
      combos: [
        { id: 'combo-a', name: 'A', description: '', price: money('1000'), validityDays: 30, trafficLimitBytes: '1', resetStrategy: 'MONTH', includedSquads: [], active: true, rolloverMinRemainingBps: 0, rolloverMaxTxbMinor: '0', rolloverMax: money('0'), createdAt: '2026-08-08T00:00:00Z', updatedAt: '2026-08-08T00:00:00Z' },
      ],
    } satisfies Catalog
    getCatalog.mockResolvedValue(catalog)
    getBalance.mockResolvedValue({ balance: money('5000'), paymentMethods: [] })
    getCouponWallet.mockResolvedValue({ items: [] })
    createPurchase
      .mockRejectedValueOnce(new Error('connection closed after commit'))
      .mockResolvedValueOnce({ id: 'purchase-1' })

    const scope = effectScope()
    const state = scope.run(() => useCatalog())!
    await state.load()
    state.reviewPurchase()

    expect(await state.confirmPurchase()).toBe(false)
    expect(await state.confirmPurchase()).toBe(true)
    expect(createPurchase).toHaveBeenCalledTimes(2)
    expect(createPurchase.mock.calls[0]?.[3]).toBeTruthy()
    expect(createPurchase.mock.calls[1]?.[3]).toBe(createPurchase.mock.calls[0]?.[3])
    scope.stop()
  })
})
