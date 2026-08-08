<script setup lang="ts">
import type { Component } from 'vue'
import { computed, shallowRef, watch } from 'vue'
import {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui'
import {
  PhArrowSquareOut,
  PhCheckCircle,
  PhCreditCard,
  PhCurrencyCircleDollar,
  PhTelegramLogo,
  PhX,
} from '@phosphor-icons/vue'

import type { FeaturePaymentMethod } from '@/api/features'
import type { PaymentProvider } from '@/api/types'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { usePaymentOrder } from '@/composables/usePaymentOrder'
import { formatDateTime, formatMoney } from '@/utils/format'

const props = defineProps<{ methods: readonly FeaturePaymentMethod[] }>()
const emit = defineEmits<{ paid: [] }>()
const open = defineModel<boolean>('open', { required: true })
const selectedProvider = shallowRef<PaymentProvider | null>(null)

const methods = computed(() => props.methods)
const providers = computed(() => {
  const seen = new Set<PaymentProvider>()
  return methods.value.filter((method) => {
    if (seen.has(method.provider)) return false
    seen.add(method.provider)
    return true
  })
})
const channels = computed(() => methods.value.filter((method) => method.provider === selectedProvider.value))

const {
  amount,
  selectedMethodId,
  stage,
  order,
  qrDataUrl,
  error,
  amountValid,
  canCreate,
  reset,
  chooseMethod,
  createOrder,
  cancelOrder,
  openPaymentTarget,
  stopPolling,
} = usePaymentOrder({
  onPaid: () => {
    open.value = false
    emit('paid')
  },
})

const icons: Record<PaymentProvider, Component> = {
  ezpay: PhCreditCard,
  bepusdt: PhCurrencyCircleDollar,
  stars: PhTelegramLogo,
}

function chooseProvider(provider: PaymentProvider): void {
  selectedProvider.value = provider
  const first = methods.value.find((method) => method.provider === provider && method.available)
  if (first) chooseMethod(first.id)
}

watch(open, (next) => {
  if (next) {
    reset(methods.value)
    selectedProvider.value = methods.value.find((method) => method.available)?.provider ?? null
  } else stopPolling()
})
</script>

<template>
  <DialogRoot v-model:open="open">
    <DialogPortal to="#overlays">
      <DialogOverlay class="dialog-overlay" />
      <DialogContent class="dialog-content payment-sheet" @escape-key-down="['pending', 'cancelling'].includes(stage) && $event.preventDefault()" @pointer-down-outside="['pending', 'cancelling'].includes(stage) && $event.preventDefault()">
        <div class="dialog-handle" aria-hidden="true" />
        <header class="dialog-header">
          <div>
            <DialogTitle class="dialog-title">Add TXB</DialogTitle>
            <DialogDescription class="dialog-description">
              {{ stage === 'configure' ? 'Choose an amount, provider, and payment channel.' : stage === 'cancelled' ? 'This order is no longer being polled.' : 'Use the exact provider details below.' }}
            </DialogDescription>
          </div>
          <DialogClose v-if="!['pending', 'cancelling'].includes(stage)" class="icon-button" aria-label="Close payment"><PhX :size="20" /></DialogClose>
        </header>

        <template v-if="stage === 'configure' || stage === 'creating'">
          <TxbAmountField id="txb-amount" v-model="amount" label="TXB amount" hint="Minimum top-up is 1.00 TXB." min-minor="100" required />

          <fieldset class="provider-picker">
            <legend>Provider</legend>
            <button v-for="provider in providers" :key="provider.provider" class="provider-option" :class="{ 'provider-option--selected': selectedProvider === provider.provider }" type="button" :disabled="!methods.some((method) => method.provider === provider.provider && method.available)" :aria-pressed="selectedProvider === provider.provider" @click="chooseProvider(provider.provider)">
              <span class="provider-option__icon"><component :is="icons[provider.provider]" :size="22" /></span><span><strong>{{ provider.provider === 'bepusdt' ? 'USDT' : provider.name }}</strong><small>{{ provider.provider === 'ezpay' ? 'Local payment rails' : provider.note }}</small></span>
            </button>
          </fieldset>
          <p v-if="!methods.some((method) => method.available)" class="field-error" role="status">No payment channel is currently enabled. An administrator must enter the new TXB-per-currency rates.</p>

          <fieldset v-if="channels.length" class="channel-picker">
            <legend>Channel</legend>
            <button v-for="method in channels" :key="method.id" class="channel-option" :class="{ 'channel-option--selected': selectedMethodId === method.id }" type="button" :disabled="!method.available" :aria-pressed="selectedMethodId === method.id" @click="chooseMethod(method.id)"><span><strong>{{ method.name }}</strong><small>{{ method.note }}</small></span><span v-if="!method.available" class="status-badge">Unavailable</span></button>
          </fieldset>

          <p v-if="error" class="field-error" role="alert">{{ error }}</p>
          <button class="button button--primary button--wide" type="button" :disabled="!canCreate || stage === 'creating' || !amountValid" @click="createOrder">{{ stage === 'creating' ? 'Creating secure order' : 'Continue to payment' }}</button>
        </template>

        <template v-else-if="(stage === 'pending' || stage === 'cancelling') && order">
          <div class="payment-amount"><span>Exact amount</span><strong>{{ order.actualCryptoAmount ?? order.payableAmount }} {{ order.actualCryptoCurrency ?? order.payableCurrency }}</strong><small>Credits {{ formatMoney(order.txb) }} after confirmation</small></div>
          <div v-if="qrDataUrl" class="qr-frame"><img :src="qrDataUrl" alt="Payment QR code" width="260" height="260" /></div>
          <div v-if="order.receivingAddress" class="payment-address"><span>Receiving address</span><code>{{ order.receivingAddress }}</code></div>
          <p class="payment-expiry">Order expires {{ formatDateTime(order.expiresAt) }}</p>
          <button v-if="order.paymentUrl" class="button button--secondary button--wide" type="button" @click="openPaymentTarget">Open payment page<PhArrowSquareOut :size="19" /></button>
          <div class="payment-waiting" role="status"><span class="payment-waiting__pulse" />{{ stage === 'cancelling' ? 'Cancelling with provider' : 'Waiting for provider confirmation' }}</div>
          <button class="button button--ghost-danger button--wide" type="button" :disabled="stage === 'cancelling'" @click="cancelOrder">Cancel this order</button>
          <p class="field-hint">A later authoritative paid callback can still credit this order once.</p>
          <p v-if="error" class="field-error" role="alert">{{ error }}</p>
        </template>

        <div v-else-if="stage === 'cancelled'" class="payment-success" role="status"><PhX :size="50" /><h2>Order cancelled.</h2><p>Polling stopped. You can close this sheet.</p></div>
        <div v-else class="payment-success" role="status"><PhCheckCircle :size="50" weight="fill" /><h2>Balance added.</h2><p>Your TXB is ready to use.</p></div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<style scoped>
.channel-picker { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.5rem; margin: 0 0 1rem; padding: 0; border: 0; }
.channel-picker legend { grid-column: 1 / -1; margin-bottom: 0.1rem; color: var(--text-muted); font-size: 0.78rem; font-weight: 700; }
.channel-option { min-height: 58px; display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; padding: 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); color: var(--text); background: var(--surface-raised); text-align: left; cursor: pointer; }
.channel-option--selected { border-color: #557763; background: var(--accent-soft); }
.channel-option strong, .channel-option small { display: block; }
.channel-option strong { font-size: 0.75rem; }
.channel-option small { margin-top: 0.2rem; color: var(--text-faint); font-size: 0.62rem; }
</style>
