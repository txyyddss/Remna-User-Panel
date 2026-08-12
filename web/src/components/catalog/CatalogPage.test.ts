import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { nextTick, shallowRef } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const catalogMock = vi.hoisted(() => ({ useCatalog: vi.fn() }))

vi.mock('@/composables/useCatalog', () => catalogMock)
vi.mock('@/stores/session', () => ({
  useSessionStore: () => ({ user: { id: 'user-1' } }),
}))

import CatalogPage from './CatalogPage.vue'

const combo = {
  id: 'combo-1',
  name: 'Combo',
  description: '',
  price: { currency: 'TXB' as const, minor: '1000', display: '10.00 TXB' },
  validityDays: 30,
  trafficLimitBytes: '1',
  resetStrategy: 'MONTH' as const,
  includedSquads: [],
  active: true,
  rolloverMinRemainingBps: 0,
  rolloverMaxTxbMinor: '0',
  rolloverMax: { currency: 'TXB' as const, minor: '0', display: '0.00 TXB' },
  createdAt: '2026-08-08T00:00:00Z',
  updatedAt: '2026-08-08T00:00:00Z',
}

async function mountPage(purchase: object | null, step = 5) {
  const refreshQuote = vi.fn().mockResolvedValue(true)
  catalogMock.useCatalog.mockReturnValue({
    catalog: shallowRef({ combos: [combo], addons: [] }),
    balance: shallowRef({ currency: 'TXB', minor: '1000', display: '10.00 TXB' }),
    loading: shallowRef(false),
    purchasing: shallowRef(false),
    quoting: shallowRef(false),
    error: shallowRef(null),
    purchase: shallowRef(purchase),
    quote: shallowRef(null),
    selectedComboId: shallowRef(combo.id),
    selectedSquadIds: shallowRef<string[]>([]),
    selectedCouponGrantId: shallowRef<string | null>(null),
    couponDiscarding: shallowRef(false),
    needsBalance: shallowRef(false),
    visibleCombos: shallowRef([combo]),
    visibleSquads: shallowRef([]),
    selectedCombo: shallowRef(combo),
    selectedSquads: shallowRef([]),
    includedSquadIds: shallowRef<string[]>([]),
    couponGrants: shallowRef([]),
    eligibleCoupons: shallowRef([]),
    load: vi.fn().mockResolvedValue(undefined),
    selectCombo: vi.fn(),
    toggleSquad: vi.fn(),
    refreshQuote,
    discardCoupon: vi.fn(),
    confirmPurchase: vi.fn(),
  })

  sessionStorage.setItem('txc-catalog-step:user-1', String(step))
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/catalog', component: { template: '<div />' } }],
  })
  await router.push('/catalog')
  await router.isReady()
  const wrapper = mount(CatalogPage, {
    global: {
      plugins: [router],
      stubs: {
        CatalogCheckout: true,
        CatalogCouponStep: true,
        CatalogFlowControls: true,
        CatalogFlowProgress: true,
        CatalogNodes: true,
        SquadSelector: true,
      },
    },
  })
  await nextTick()
  return { refreshQuote, wrapper }
}

describe('CatalogPage quote restoration', () => {
  beforeEach(() => {
    sessionStorage.clear()
    catalogMock.useCatalog.mockReset()
  })

  it('does not request another quote after purchase confirmation', async () => {
    const { refreshQuote, wrapper } = await mountPage({ id: 'purchase-1' })

    await nextTick()

    expect(refreshQuote).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('restores accessible nodes after returning to the catalog on step 3', async () => {
    const { refreshQuote, wrapper } = await mountPage(null, 3)

    await nextTick()

    expect(refreshQuote).toHaveBeenCalled()
    wrapper.unmount()
  })
})
