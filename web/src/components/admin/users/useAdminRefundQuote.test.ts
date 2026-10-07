import { effectScope, nextTick, shallowRef } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { AdminEntitlementRefundQuote } from '@/api/adminOperations'
const mocks = vi.hoisted(() => ({ getEntitlementRefundQuote: vi.fn() }))
vi.mock('@/api/adminOperations', () => ({ adminOperationsApi: mocks }))
import { useAdminRefundQuote } from './useAdminRefundQuote'

const quote: AdminEntitlementRefundQuote = { purchaseId: 'purchase-1', paid: { currency: 'TXB', minor: '10000', display: '100.00 TXB' }, suggestedRefund: { currency: 'TXB', minor: '7500', display: '75.00 TXB' }, allocatedTrafficBytes: '1000', usedTrafficBytes: '250', quotedAt: '2026-10-07T00:00:00Z', reasonCode: null }
async function settle(): Promise<void> { await nextTick(); await Promise.resolve(); await nextTick() }
describe('admin refund default', () => {
  beforeEach(() => vi.resetAllMocks())
  it('prefills the suggested amount and preserves edits on retry', async () => {
    mocks.getEntitlementRefundQuote.mockResolvedValue(quote)
    const scope = effectScope()
    const state = scope.run(() => useAdminRefundQuote(shallowRef(true), () => ({ id: 'purchase-1', userId: 'user-1' })))!
    await settle(); expect(state.amount.value).toBe('75.00')
    state.amount.value = '50.00'; await state.retry()
    expect(state.amount.value).toBe('50.00'); scope.stop()
  })
  it('does not overwrite a manual amount when a delayed quote arrives', async () => {
    let resolve!: (value: AdminEntitlementRefundQuote) => void
    mocks.getEntitlementRefundQuote.mockReturnValue(new Promise<AdminEntitlementRefundQuote>(done => { resolve = done }))
    const scope = effectScope()
    const state = scope.run(() => useAdminRefundQuote(shallowRef(true), () => ({ id: 'purchase-1', userId: 'user-1' })))!
    state.amount.value = '60.00'; resolve(quote); await settle()
    expect(state.amount.value).toBe('60.00'); scope.stop()
  })
  it('ignores an old entitlement response after the target changes', async () => {
    let resolve!: (value: AdminEntitlementRefundQuote) => void
    mocks.getEntitlementRefundQuote.mockReturnValueOnce(new Promise<AdminEntitlementRefundQuote>(done => { resolve = done })).mockResolvedValueOnce({ ...quote, purchaseId: 'purchase-2', suggestedRefund: { ...quote.paid, minor: '4000' } })
    const target = shallowRef({ id: 'purchase-1', userId: 'user-1' })
    const scope = effectScope()
    const state = scope.run(() => useAdminRefundQuote(shallowRef(true), () => target.value))!
    target.value = { id: 'purchase-2', userId: 'user-1' }; await settle()
    resolve(quote); await settle()
    expect(state.amount.value).toBe('40.00'); scope.stop()
  })
  it('leaves the amount empty on unavailable usage while allowing manual entry', async () => {
    mocks.getEntitlementRefundQuote.mockRejectedValue(new Error('offline'))
    const scope = effectScope()
    const state = scope.run(() => useAdminRefundQuote(shallowRef(true), () => ({ id: 'purchase-1', userId: 'user-1' })))!
    await settle(); expect(state.amount.value).toBe(''); expect(state.error.value).toBeTruthy()
    state.amount.value = '30.00'; expect(state.amount.value).toBe('30.00'); scope.stop()
  })
})
