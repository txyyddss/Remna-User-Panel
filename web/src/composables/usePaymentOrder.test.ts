import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeaturePaymentOrder } from '@/api/features'
import type { OperationReceipt, PaymentOperation } from '@/api/types'
import { setLocale, t } from '@/i18n'

const { createPaymentOrder, getPaymentOrder, cancelPaymentOrder, getOperation } = vi.hoisted(() => ({
  createPaymentOrder: vi.fn(),
  getPaymentOrder: vi.fn(),
  cancelPaymentOrder: vi.fn(),
  getOperation: vi.fn(),
}))

vi.mock('@/api/client', () => ({ api: { createPaymentOrder, getPaymentOrder, cancelPaymentOrder } }))
vi.mock('@/api/memberOperations', () => ({ memberOperationsApi: { getOperation } }))
vi.mock('qrcode', () => ({
  default: { toDataURL: vi.fn().mockResolvedValue('data:image/png;base64,qr') },
}))

import { isTerminalPaymentStatus, usePaymentOrder } from './usePaymentOrder'

const timestamp = '2026-08-07T00:00:00Z'
const pendingOrder: FeaturePaymentOrder = {
  id: 'payment-1', methodId: 'ezpay:alipay', provider: 'ezpay', providerRail: 'alipay', status: 'pending',
  txb: { currency: 'TXB', minor: '2000', display: '20.00 TXB' }, payableAmount: '20.00', payableCurrency: 'CNY',
  rateSnapshot: '1', rateDirection: 'txb_per_currency', paymentUrl: 'https://pay.example/order-1',
  qrPayload: 'https://pay.example/order-1', receivingAddress: null, actualCryptoAmount: null,
  actualCryptoCurrency: null, expiresAt: '2026-08-07T01:00:00Z', paidAt: null, refundedAt: null,
  cancelledAt: null, cancelReason: '', providerCancelStatus: '', createdAt: timestamp, updatedAt: timestamp,
}

function receipt(kind: 'payment_create' | 'payment_cancel', status: OperationReceipt['status']): OperationReceipt {
  return { id: `${kind}-1`, kind, status, createdAt: timestamp, updatedAt: timestamp }
}

function accepted(kind: 'payment_create' | 'payment_cancel'): PaymentOperation {
  return { paymentOrderId: pendingOrder.id, operation: receipt(kind, 'queued') }
}

async function finishOperation(kind: 'payment_create' | 'payment_cancel'): Promise<void> {
  getOperation.mockResolvedValueOnce(receipt(kind, 'succeeded'))
  await vi.advanceTimersByTimeAsync(1500)
}

describe('payment order polling', () => {
  afterEach(() => {
    vi.useRealTimers()
    vi.clearAllMocks()
    setLocale('en')
  })

  it('recognizes authoritative terminal provider states', () => {
    expect(isTerminalPaymentStatus('paid')).toBe(true)
    expect(isTerminalPaymentStatus('refunded')).toBe(true)
    expect(isTerminalPaymentStatus('pending')).toBe(false)
  })

  it('opens a checkout only after its durable create operation succeeds', async () => {
    vi.useFakeTimers()
    createPaymentOrder.mockResolvedValue(accepted('payment_create'))
    getPaymentOrder
      .mockResolvedValueOnce(pendingOrder)
      .mockResolvedValueOnce({ ...pendingOrder, status: 'paid', paidAt: '2026-08-07T00:01:00Z' })
    const onPaid = vi.fn()
    const scope = effectScope()
    const payment = scope.run(() => usePaymentOrder({ onPaid }))!

    payment.chooseMethod('ezpay:alipay')
    await payment.createOrder()
    expect(payment.stage.value).toBe('creating')
    await finishOperation('payment_create')
    expect(payment.stage.value).toBe('pending')

    await vi.advanceTimersByTimeAsync(2000)
    expect(payment.stage.value).toBe('paid')
    await vi.advanceTimersByTimeAsync(700)
    expect(onPaid).toHaveBeenCalledOnce()
    scope.stop()
  })

  it('reuses one create key after an ambiguous request failure', async () => {
    createPaymentOrder.mockRejectedValueOnce(new Error('network')).mockResolvedValueOnce(accepted('payment_create'))
    const scope = effectScope()
    const payment = scope.run(() => usePaymentOrder({ onPaid: vi.fn() }))!
    payment.chooseMethod('ezpay:alipay')

    await payment.createOrder()
    await payment.createOrder()

    const firstKey = createPaymentOrder.mock.calls[0]?.[2]
    expect(firstKey).toMatch(/^[0-9a-f-]{36}$/)
    expect(createPaymentOrder.mock.calls[1]?.[2]).toBe(firstKey)
    scope.stop()
  })

  it('queues cancellation and waits for its receipt', async () => {
    vi.useFakeTimers()
    createPaymentOrder.mockResolvedValue(accepted('payment_create'))
    cancelPaymentOrder.mockResolvedValue(accepted('payment_cancel'))
    getPaymentOrder
      .mockResolvedValueOnce(pendingOrder)
      .mockResolvedValueOnce({ ...pendingOrder, status: 'cancelled', cancelledAt: timestamp })
    const scope = effectScope()
    const payment = scope.run(() => usePaymentOrder({ onPaid: vi.fn() }))!

    payment.chooseMethod('ezpay:alipay')
    await payment.createOrder()
    await finishOperation('payment_create')
    await payment.cancelOrder()
    expect(payment.stage.value).toBe('cancelling')
    await finishOperation('payment_cancel')

    expect(payment.stage.value).toBe('cancelled')
    expect(cancelPaymentOrder.mock.calls[0]?.[0]).toBe('payment-1')
    expect(cancelPaymentOrder.mock.calls[0]?.[1]).toMatch(/^[0-9a-f-]{36}$/)
    scope.stop()
  })

  it('blocks duplicate payment when provider outcome needs review', async () => {
    vi.useFakeTimers()
    createPaymentOrder.mockResolvedValue(accepted('payment_create'))
    getOperation.mockResolvedValueOnce(receipt('payment_create', 'pending_review'))
    getPaymentOrder.mockResolvedValue({ ...pendingOrder, status: 'creating', paymentUrl: null, qrPayload: null })
    const scope = effectScope()
    const payment = scope.run(() => usePaymentOrder({ onPaid: vi.fn() }))!

    payment.chooseMethod('ezpay:alipay')
    await payment.createOrder()
    await vi.advanceTimersByTimeAsync(1500)

    expect(payment.stage.value).toBe('review')
    expect(payment.canCreate.value).toBe(false)
    expect(payment.error.value).toBe(t('payment.operationStatus', { status: t('operations.status.pending_review') }))
    scope.stop()
  })

  it('uses localized copy when a succeeded operation has no payment target', async () => {
    vi.useFakeTimers()
    setLocale('zh-CN')
    createPaymentOrder.mockResolvedValue(accepted('payment_create'))
    getPaymentOrder.mockResolvedValue({ ...pendingOrder, paymentUrl: null, qrPayload: null, receivingAddress: null })
    const scope = effectScope()
    const payment = scope.run(() => usePaymentOrder({ onPaid: vi.fn() }))!

    payment.chooseMethod('ezpay:alipay')
    await payment.createOrder()
    await finishOperation('payment_create')

    expect(payment.stage.value).toBe('review')
    expect(payment.error.value).toBe(t('payment.targetMissing'))
    scope.stop()
  })
})
