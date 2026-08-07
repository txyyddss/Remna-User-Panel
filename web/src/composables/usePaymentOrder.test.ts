import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { PaymentOrder } from '@/api/types'

const { createPaymentOrder, getPaymentOrder } = vi.hoisted(() => ({
  createPaymentOrder: vi.fn(),
  getPaymentOrder: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  api: { createPaymentOrder, getPaymentOrder },
}))

vi.mock('qrcode', () => ({
  default: { toDataURL: vi.fn().mockResolvedValue('data:image/png;base64,qr') },
}))

import { isTerminalPaymentStatus, usePaymentOrder } from './usePaymentOrder'

const pendingOrder: PaymentOrder = {
  id: 'payment-1',
  provider: 'ezpay',
  status: 'pending',
  txb: { currency: 'TXB', minor: '2000', display: '20.00 TXB' },
  payableAmount: '20.00',
  payableCurrency: 'CNY',
  rateSnapshot: '1',
  paymentUrl: 'https://pay.example/order-1',
  qrPayload: 'https://pay.example/order-1',
  expiresAt: '2026-08-07T01:00:00Z',
  paidAt: null,
  refundedAt: null,
  createdAt: '2026-08-07T00:00:00Z',
  updatedAt: '2026-08-07T00:00:00Z',
}

describe('payment order polling', () => {
  afterEach(() => {
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  it('recognizes authoritative terminal provider states', () => {
    expect(isTerminalPaymentStatus('paid')).toBe(true)
    expect(isTerminalPaymentStatus('refunded')).toBe(true)
    expect(isTerminalPaymentStatus('pending')).toBe(false)
  })

  it('closes only after the server reports paid', async () => {
    vi.useFakeTimers()
    createPaymentOrder.mockResolvedValue(pendingOrder)
    getPaymentOrder.mockResolvedValue({ ...pendingOrder, status: 'paid', paidAt: '2026-08-07T00:01:00Z' })
    const onPaid = vi.fn()
    const scope = effectScope()
    const payment = scope.run(() => usePaymentOrder({ onPaid }))!

    payment.chooseProvider('ezpay')
    await payment.createOrder()
    expect(payment.stage.value).toBe('pending')
    expect(onPaid).not.toHaveBeenCalled()

    await vi.advanceTimersByTimeAsync(2000)
    expect(payment.stage.value).toBe('paid')
    expect(onPaid).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(700)
    expect(onPaid).toHaveBeenCalledOnce()
    scope.stop()
  })
})
