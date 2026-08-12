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
const { t } = useI18n()
const couponCode = shallowRef('')
const couponBusy = shallowRef(false)
const couponError = shallowRef<string | null>(null)

const externalMethods = computed(() => props.methods.filter((method) => method.mode === 'order'))
const selectedMethod = computed(() => props.methods.find((method) => method.id === props.selectedMethodId) ?? null)
const selectedProvider = computed<PaymentProvider | null>(() => selectedMethod.value?.provider ?? null)
const selectedProfileKey = computed<string | undefined>({
  get: () => selectedMethod.value ? profileKey(selectedMethod.value) : undefined,
  set: (key) => chooseProfile(key),
})
const profileItems = computed(() => {
  const items: { label: string; value: string; description: string }[] = []
  const seen = new Set<string>()
  for (const method of externalMethods.value) {
    const value = profileKey(method)
    if (seen.has(value)) continue
    seen.add(value)
    items.push({
      label: method.providerName || t(`payment.providers.${method.provider}`),
      value,
      description: method.available ? t(`payment.providers.${method.provider}`) : t('payment.rateUnavailable'),
    })
  }
  const coupon = props.methods.find((method) => method.provider === 'coupon')
  if (coupon) items.push({ label: t('payment.providers.coupon'), value: 'coupon', description: t('payment.couponHint') })
  return items
})
const channels = computed(() => externalMethods.value.filter((method) => profileKey(method) === selectedProfileKey.value))
const channelItems = computed(() => channels.value.map((method) => ({
  label: method.name,
  value: method.id,
  description: method.available ? '' : methodNote(method),
  disabled: !method.available,
})))

function profileKey(method: FeaturePaymentMethod): string {
  return `${method.provider}:${method.profileId || 'legacy'}`
}

function chooseProfile(key: string | undefined): void {
  if (!key) return
  if (key === 'coupon') {
    emit('chooseMethod', 'coupon')
    couponError.value = null
    return
  }
  const first = externalMethods.value.find((method) => profileKey(method) === key && method.available)
    ?? externalMethods.value.find((method) => profileKey(method) === key)
  if (first) {
    emit('chooseMethod', first.id)
    couponError.value = null
  }
}

function methodNote(method: FeaturePaymentMethod): string {
  return method.available ? '' : t('payment.rateUnavailable')
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
  <UFormField :label="$t('payment.provider')">
    <USelectMenu v-model="selectedProfileKey" :items="profileItems" :placeholder="$t('payment.chooseProvider')" value-key="value" :search-input="false" />
  </UFormField>
  <UAlert v-if="selectedProvider !== 'coupon' && !externalMethods.some((method) => method.available)" color="warning" variant="soft" icon="i-ph-warning-circle" :description="$t('payment.noChannel')" />
  <UFormField v-if="selectedProvider !== 'coupon' && channels.length" :label="$t('payment.channel')">
    <URadioGroup
      :model-value="selectedMethodId ?? undefined"
      :items="channelItems"
      orientation="vertical"
      variant="card"
      @update:model-value="emit('chooseMethod', String($event))"
    />
  </UFormField>
  <div v-if="selectedProvider === 'coupon'" class="coupon-redemption">
    <div class="coupon-redemption__form">
      <UInput v-model="couponCode" :placeholder="$t('payment.couponCode')" :aria-label="$t('payment.couponCode')" />
      <UButton block :disabled="!couponCode.trim() || couponBusy" :loading="couponBusy" :label="$t('payment.redeemCoupon')" data-haptic @click="redeemCoupon" />
    </div>
    <UAlert v-if="couponError" color="error" variant="soft" :description="couponError" />
  </div>
  <UAlert v-if="error && selectedProvider !== 'coupon'" color="error" variant="soft" :description="error" />
  <UButton v-if="selectedProvider !== 'coupon'" block :disabled="!canCreate || stage === 'creating' || !amountValid" :loading="stage === 'creating'" :label="stage === 'creating' ? $t('payment.creating') : canReissue ? $t('payment.reissue') : $t('payment.continue')" data-haptic @click="emit('createOrder')" />
</template>
