import { onScopeDispose, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { FeaturePaymentOrder } from '@/api/features'
import { notifyHaptic, openExternalLink, openTelegramInvoice } from '@/utils/telegram'
import { isTerminalPaymentStatus, paymentQrDataUrl } from './paymentOrderHelpers'

export type PaymentTargetResult = 'paid' | 'pending' | 'terminal' | 'missing'

interface PaymentTargetOptions {
  onPaid: () => void
  onTerminal: (order: FeaturePaymentOrder) => void
}

export function usePaymentTarget(options: PaymentTargetOptions) {
  const order = shallowRef<FeaturePaymentOrder | null>(null)
  const qrDataUrl = shallowRef<string | null>(null)
  let timer: ReturnType<typeof setTimeout> | undefined

  function stop(): void {
    if (timer !== undefined) clearTimeout(timer)
    timer = undefined
  }

  async function remember(next: FeaturePaymentOrder): Promise<void> {
    order.value = next
    qrDataUrl.value = await paymentQrDataUrl(next)
  }

  async function present(next: FeaturePaymentOrder): Promise<PaymentTargetResult> {
    await remember(next)
    if (next.status === 'paid') {
      options.onPaid()
      return 'paid'
    }
    if (isTerminalPaymentStatus(next.status)) {
      options.onTerminal(next)
      return 'terminal'
    }
    if (!next.paymentUrl && !next.qrPayload && !next.receivingAddress) return 'missing'
    if (next.provider === 'stars' && next.paymentUrl
      && !openTelegramInvoice(next.paymentUrl)) openExternalLink(next.paymentUrl)
    schedule()
    return 'pending'
  }

  function schedule(): void {
    stop()
    timer = setTimeout(() => void poll(), 2000)
  }

  async function poll(): Promise<void> {
    if (!order.value) return
    try {
      const refreshed = await api.getPaymentOrder(order.value.id)
      order.value = refreshed
      if (refreshed.status === 'paid') {
        options.onPaid()
        return
      }
      if (isTerminalPaymentStatus(refreshed.status)) {
        options.onTerminal(refreshed)
        return
      }
    } catch {
      // A transient status failure must not discard a valid checkout.
    }
    schedule()
  }

  function reset(): void {
    stop()
    order.value = null
    qrDataUrl.value = null
  }

  function open(): void {
    if (order.value?.paymentUrl) openExternalLink(order.value.paymentUrl)
  }

  function cancelled(): void {
    notifyHaptic('success')
  }

  onScopeDispose(stop)
  return { order: readonly(order), qrDataUrl: readonly(qrDataUrl), remember, present, reset, stop, open, cancelled }
}
