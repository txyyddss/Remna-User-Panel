<script setup lang="ts">
import { computed, onScopeDispose, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { PaymentReturnProvider } from '@/api/features'
import { useI18n } from '@/i18n'
import { usePaymentReturn } from '@/composables/usePaymentReturn'
import { isTelegramWebAppDetected } from '@/utils/telegram'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const orderId = shallowRef('')
const provider = shallowRef<PaymentReturnProvider | null>(null)
const capability = shallowRef('')
const browserReturn = !isTelegramWebAppDetected()
let homeTimer: ReturnType<typeof globalThis.setTimeout> | undefined

watch(() => [route.query.paymentOrder, route.query.provider, route.query.paymentCapability], ([rawOrder, rawProvider, rawCapability]) => {
  provider.value = rawProvider === 'ezpay' || rawProvider === 'bepusdt' ? rawProvider : null
  capability.value = typeof rawCapability === 'string' ? rawCapability : ''
  orderId.value = typeof rawOrder === 'string' ? rawOrder : ''
}, { immediate: true })

const { state, order, orderStatus, isConfirmed, refresh } = usePaymentReturn(orderId, { browserStatus: browserReturn, provider, capability })
const title = computed(() => ({
  checking: t('payment.returnCheckingTitle'),
  pending: t('payment.returnCheckingTitle'),
  confirmed: t('payment.returnConfirmedTitle'),
  terminal: t('payment.returnTerminalTitle'),
  missing: t('payment.returnMissingTitle'),
  unavailable: t('payment.returnUnavailableTitle'),
})[state.value])
const description = computed(() => state.value === 'terminal'
  ? t('payment.returnTerminalDescription', { status: t(`payment.status.${orderStatus.value ?? order.value?.status ?? 'failed'}`) })
  : t(`payment.return${state.value[0]?.toUpperCase()}${state.value.slice(1)}Description`))
const icon = computed(() => isConfirmed.value ? 'i-ph-check-circle-fill' : state.value === 'terminal' ? 'i-ph-x-circle-fill' : 'i-ph-spinner-gap')
const canReissue = computed(() => order.value?.status === 'expired' || order.value?.status === 'failed')

function reissue(): void {
  if (!canReissue.value || !order.value) return
  void router.replace({ name: 'home', query: { reissue: order.value.id } })
}

watch(isConfirmed, (confirmed) => {
  if (homeTimer !== undefined) globalThis.clearTimeout(homeTimer)
  homeTimer = confirmed && !browserReturn ? globalThis.setTimeout(() => void router.replace('/home'), 2600) : undefined
}, { immediate: true })
onScopeDispose(() => {
  if (homeTimer !== undefined) globalThis.clearTimeout(homeTimer)
})
</script>

<template>
  <main class="payment-return">
    <section class="payment-return__card" :class="{ 'payment-return__card--confirmed': isConfirmed }" role="status" aria-live="polite">
      <span class="payment-return__icon" :class="{ 'payment-return__icon--pending': !isConfirmed && state !== 'terminal' }"><UIcon :name="icon" /></span>
      <h1>{{ title }}</h1>
      <p>{{ description }}</p>
      <p v-if="isConfirmed && !browserReturn" class="payment-return__hint">{{ $t('payment.returnAutoHome') }}</p>
      <UButton v-if="isConfirmed && !browserReturn" block trailing-icon="i-ph-house" :label="$t('payment.returnHome')" @click="router.replace('/home')" />
      <UButton v-else-if="canReissue" block trailing-icon="i-ph-arrow-clockwise" :label="$t('payment.reissue')" @click="reissue" />
      <UButton v-else-if="state === 'checking' || state === 'pending'" block color="neutral" variant="outline" :loading="true" :label="$t('payment.returnCheckingAction')" @click="refresh" />
      <UButton v-else-if="!browserReturn" block color="neutral" variant="outline" :label="$t('payment.returnHome')" @click="router.replace('/home')" />
    </section>
  </main>
</template>

<style scoped>
.payment-return { min-height: 100dvh; display: grid; place-items: center; padding: calc(2rem + var(--tg-content-safe-area-inset-top, env(safe-area-inset-top))) 1.25rem calc(2rem + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom))); background: var(--canvas); }
.payment-return__card { width: min(100%, 28rem); display: grid; justify-items: center; gap: 0.85rem; padding: 1.5rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface); text-align: center; }
.payment-return__card--confirmed { border-color: var(--accent); background: var(--accent-soft); }
.payment-return__icon { width: 64px; height: 64px; display: inline-grid; place-items: center; border-radius: 18px; color: var(--danger); background: var(--danger-soft); font-size: 2.3rem; }
.payment-return__card--confirmed .payment-return__icon { color: var(--accent-ink); background: var(--accent); }
.payment-return__icon--pending { color: var(--warning); background: var(--warning-soft); }
.payment-return__icon--pending :deep(svg) { animation: icon-spin 900ms linear infinite; }
.payment-return__card h1, .payment-return__card p { margin: 0; }
.payment-return__card h1 { font-size: 1.35rem; letter-spacing: -0.03em; }
.payment-return__card p { max-width: 24rem; color: var(--text-muted); font-size: 0.84rem; line-height: 1.55; }
.payment-return__card .payment-return__hint { color: var(--text-faint); font-size: 0.7rem; }
</style>
