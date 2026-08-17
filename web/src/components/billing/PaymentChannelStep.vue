<script setup lang="ts">
import type { PaymentStage } from '@/composables/usePaymentOrder'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useI18n } from '@/i18n'
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
}>()

const emit = defineEmits<{
  'update:amount': [value: string]
  chooseMethod: [id: string]
  back: []
  createOrder: []
}>()

const { t } = useI18n()
</script>

<template>
  <div class="payment-step-heading">
    <UButton
      class="payment-step-heading__back"
      color="neutral"
      variant="ghost"
      icon="i-ph-arrow-left"
      :aria-label="t('payment.backToProvider')"
      :label="t('payment.backToProvider')"
      data-haptic
      @click="emit('back')"
    />
    <h2>{{ t('payment.chooseChannel') }}</h2>
  </div>
  <TxbAmountField
    id="txb-amount"
    :model-value="amount"
    :label="t('payment.amount')"
    :hint="t('payment.minimumTopUp')"
    min-minor="100"
    required
    @update:model-value="emit('update:amount', String($event))"
  />
  <UAlert v-if="!channels.length" color="warning" variant="soft" icon="i-ph-warning-circle" :description="t('payment.noChannel')" />
  <fieldset v-else class="provider-picker" role="radiogroup">
    <legend>{{ t('payment.channel') }}</legend>
    <UButton
      v-for="item in channels"
      :key="item.value"
      class="provider-option"
      :class="{ 'provider-option--selected': selectedMethodId === item.value }"
      color="neutral"
      variant="ghost"
      :disabled="Boolean((item as any).disabled || (item as any).available === false)"
      :aria-pressed="selectedMethodId === item.value"
      data-haptic
      @click="emit('chooseMethod', item.value)"
    >
      <span v-if="(item as any).icon" class="provider-option__icon"><UIcon :name="(item as any).icon" aria-hidden="true" /></span>
      <span><strong>{{ item.label }}</strong><small v-if="item.description">{{ item.description }}</small></span>
      <UIcon v-if="selectedMethodId === item.value" class="provider-option__check" name="i-ph-check-circle-fill" aria-hidden="true" />
    </UButton>
  </fieldset>
  <UAlert v-if="error" color="error" variant="soft" :description="error" />
  <UButton
    class="payment-submit"
    data-test="payment-submit"
    block
    :disabled="!canCreate || stage === 'creating' || !amountValid"
    :loading="stage === 'creating'"
    :label="stage === 'creating' ? t('payment.creating') : canReissue ? t('payment.reissue') : t('payment.proceedToPayment')"
    data-haptic
    @click="emit('createOrder')"
  />
</template>