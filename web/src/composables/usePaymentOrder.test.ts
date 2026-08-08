import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeaturePaymentOrder } from '@/api/features'

const { createPaymentOrder, getPaymentOrder, cancelPaymentOrder } = vi.hoisted(() => ({
  createPaymentOrder: vi.fn(),
  getPaymentOrder: vi.fn(),
  cancelPaymentOrder: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  api: { createPaymentOrder, getPaymentOrder, cancelPaymentOrder },
}))

vi.mock('qrcode', () => ({
  default: { toDataURL: vi.fn().mockResolvedValue('data:image/png;base64,qr') },
}))

import { isTerminalPaymentStatus, usePaymentOrder } from './usePaymentOrder'

const pendingOrder: FeaturePaymentOrder = {
  id: 'payment-1',
  methodId: 'ezpay:alipay',
  provider: 'ezpay',
  providerRail: 'alipay',
  status: 'pending',
  txb: { currency: 'TXB', minor: '2000', display: '20.00 TXB' },
  payableAmount: '20.00',
  payableCurrency: 'CNY',
  rateSnapshot: '1',
  rateDirection: 'txb_per_currency',
  paymentUrl: 'https://pay.example/order-1',
  qrPayload: 'https://pay.example/order-1',
  receivingAddress: null,
  actualCryptoAmount: null,
  actualCryptoCurrency: null,
  expiresAt: '2026-08-07T01:00:00Z',
  paidAt: null,
  refundedAt: null,
  cancelledAt: null,
  cancelReason: '',
  providerCancelStatus: '',
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

    payment.chooseMethod('ezpay:alipay')
    await payment.createOrder()
    expect(createPaymentOrder).toHaveBeenCalledWith('ezpay:alipay', '2000')
    expect(payment.stage.value).toBe('pending')
    expect(onPaid).not.toHaveBeenCalled()

    await vi.advanceTimersByTimeAsync(2000)
    expect(payment.stage.value).toBe('paid')
    expect(onPaid).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(700)
    expect(onPaid).toHaveBeenCalledOnce()
    scope.stop()
  })

  it('stops polling after a user-owned cancellation', async () => {
    createPaymentOrder.mockResolvedValue(pendingOrder)
    cancelPaymentOrder.mockResolvedValue({ ...pendingOrder, status: 'cancelled' })
    const scope = effectScope()
    const payment = scope.run(() => usePaymentOrder({ onPaid: vi.fn() }))!

    payment.chooseMethod('ezpay:alipay')
    await payment.createOrder()
    await payment.cancelOrder()

    expect(cancelPaymentOrder).toHaveBeenCalledWith('payment-1')
    expect(payment.stage.value).toBe('cancelled')
    scope.stop()
  })

  it('keeps QR-only and address-only provider checkouts usable', async () => {
    createPaymentOrder
      .mockResolvedValueOnce({ ...pendingOrder, paymentUrl: null, qrPayload: 'usdt:wallet?amount=1', receivingAddress: null })
      .mockResolvedValueOnce({ ...pendingOrder, paymentUrl: null, qrPayload: null, receivingAddress: 'TExampleAddress' })
    const firstScope = effectScope()
    const qrPayment = firstScope.run(() => usePaymentOrder({ onPaid: vi.fn() }))!
    qrPayment.chooseMethod('bepusdt:usdt.trc20')
    await qrPayment.createOrder()
    expect(qrPayment.stage.value).toBe('pending')
    expect(qrPayment.qrDataUrl.value).toContain('data:image/png')
    firstScope.stop()

    const secondScope = effectScope()
    const addressPayment = secondScope.run(() => usePaymentOrder({ onPaid: vi.fn() }))!
    addressPayment.chooseMethod('bepusdt:usdt.trc20')
    await addressPayment.createOrder()
    expect(addressPayment.stage.value).toBe('pending')
    expect(addressPayment.order.value?.receivingAddress).toBe('TExampleAddress')
    secondScope.stop()
  })

  it('honors a paid response that wins the cancellation race', async () => {
    vi.useFakeTimers()
    createPaymentOrder.mockResolvedValue(pendingOrder)
    cancelPaymentOrder.mockResolvedValue({ ...pendingOrder, status: 'paid', paidAt: '2026-08-07T00:01:00Z' })
    const onPaid = vi.fn()
    const scope = effectScope()
    const payment = scope.run(() => usePaymentOrder({ onPaid }))!
    payment.chooseMethod('ezpay:alipay')
    await payment.createOrder()
    await payment.cancelOrder()

    expect(payment.stage.value).toBe('paid')
    await vi.advanceTimersByTimeAsync(700)
    expect(onPaid).toHaveBeenCalledOnce()
    scope.stop()
  })
})
