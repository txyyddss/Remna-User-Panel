<script setup lang="ts">
import { watch } from 'vue'
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
  PhCoins,
  PhCreditCard,
  PhCurrencyCircleDollar,
  PhTelegramLogo,
  PhX,
} from '@phosphor-icons/vue'

import type { Component } from 'vue'
import type { PaymentMethod, PaymentProvider } from '@/api/types'
import { usePaymentOrder } from '@/composables/usePaymentOrder'
import { formatDateTime, formatMoney } from '@/utils/format'

const props = defineProps<{
  methods: readonly PaymentMethod[]
}>()

const emit = defineEmits<{ paid: [] }>()
const open = defineModel<boolean>('open', { required: true })

const {
  amount,
  selectedProvider,
  stage,
  order,
  qrDataUrl,
  error,
  amountValid,
  canCreate,
  reset,
  chooseProvider,
  createOrder,
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

watch(open, (next) => {
  if (next) reset(props.methods)
  else stopPolling()
})
</script>

<template>
  <DialogRoot v-model:open="open">
    <DialogPortal to="#overlays">
      <DialogOverlay class="dialog-overlay" />
      <DialogContent class="dialog-content payment-sheet" @escape-key-down="stage === 'pending' && $event.preventDefault()" @pointer-down-outside="stage === 'pending' && $event.preventDefault()">
        <div class="dialog-handle" aria-hidden="true" />
        <header class="dialog-header">
          <div>
            <DialogTitle class="dialog-title">Add TXB</DialogTitle>
            <DialogDescription class="dialog-description">
              {{ stage === 'configure' ? 'Choose an amount and payment method.' : 'Keep this screen open until payment is confirmed.' }}
            </DialogDescription>
          </div>
          <DialogClose v-if="stage !== 'pending'" class="icon-button" aria-label="Close payment"><PhX :size="20" /></DialogClose>
        </header>

        <template v-if="stage === 'configure' || stage === 'creating'">
          <div class="amount-field">
            <label class="field-label" for="txb-amount">TXB amount</label>
            <div class="amount-input">
              <PhCoins :size="24" weight="fill" />
              <input id="txb-amount" v-model="amount" inputmode="decimal" type="text" autocomplete="off" aria-describedby="amount-help" />
              <span>TXB</span>
            </div>
            <p id="amount-help" class="field-hint">Minimum top-up is 1.00 TXB.</p>
          </div>

          <fieldset class="provider-picker">
            <legend>Payment method</legend>
            <button
              v-for="method in methods"
              :key="method.provider"
              class="provider-option"
              :class="{ 'provider-option--selected': selectedProvider === method.provider }"
              type="button"
              :disabled="!method.available"
              :aria-pressed="selectedProvider === method.provider"
              @click="chooseProvider(method.provider)"
            >
              <span class="provider-option__icon"><component :is="icons[method.provider]" :size="22" /></span>
              <span>
                <strong>{{ method.name }}</strong>
                <small>{{ method.note }}</small>
              </span>
              <span v-if="!method.available" class="status-badge">Unavailable</span>
            </button>
          </fieldset>

          <p v-if="error" class="field-error" role="alert">{{ error }}</p>
          <button class="button button--primary button--wide" type="button" :disabled="!canCreate || stage === 'creating' || !amountValid" @click="createOrder">
            {{ stage === 'creating' ? 'Creating secure order' : 'Continue to payment' }}
          </button>
        </template>

        <template v-else-if="stage === 'pending' && order">
          <div class="payment-amount">
            <span>Exact amount</span>
            <strong>{{ order.payableAmount }} {{ order.payableCurrency }}</strong>
            <small>Credits {{ formatMoney(order.txb) }} after confirmation</small>
          </div>

          <div class="qr-frame">
            <img v-if="qrDataUrl" :src="qrDataUrl" alt="Payment QR code" width="260" height="260" />
            <div v-else class="skeleton-block" aria-label="Generating QR code" />
          </div>

          <div v-if="order.provider === 'bepusdt' && order.qrPayload" class="payment-address">
            <span>Receiving address</span>
            <code>{{ order.qrPayload }}</code>
          </div>
          <p class="payment-expiry">Order expires {{ formatDateTime(order.expiresAt) }}</p>

          <button v-if="order.paymentUrl" class="button button--secondary button--wide" type="button" @click="openPaymentTarget">
            Open payment page
            <PhArrowSquareOut :size="19" />
          </button>
          <div class="payment-waiting" role="status">
            <span class="payment-waiting__pulse" />
            Waiting for provider confirmation
          </div>
          <p v-if="error" class="field-error" role="alert">{{ error }}</p>
        </template>

        <div v-else class="payment-success" role="status">
          <PhCheckCircle :size="50" weight="fill" />
          <h2>Balance added.</h2>
          <p>Your TXB is ready to use.</p>
        </div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>
