<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import type { FeaturePaymentMethod } from '@/api/features'
import { featuresApi } from '@/api/features'
import type { PaymentProvider } from '@/api/types'
import type { PaymentStage } from '@/composables/usePaymentOrder'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { localizedError, useI18n } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'

const props = defineProps<{
  methods: readonly FeaturePaymentMethod[]
  selectedMethodId: string | null
  stage: PaymentStage
  error: string | null
  amountValid: boolean
  canCreate: boolean
  canReissue: boolean
}>()
const emit = defineEmits<{
  chooseMethod: [id: string]
  createOrder: []
  paid: []
}>()
const amount = defineModel<string>('amount', { required: true })
const selectedProvider = defineModel<PaymentProvider | null>('selectedProvider', { required: true })
const { t } = useI18n()
const couponCode = shallowRef('')
const couponBusy = shallowRef(false)
const couponError = shallowRef<string | null>(null)
const externalMethods = computed(() => props.methods.filter((method) => method.mode === 'order'))
const couponMethod = computed(() => props.methods.find((method) => method.mode === 'coupon_redemption'))
const providers = computed(() => [...new Set(props.methods.map((method) => method.provider))])
const channels = computed(() => externalMethods.value.filter((method) => method.provider === selectedProvider.value))
const icons: Record<PaymentProvider, string> = {
  ezpay: 'i-ph-credit-card',
  bepusdt: 'i-ph-currency-circle-dollar',
  stars: 'i-ph-telegram-logo',
  coupon: 'i-ph-ticket',
}

function providerLabel(provider: PaymentProvider): string {
  return t(`payment.providers.${provider}`)
}
function providerNote(provider: PaymentProvider): string {
  if (provider === 'coupon') return t('payment.couponHint')
  return externalMethods.value.some((method) => method.provider === provider && method.available) ? '' : t('payment.rateUnavailable')
}
function methodNote(method: FeaturePaymentMethod): string {
  return method.available ? '' : t('payment.rateUnavailable')
}
function chooseProvider(provider: PaymentProvider): void {
  selectedProvider.value = provider
  couponError.value = null
  const first = externalMethods.value.find((method) => method.provider === provider && method.available)
  if (first) emit('chooseMethod', first.id)
}
async function redeemCoupon(): Promise<void> {
  if (!couponCode.value.trim() || couponBusy.value) return
  couponBusy.value = true
  couponError.value = null
  try {
    await featuresApi.redeemCoupon(couponCode.value, createUuid())
    emit('paid')
  } catch (caught) {
    couponError.value = localizedError(caught, 'errors.couponRedeem')
  } finally {
    couponBusy.value = false
  }
}
</script>

<template>
  <TxbAmountField v-if="selectedProvider !== 'coupon'" id="txb-amount" v-model="amount" :label="$t('payment.amount')" :hint="$t('payment.minimumTopUp')" min-minor="100" required />
  <fieldset class="provider-picker">
    <legend>{{ $t('payment.provider') }}</legend>
    <UButton v-for="provider in providers" :key="provider" class="provider-option" :class="{ 'provider-option--selected': selectedProvider === provider }" color="neutral" variant="ghost" :disabled="provider !== 'coupon' && !externalMethods.some((method) => method.provider === provider && method.available)" :aria-pressed="selectedProvider === provider" data-haptic @click="chooseProvider(provider)">
      <span class="provider-option__icon"><UIcon :name="icons[provider]" /></span>
      <span><strong>{{ providerLabel(provider) }}</strong><small>{{ providerNote(provider) }}</small></span>
    </UButton>
  </fieldset>
  <UAlert v-if="selectedProvider !== 'coupon' && !externalMethods.some((method) => method.available)" color="warning" variant="soft" icon="i-ph-warning-circle" :description="$t('payment.noChannel')" />
  <fieldset v-if="selectedProvider !== 'coupon' && channels.length" class="channel-picker">
    <legend>{{ $t('payment.channel') }}</legend>
    <UButton v-for="method in channels" :key="method.id" class="channel-option" :class="{ 'channel-option--selected': selectedMethodId === method.id }" color="neutral" variant="ghost" :disabled="!method.available" :aria-pressed="selectedMethodId === method.id" data-haptic @click="emit('chooseMethod', method.id)">
      <span><strong>{{ method.name }}</strong><small>{{ methodNote(method) }}</small></span>
      <UBadge v-if="!method.available" color="neutral" variant="soft" :label="$t('common.unavailable')" />
    </UButton>
  </fieldset>
  <div v-if="selectedProvider === 'coupon'" class="coupon-redemption">
    <UInput v-model="couponCode" :placeholder="$t('payment.couponCode')" :aria-label="$t('payment.couponCode')" />
    <UAlert v-if="couponError" color="error" variant="soft" :description="couponError" />
    <UButton block :disabled="!couponCode.trim() || couponBusy" :loading="couponBusy" :label="$t('payment.redeemCoupon')" data-haptic @click="redeemCoupon" />
  </div>
  <UAlert v-if="error && selectedProvider !== 'coupon'" color="error" variant="soft" :description="error" />
  <UButton v-if="selectedProvider !== 'coupon'" block :disabled="!canCreate || stage === 'creating' || !amountValid" :loading="stage === 'creating'" :label="stage === 'creating' ? $t('payment.creating') : canReissue ? $t('payment.reissue') : $t('payment.continue')" data-haptic @click="emit('createOrder')" />
</template>

