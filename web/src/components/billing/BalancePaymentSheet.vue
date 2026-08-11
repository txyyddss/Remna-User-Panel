<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import type { FeaturePaymentMethod, FeaturePaymentOrder } from '@/api/features'
import type { PaymentProvider } from '@/api/types'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { usePaymentOrder } from '@/composables/usePaymentOrder'
import { useI18n } from '@/i18n'
import { formatDateTime, formatMoney } from '@/utils/format'

const props = defineProps<{
  methods: readonly FeaturePaymentMethod[]
  reissueOrder?: FeaturePaymentOrder | null
}>()
const emit = defineEmits<{ paid: [] }>()
const open = defineModel<boolean>('open', { required: true })
const selectedProvider = shallowRef<PaymentProvider | null>(null)
const { t } = useI18n()

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
  canReissue,
  amountValid,
  canCreate,
  reset,
  hydrateReissueOrder,
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

const icons: Record<PaymentProvider, string> = {
  ezpay: 'i-ph-credit-card',
  bepusdt: 'i-ph-currency-circle-dollar',
  stars: 'i-ph-telegram-logo',
}
const description = computed(() => stage.value === 'configure'
  ? t('payment.configureHint')
  : stage.value === 'cancelled'
    ? t('payment.cancelledHint')
    : t('payment.providerHint'))

function chooseProvider(provider: PaymentProvider): void {
  selectedProvider.value = provider
  const first = methods.value.find((method) => method.provider === provider && method.available)
  if (first) chooseMethod(first.id)
}

function providerNote(provider: FeaturePaymentMethod): string {
  if (provider.provider === 'ezpay') return t('payment.localRails')
  const available = methods.value.some((method) => method.provider === provider.provider && method.available)
  return available ? '' : t('payment.rateUnavailable')
}

function methodNote(method: FeaturePaymentMethod): string {
  return method.available ? '' : t('payment.rateUnavailable')
}

function prepareOrder(): void {
  reset(methods.value)
  if (props.reissueOrder && hydrateReissueOrder(props.reissueOrder, methods.value)) selectedProvider.value = props.reissueOrder.provider
  else selectedProvider.value = methods.value.find((method) => method.available)?.provider ?? null
}

watch(open, (next) => {
  if (next) prepareOrder()
  else stopPolling()
})
</script>

<template>
  <UModal
    v-model:open="open"
    :title="$t('payment.addTxb')"
    :description="description"
    :dismissible="!['pending', 'cancelling'].includes(stage)"
  >
    <template #body>
      <template v-if="stage === 'configure' || stage === 'creating'">
        <TxbAmountField
          id="txb-amount"
          v-model="amount"
          :label="$t('payment.amount')"
          :hint="$t('payment.minimumTopUp')"
          min-minor="100"
          required
        />

        <fieldset class="provider-picker">
          <legend>{{ $t('payment.provider') }}</legend>
          <UButton
            v-for="provider in providers"
            :key="provider.provider"
            class="provider-option"
            :class="{ 'provider-option--selected': selectedProvider === provider.provider }"
            color="neutral"
            variant="ghost"
            :disabled="!methods.some((method) => method.provider === provider.provider && method.available)"
            :aria-pressed="selectedProvider === provider.provider"
            @click="chooseProvider(provider.provider)"
          >
            <span class="provider-option__icon"><UIcon :name="icons[provider.provider]" /></span>
            <span>
              <strong>{{ provider.provider === 'bepusdt' ? 'USDT' : provider.name }}</strong>
              <small>{{ providerNote(provider) }}</small>
            </span>
          </UButton>
        </fieldset>
        <UAlert
          v-if="!methods.some((method) => method.available)"
          color="warning"
          variant="soft"
          icon="i-ph-warning-circle"
          :description="$t('payment.noChannel')"
        />

        <fieldset v-if="channels.length" class="channel-picker">
          <legend>{{ $t('payment.channel') }}</legend>
          <UButton
            v-for="method in channels"
            :key="method.id"
            class="channel-option"
            :class="{ 'channel-option--selected': selectedMethodId === method.id }"
            color="neutral"
            variant="ghost"
            :disabled="!method.available"
            :aria-pressed="selectedMethodId === method.id"
            @click="chooseMethod(method.id)"
          >
            <span><strong>{{ method.name }}</strong><small>{{ methodNote(method) }}</small></span>
            <UBadge v-if="!method.available" color="neutral" variant="soft" :label="$t('common.unavailable')" />
          </UButton>
        </fieldset>

        <UAlert v-if="error" color="error" variant="soft" :description="error" />
        <UButton
          block
          :disabled="!canCreate || stage === 'creating' || !amountValid"
          :loading="stage === 'creating'"
          :label="stage === 'creating' ? $t('payment.creating') : canReissue ? $t('payment.reissue') : $t('payment.continue')"
          @click="createOrder"
        />
      </template>

      <template v-else-if="(stage === 'pending' || stage === 'cancelling') && order">
        <div class="payment-amount">
          <span>{{ $t('payment.exactAmount') }}</span>
          <strong>{{ order.actualCryptoAmount ?? order.payableAmount }} {{ order.actualCryptoCurrency ?? order.payableCurrency }}</strong>
          <small>{{ $t('payment.creditsAfter', { amount: formatMoney(order.txb) }) }}</small>
        </div>
        <div v-if="qrDataUrl" class="qr-frame"><img :src="qrDataUrl" :alt="$t('payment.qrAlt')" width="260" height="260" /></div>
        <div v-if="order.receivingAddress" class="payment-address"><span>{{ $t('payment.receivingAddress') }}</span><code>{{ order.receivingAddress }}</code></div>
        <p class="payment-expiry">{{ $t('payment.expires', { date: formatDateTime(order.expiresAt) }) }}</p>
        <UButton v-if="order.paymentUrl" block color="neutral" variant="outline" trailing-icon="i-ph-arrow-square-out" :label="$t('payment.openPage')" @click="openPaymentTarget" />
        <div class="payment-waiting" role="status"><span class="payment-waiting__pulse" />{{ stage === 'cancelling' ? $t('payment.cancelling') : $t('payment.waiting') }}</div>
        <UButton block color="error" variant="ghost" :disabled="stage === 'cancelling'" :label="$t('payment.cancelOrder')" @click="cancelOrder" />
        <p class="field-hint">{{ $t('payment.callbackHint') }}</p>
        <UAlert v-if="error" color="error" variant="soft" :description="error" />
      </template>

      <div v-else-if="stage === 'cancelled'" class="payment-success" role="status"><UIcon name="i-ph-x" /><h2>{{ $t('payment.cancelled') }}</h2><p>{{ $t('payment.cancelledClose') }}</p></div>
      <div v-else class="payment-success" role="status"><UIcon name="i-ph-check-circle-fill" /><h2>{{ $t('payment.added') }}</h2><p>{{ $t('payment.ready') }}</p></div>
    </template>
  </UModal>
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
