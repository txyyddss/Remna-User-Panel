import { computed, onScopeDispose, readonly, shallowRef } from 'vue'
import QRCode from 'qrcode'

import type { FeaturePaymentMethod, FeaturePaymentOrder } from '@/api/features'
import { api } from '@/api/client'
import { localizedError, t } from '@/i18n'
import { moneyFromTxbInput } from '@/utils/format'
import { getTelegramWebApp, notifyHaptic, openExternalLink } from '@/utils/telegram'

export type PaymentStage = 'configure' | 'creating' | 'pending' | 'cancelling' | 'cancelled' | 'paid'

export function isTerminalPaymentStatus(status: FeaturePaymentOrder['status']): boolean {
  return ['paid', 'cancelled', 'expired', 'failed', 'refunded'].includes(status)
}

export function usePaymentOrder(options: { onPaid: () => void }) {
  const amount = shallowRef('20.00')
  const selectedMethodId = shallowRef<string | null>(null)
  const stage = shallowRef<PaymentStage>('configure')
  const order = shallowRef<FeaturePaymentOrder | null>(null)
  const qrDataUrl = shallowRef<string | null>(null)
  const error = shallowRef<string | null>(null)
  let pollTimer: ReturnType<typeof setTimeout> | undefined
  let closeTimer: ReturnType<typeof setTimeout> | undefined

  const amountMinor = computed(() => moneyFromTxbInput(amount.value))
  const amountValid = computed(() => amountMinor.value !== '' && BigInt(amountMinor.value) >= 100n)
  const canCreate = computed(() => amountValid.value && selectedMethodId.value !== null && stage.value === 'configure')

  function stopPolling(): void {
    if (pollTimer !== undefined) clearTimeout(pollTimer)
    pollTimer = undefined
  }

  function reset(methods: readonly FeaturePaymentMethod[]): void {
    stopPolling()
    if (closeTimer !== undefined) clearTimeout(closeTimer)
    order.value = null
    qrDataUrl.value = null
    error.value = null
    stage.value = 'configure'
    selectedMethodId.value = methods.find((method) => method.available)?.id ?? null
  }

  function chooseMethod(methodId: string): void {
    if (stage.value !== 'configure') return
    selectedMethodId.value = methodId
  }

  async function createOrder(): Promise<void> {
    if (!canCreate.value || !selectedMethodId.value) return
    stage.value = 'creating'
    error.value = null
    try {
      order.value = await api.createPaymentOrder(selectedMethodId.value, amountMinor.value)
      qrDataUrl.value = order.value.qrPayload
        ? await QRCode.toDataURL(order.value.qrPayload, {
          width: 360,
          margin: 1,
          color: { dark: '#111512', light: '#edf2ee' },
          errorCorrectionLevel: 'M',
        })
        : null
      if (!order.value.paymentUrl && !order.value.qrPayload && !order.value.receivingAddress) {
        stage.value = 'configure'
        error.value = t('payment.targetMissing')
        notifyHaptic('error')
        return
      }
      stage.value = order.value.status === 'paid' ? 'paid' : 'pending'

      if (order.value.provider === 'stars' && order.value.paymentUrl) {
        getTelegramWebApp()?.openInvoice(order.value.paymentUrl)
      }

      if (stage.value === 'paid') handlePaid()
      else schedulePoll()
    } catch (caught) {
      stage.value = 'configure'
      error.value = localizedError(caught, 'errors.paymentCreate')
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
          ? t('payment.orderExpired')
          : t('payment.terminalStatus', { status: t(`payment.status.${refreshed.status}`) })
        stage.value = refreshed.status === 'cancelled' ? 'cancelled' : 'configure'
        return
      }
    } catch {
      // A transient polling failure should not discard a valid order.
    }
    schedulePoll()
  }

  async function cancelOrder(): Promise<void> {
    if (!order.value || stage.value !== 'pending') return
    stopPolling()
    stage.value = 'cancelling'
    error.value = null
    try {
      order.value = await api.cancelPaymentOrder(order.value.id)
      stage.value = order.value.status === 'paid' ? 'paid' : 'cancelled'
      if (stage.value === 'paid') handlePaid()
      else notifyHaptic('success')
    } catch (caught) {
      stage.value = 'pending'
      error.value = localizedError(caught, 'errors.paymentCancel')
      schedulePoll()
    }
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
    selectedMethodId: readonly(selectedMethodId),
    stage: readonly(stage),
    order: readonly(order),
    qrDataUrl: readonly(qrDataUrl),
    error: readonly(error),
    amountMinor,
    amountValid,
    canCreate,
    reset,
    chooseMethod,
    createOrder,
    cancelOrder,
    openPaymentTarget,
    stopPolling,
  }
}
