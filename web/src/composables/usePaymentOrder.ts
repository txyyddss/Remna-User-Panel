import { computed, onScopeDispose, readonly, shallowRef } from 'vue'
import QRCode from 'qrcode'

import { api } from '@/api/client'
import type { PaymentMethod, PaymentOrder, PaymentProvider } from '@/api/types'
import { moneyFromTxbInput } from '@/utils/format'
import { getTelegramWebApp, notifyHaptic, openExternalLink } from '@/utils/telegram'

export type PaymentStage = 'configure' | 'creating' | 'pending' | 'paid'

export function isTerminalPaymentStatus(status: PaymentOrder['status']): boolean {
  return ['paid', 'expired', 'failed', 'refunded'].includes(status)
}

export function usePaymentOrder(options: { onPaid: () => void }) {
  const amount = shallowRef('20.00')
  const selectedProvider = shallowRef<PaymentProvider | null>(null)
  const stage = shallowRef<PaymentStage>('configure')
  const order = shallowRef<PaymentOrder | null>(null)
  const qrDataUrl = shallowRef<string | null>(null)
  const error = shallowRef<string | null>(null)
  let pollTimer: ReturnType<typeof setTimeout> | undefined
  let closeTimer: ReturnType<typeof setTimeout> | undefined

  const amountMinor = computed(() => moneyFromTxbInput(amount.value))
  const amountValid = computed(() => amountMinor.value !== '' && BigInt(amountMinor.value) >= 100n)
  const canCreate = computed(() => amountValid.value && selectedProvider.value !== null && stage.value === 'configure')

  function stopPolling(): void {
    if (pollTimer !== undefined) clearTimeout(pollTimer)
    pollTimer = undefined
  }

  function reset(methods: readonly PaymentMethod[]): void {
    stopPolling()
    if (closeTimer !== undefined) clearTimeout(closeTimer)
    order.value = null
    qrDataUrl.value = null
    error.value = null
    stage.value = 'configure'
    selectedProvider.value = methods.find((method) => method.available)?.provider ?? null
  }

  function chooseProvider(provider: PaymentProvider): void {
    if (stage.value !== 'configure') return
    selectedProvider.value = provider
  }

  async function createOrder(): Promise<void> {
    if (!canCreate.value || !selectedProvider.value) return
    stage.value = 'creating'
    error.value = null
    try {
      order.value = await api.createPaymentOrder(selectedProvider.value, amountMinor.value)
      if (!order.value.qrPayload) throw new Error('The provider did not return a payment QR payload.')
      qrDataUrl.value = await QRCode.toDataURL(order.value.qrPayload, {
        width: 360,
        margin: 1,
        color: { dark: '#111512', light: '#edf2ee' },
        errorCorrectionLevel: 'M',
      })
      stage.value = order.value.status === 'paid' ? 'paid' : 'pending'

      if (order.value.provider === 'stars' && order.value.paymentUrl) {
        getTelegramWebApp()?.openInvoice(order.value.paymentUrl)
      }

      if (stage.value === 'paid') handlePaid()
      else schedulePoll()
    } catch (caught) {
      stage.value = 'configure'
      error.value = caught instanceof Error ? caught.message : 'The payment order could not be created.'
      notifyHaptic('error')
    }
  }

  function schedulePoll(): void {
    stopPolling()
    pollTimer = setTimeout(() => void poll(), 2000)
  }

  async function poll(): Promise<void> {
    if (!order.value || stage.value !== 'pending') return
    try {
      const refreshed = await api.getPaymentOrder(order.value.id)
      order.value = refreshed
      if (refreshed.status === 'paid') {
        handlePaid()
        return
      }
      if (isTerminalPaymentStatus(refreshed.status)) {
        error.value = refreshed.status === 'expired'
          ? 'This payment order expired. Start a new one when you are ready.'
          : `Payment status: ${refreshed.status.toLowerCase()}.`
        stage.value = 'configure'
        return
      }
    } catch {
      // A transient polling failure should not discard a valid order.
    }
    schedulePoll()
  }

  function handlePaid(): void {
    stopPolling()
    stage.value = 'paid'
    notifyHaptic('success')
    closeTimer = setTimeout(options.onPaid, 700)
  }

  function openPaymentTarget(): void {
    const target = order.value?.paymentUrl
    if (target) openExternalLink(target)
  }

  onScopeDispose(() => {
    stopPolling()
    if (closeTimer !== undefined) clearTimeout(closeTimer)
  })

  return {
    amount,
    selectedProvider,
    stage: readonly(stage),
    order: readonly(order),
    qrDataUrl: readonly(qrDataUrl),
    error: readonly(error),
    amountMinor,
    amountValid,
    canCreate,
    reset,
    chooseProvider,
    createOrder,
    openPaymentTarget,
    stopPolling,
  }
}
