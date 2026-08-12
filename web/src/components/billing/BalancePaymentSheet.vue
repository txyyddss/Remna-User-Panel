<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { CouponRedemption, FeaturePaymentMethod, FeaturePaymentOrder } from '@/api/features'
import BalancePaymentConfiguration from '@/components/billing/BalancePaymentConfiguration.vue'
import { usePaymentOrder } from '@/composables/usePaymentOrder'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useI18n } from '@/i18n'
import { formatDateTime, formatMoney } from '@/utils/format'

const props = defineProps<{
  methods: readonly FeaturePaymentMethod[]
  reissueOrder?: FeaturePaymentOrder | null
}>()
const emit = defineEmits<{ paid: [] }>()
const open = defineModel<boolean>('open', { required: true })
const { t } = useI18n()

const methods = computed(() => props.methods)
const externalMethods = computed(() => methods.value.filter((method) => method.mode === 'order'))
const couponRedemption = shallowRef<CouponRedemption | null>(null)

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
  couponRedemption.value = null
  reset(externalMethods.value)
  if (props.reissueOrder) hydrateReissueOrder(props.reissueOrder, externalMethods.value)
}

function closeSheet(): void {
  open.value = false
}

const showTelegramBack = computed(() => open.value)
useTelegramBackButton(showTelegramBack, closeSheet)

watch(open, (next, previous) => {
  if (next) prepareOrder()
  else {
    stopPolling()
    if (previous && couponRedemption.value) emit('paid')
  }
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
      <template v-if="!couponRedemption && (stage === 'configure' || stage === 'creating')">
        <BalancePaymentConfiguration
          v-model:amount="amount"
          :methods="methods"
          :selected-method-id="selectedMethodId"
          :stage="stage"
          :error="error"
          :amount-valid="amountValid"
          :can-create="canCreate"
          :can-reissue="canReissue"
          @choose-method="chooseMethod"
          @create-order="createOrder"
          @coupon-redeemed="couponRedemption = $event"
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

      <div v-else-if="couponRedemption" class="payment-success payment-success--coupon" role="status">
        <UIcon name="i-ph-check-circle-fill" aria-hidden="true" />
        <h2>{{ $t('payment.added') }}</h2>
        <div class="payment-amount">
          <span>{{ $t('payment.couponGain') }}</span>
          <strong>{{ formatMoney({ currency: 'TXB', minor: couponRedemption.balanceDeltaMinor, display: '' }) }}</strong>
          <small>{{ $t('payment.balanceAfter', { amount: formatMoney({ currency: 'TXB', minor: couponRedemption.balanceAfterMinor, display: '' }) }) }}</small>
        </div>
        <UButton block :label="$t('common.close')" data-haptic @click="closeSheet" />
      </div>
      <div v-else-if="stage === 'cancelled'" class="payment-success payment-success--cancelled" role="status">
        <UIcon name="i-ph-x-circle-fill" class="payment-success__error-icon" aria-hidden="true" />
        <h2>{{ $t('payment.cancelled') }}</h2>
        <p>{{ $t('payment.cancelledClose') }}</p>
      </div>
      <div v-else class="payment-success" role="status"><UIcon name="i-ph-check-circle-fill" /><h2>{{ $t('payment.added') }}</h2><p>{{ $t('payment.ready') }}</p></div>
    </template>
  </UModal>
</template>

<style scoped>
.payment-success--coupon { gap: 0.8rem; }
.payment-success--coupon > .payment-amount { width: 100%; }
.payment-success__error-icon { width: 5rem; height: 5rem; color: var(--danger); font-size: 5rem; }
</style>
