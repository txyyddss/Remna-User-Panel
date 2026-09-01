<script setup lang="ts">
import { computed } from 'vue'

import type { PaymentStage } from '@/composables/usePaymentOrder'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useI18n } from '@/i18n'
import { txbInputFromMinor } from '@/utils/format'
import { selectionHaptic } from '@/utils/telegram'
import PaymentCryptoChannelPicker from './PaymentCryptoChannelPicker.vue'
import type { PaymentChannelOption } from './paymentOptions'

const props = defineProps<{
  channels: readonly PaymentChannelOption[]
  selectedMethodId: string | null
  amountValid: boolean
  amount: string
  stage: PaymentStage
  error: string | null
  canCreate: boolean
  canReissue: boolean
  minimumMinor: string
  maximumMinor: string
}>()

const emit = defineEmits<{
  'update:amount': [value: string]
  chooseMethod: [id: string]
  createOrder: []
  retryOperation: []
}>()

const { t } = useI18n()
const isCrypto = computed(() => props.channels.some((channel) => channel.cryptoCurrency))

function chooseMethod(id: string): void {
  if (id === props.selectedMethodId) return
  selectionHaptic()
  emit('chooseMethod', id)
}
</script>

<template>
  <div class="payment-channel-step">
    <TxbAmountField
      id="txb-amount"
      :model-value="amount"
      :label="t('payment.amount')"
      :hint="t('payment.amountRange', { min: txbInputFromMinor(minimumMinor), max: txbInputFromMinor(maximumMinor) })"
      :min-minor="minimumMinor"
      :max-minor="maximumMinor"
      slider
      required
      @update:model-value="emit('update:amount', String($event))"
    />
    <UAlert v-if="!channels.length" color="warning" variant="soft" icon="i-ph-warning-circle" :description="t('payment.noChannel')" />
    <PaymentCryptoChannelPicker
      v-else-if="isCrypto"
      :channels="channels"
      :selected-method-id="selectedMethodId"
      @choose="chooseMethod"
    />
    <fieldset v-else class="provider-picker" role="radiogroup">
      <legend class="sr-only">{{ t('payment.chooseChannel') }}</legend>
      <UButton
        v-for="item in channels"
        :key="item.value"
        class="provider-option"
        :class="{ 'provider-option--selected': selectedMethodId === item.value }"
        color="neutral"
        variant="ghost"
        :disabled="item.disabled"
        :aria-pressed="selectedMethodId === item.value"
        @click="chooseMethod(item.value)"
      >
        <span class="provider-option__icon">
          <img v-if="item.logo" :src="item.logo" alt="" width="28" height="28" />
          <UIcon v-else name="i-ph-credit-card" aria-hidden="true" />
        </span>
        <span><strong>{{ item.label }}</strong><small v-if="item.description">{{ item.description }}</small></span>
        <UIcon v-if="selectedMethodId === item.value" class="provider-option__check" name="i-ph-check-circle-fill" aria-hidden="true" />
      </UButton>
    </fieldset>
    <UAlert v-if="error" color="error" variant="soft" :description="error" />
    <UButton
      v-if="stage === 'creating' && error"
      class="payment-submit"
      block
      color="neutral"
      variant="outline"
      icon="i-ph-arrows-clockwise"
      :label="t('operations.checkStatus')"
      data-haptic="retry"
      @click="emit('retryOperation')"
    />
    <UButton
      v-else
      class="payment-submit"
      data-test="payment-submit"
      block
      :disabled="!canCreate || stage === 'creating' || !amountValid"
      :loading="stage === 'creating'"
      :label="stage === 'creating' ? t('payment.creating') : canReissue ? t('payment.reissue') : t('payment.proceedToPayment')"
      data-haptic="confirm"
      @click="emit('createOrder')"
    />
  </div>
</template>

<style scoped>
.payment-channel-step { display: grid; min-width: 0; gap: 0.75rem; }
.payment-channel-step :deep(.amount-field), .payment-channel-step .provider-picker, .payment-channel-step .payment-submit { margin: 0; }
</style>
