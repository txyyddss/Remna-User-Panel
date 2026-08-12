<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { FeaturePaymentMethod, FeaturePaymentOrder } from '@/api/features'
import type { PaymentProvider } from '@/api/types'
import BalancePaymentConfiguration from '@/components/billing/BalancePaymentConfiguration.vue'
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
const externalMethods = computed(() => methods.value.filter((method) => method.mode === 'order'))
const couponMethod = computed(() => methods.value.find((method) => method.mode === 'coupon_redemption'))

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

const description = computed(() => stage.value === 'configure'
  ? t('payment.configureHint')
  : stage.value === 'cancelled'
    ? t('payment.cancelledHint')
    : t('payment.providerHint'))

function prepareOrder(): void {
  reset(externalMethods.value)
  if (props.reissueOrder && hydrateReissueOrder(props.reissueOrder, externalMethods.value)) selectedProvider.value = props.reissueOrder.provider
  else selectedProvider.value = externalMethods.value.find((method) => method.available)?.provider ?? couponMethod.value?.provider ?? null
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
        <BalancePaymentConfiguration
          v-model:amount="amount"
          v-model:selected-provider="selectedProvider"
          :methods="methods"
          :selected-method-id="selectedMethodId"
          :stage="stage"
          :error="error"
          :amount-valid="amountValid"
          :can-create="canCreate"
          :can-reissue="canReissue"
          @choose-method="chooseMethod"
          @create-order="createOrder"
          @paid="emit('paid')"
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
        <UButton v-if="order.paymentUrl" block color="neutral" variant="outline" trailing-icon="i-ph-arrow-square-out" :label="$t('payment.openPage')" data-haptic @click="openPaymentTarget" />
        <div class="payment-waiting" role="status"><span class="payment-waiting__pulse" />{{ stage === 'cancelling' ? $t('payment.cancelling') : $t('payment.waiting') }}</div>
        <UButton block color="error" variant="ghost" :disabled="stage === 'cancelling'" :label="$t('payment.cancelOrder')" data-haptic="medium" @click="cancelOrder" />
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
