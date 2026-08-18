import { effectScope } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { OperationReceipt, TrafficResetQuote } from '@/api/types'

const apiMocks = vi.hoisted(() => ({
  getOperation: vi.fn(),
  getPurchaseRefundQuote: vi.fn(),
  getTrafficResetQuote: vi.fn(),
  refundPurchase: vi.fn(),
  resetPurchaseTraffic: vi.fn(),
}))
const createUuid = vi.hoisted(() => vi.fn())

vi.mock('@/api/memberOperations', () => ({ memberOperationsApi: apiMocks }))
vi.mock('@/utils/browserCompatibility', () => ({ createUuid }))
vi.mock('@/utils/telegram', () => ({ notifyHaptic: vi.fn() }))

import { ApiError } from '@/api/client'
import { usePurchaseOperations } from './usePurchaseOperations'

const quote: TrafficResetQuote = {
  purchaseId: 'purchase-1', eligible: true, reasonCode: null,
  price: { currency: 'TXB', minor: '34', display: '0.34 TXB' },
  resetStrategy: 'DAY', quotedAt: '2026-08-18T00:00:00Z',
}
const receipt: OperationReceipt = {
  id: 'operation-1', kind: 'purchase_traffic_reset', status: 'succeeded', errorCode: null,
  createdAt: '2026-08-18T00:00:00Z', updatedAt: '2026-08-18T00:00:01Z', completedAt: '2026-08-18T00:00:01Z',
}

describe('usePurchaseOperations', () => {
  beforeEach(() => {
    createUuid.mockReset()
    createUuid.mockReturnValueOnce('operation-key-1').mockReturnValueOnce('operation-key-2')
  })
  afterEach(() => vi.clearAllMocks())

  it('refreshes a conflicting quote before allowing confirmation again', async () => {
    apiMocks.getTrafficResetQuote.mockResolvedValue(quote)
    apiMocks.resetPurchaseTraffic
      .mockRejectedValueOnce(new ApiError(409, { code: 'OPERATION_CONFLICT', message: 'conflict' }))
      .mockResolvedValueOnce(receipt)
    const scope = effectScope()
    const state = scope.run(() => usePurchaseOperations(() => 'purchase-1'))!

    expect(await state.start('reset')).toBe(false)
    expect(state.resetQuote.value).toEqual(quote)
    expect(await state.start('reset')).toBe(true)

    expect(apiMocks.resetPurchaseTraffic).toHaveBeenNthCalledWith(1, 'purchase-1', 'operation-key-1')
    expect(apiMocks.resetPurchaseTraffic).toHaveBeenNthCalledWith(2, 'purchase-1', 'operation-key-2')
    expect(state.receipt.value?.status).toBe('succeeded')
    scope.stop()
  })

  it('retains one key after an ambiguous transport failure', async () => {
    apiMocks.resetPurchaseTraffic.mockRejectedValueOnce(new Error('network')).mockResolvedValueOnce(receipt)
    const scope = effectScope()
    const state = scope.run(() => usePurchaseOperations(() => 'purchase-1'))!

    expect(await state.start('reset')).toBe(false)
    expect(await state.start('reset')).toBe(true)

    expect(apiMocks.resetPurchaseTraffic).toHaveBeenNthCalledWith(1, 'purchase-1', 'operation-key-1')
    expect(apiMocks.resetPurchaseTraffic).toHaveBeenNthCalledWith(2, 'purchase-1', 'operation-key-1')
    scope.stop()
  })

  it('uses the server refund quote as the eligibility source', async () => {
    apiMocks.getPurchaseRefundQuote.mockResolvedValue({
      purchaseId: 'purchase-1', eligible: true, reasonCode: null,
      refund: { currency: 'TXB', minor: '1000', display: '10.00 TXB' },
      quotedAt: '2026-08-18T00:00:00Z', eligibilityExpiresAt: '2026-08-18T12:00:00Z',
    })
    const scope = effectScope()
    const state = scope.run(() => usePurchaseOperations(() => 'purchase-1'))!

    await state.loadRefundEligibility()

    expect(state.refundQuote.value?.eligible).toBe(true)
    scope.stop()
  })
})
