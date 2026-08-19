<script setup lang="ts">
import type { PaymentStage } from '@/composables/usePaymentOrder'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useI18n } from '@/i18n'
import { txbInputFromMinor } from '@/utils/format'
import type { PaymentChannelOption } from './paymentOptions'

defineProps<{
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
  back: []
  createOrder: []
  retryOperation: []
}>()

const { t } = useI18n()
</script>

<template>
  <div class="payment-channel-step">
    <div class="payment-step-heading">
      <UButton
        class="payment-step-heading__back"
        color="neutral"
        variant="ghost"
        icon="i-ph-arrow-left"
        :aria-label="t('payment.backToProvider')"
        data-haptic
        @click="emit('back')"
      />
    </div>
    <TxbAmountField
      id="txb-amount"
      :model-value="amount"
      :label="t('payment.amount')"
      :hint="t('payment.amountRange', { min: txbInputFromMinor(minimumMinor), max: txbInputFromMinor(maximumMinor) })"
      :min-minor="minimumMinor"
      :max-minor="maximumMinor"
      integer-only
      required
      @update:model-value="emit('update:amount', String($event))"
    />
    <UAlert v-if="!channels.length" color="warning" variant="soft" icon="i-ph-warning-circle" :description="t('payment.noChannel')" />
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
        data-haptic
        @click="emit('chooseMethod', item.value)"
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
      data-haptic
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
      data-haptic
      @click="emit('createOrder')"
    />
  </div>
</template>

<style scoped>
.payment-channel-step { display: grid; min-width: 0; gap: 0.75rem; }
.payment-channel-step :deep(.amount-field), .payment-channel-step .provider-picker, .payment-channel-step .payment-step-heading, .payment-channel-step .payment-submit { margin: 0; }
</style>
