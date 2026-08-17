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
  <div class="flex flex-col w-full gap-5">
    <div class="payment-step-heading flex items-center gap-3">
      <UButton
        class="payment-step-heading__back shrink-0"
        color="neutral"
        variant="ghost"
        icon="i-ph-arrow-left"
        :aria-label="t('payment.backToProvider')"
        :label="t('payment.backToProvider')"
        data-haptic
        @click="emit('back')"
      />
      <h2 class="text-lg font-semibold flex-1 truncate">{{ t('payment.chooseChannel') }}</h2>
    </div>
    <TxbAmountField
      id="txb-amount"
      class="w-full"
      :model-value="amount"
      :label="t('payment.amount')"
      :hint="t('payment.minimumTopUp')"
      min-minor="100"
      required
      @update:model-value="emit('update:amount', String($event))"
    />
    <UAlert v-if="!channels.length" class="w-full" color="warning" variant="soft" icon="i-ph-warning-circle" :description="t('payment.noChannel')" />
    <UFormField v-else class="w-full" :label="t('payment.channel')">
      <URadioGroup
        class="w-full flex flex-col gap-3"
        :model-value="selectedMethodId ?? undefined"
        :items="channels"
        orientation="vertical"
        variant="card"
        @update:model-value="emit('chooseMethod', String($event))"
      />
    </UFormField>
    <UAlert v-if="error" class="w-full" color="error" variant="soft" :description="error" />
    <UButton
      class="payment-submit w-full mt-2"
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