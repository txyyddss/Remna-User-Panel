import { effectScope, shallowRef } from 'vue'
import { describe, expect, it, vi } from 'vitest'

import type { AutoRenewal } from '@/api/types'

const { getAutoRenewal, setAutoRenewal } = vi.hoisted(() => ({
  getAutoRenewal: vi.fn(),
  setAutoRenewal: vi.fn(),
}))

vi.mock('@/api/client', () => ({ api: { getAutoRenewal, setAutoRenewal } }))

import { useAutoRenewal } from './useAutoRenewal'

const ineligibleRenewal: AutoRenewal = {
  purchaseId: 'purchase-1',
  enabled: false,
  canEnable: false,
  ineligibleReason: 'INSUFFICIENT_BALANCE',
  grossPrice: { currency: 'TXB', minor: '1000', display: '10.00 TXB' },
  discount: { currency: 'TXB', minor: '0', display: '0.00 TXB' },
  netPrice: { currency: 'TXB', minor: '1000', display: '10.00 TXB' },
  scheduledAt: '2026-08-14T00:00:00Z',
  nextCycleEndsAt: '2026-09-13T00:00:00Z',
}

describe('useAutoRenewal', () => {
  it('keeps an ineligible automatic-renewal toggle disabled', async () => {
    getAutoRenewal.mockResolvedValue(ineligibleRenewal)
    const purchaseID = shallowRef('purchase-1')
    const scope = effectScope()
    const state = scope.run(() => useAutoRenewal(() => purchaseID.value, () => false))!

    await state.load()

    expect(state.canEnable.value).toBe(false)
    expect(state.ineligibleReason.value).toBe('INSUFFICIENT_BALANCE')
    expect(await state.setEnabled(true)).toBe(false)
    expect(setAutoRenewal).not.toHaveBeenCalled()
    scope.stop()
  })
})
