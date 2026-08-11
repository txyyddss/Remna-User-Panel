<script setup lang="ts">
import { shallowRef, watch } from 'vue'

import type { FeaturePaymentMethod, FeaturePaymentOrder } from '@/api/features'
import { api } from '@/api/client'
import type { Money } from '@/api/types'
import BalancePaymentSheet from '@/components/billing/BalancePaymentSheet.vue'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { localizedError } from '@/i18n'
import { formatMoney } from '@/utils/format'

const props = defineProps<{
  balance: Money
  openTopUp?: boolean
  reissueOrderId?: string
}>()

const emit = defineEmits<{
  paid: []
  topUpRequestConsumed: []
  reissueRequestConsumed: []
}>()

const paymentMethods = shallowRef<FeaturePaymentMethod[]>([])
const reissueOrder = shallowRef<FeaturePaymentOrder | null>(null)
const paymentOpen = shallowRef(false)
const paymentLoading = shallowRef(false)
const paymentError = shallowRef<string | null>(null)

function openTopUpPayment(reissueOrderId?: string): boolean {
  if (paymentLoading.value) return false
  paymentLoading.value = true
  paymentError.value = null
  reissueOrder.value = null
  void loadTopUpPayment(reissueOrderId)
  return true
}

async function loadTopUpPayment(reissueOrderId?: string): Promise<void> {
  try {
    const [response, restoredOrder] = await Promise.all([
      api.getBalance(),
      reissueOrderId ? api.getPaymentOrder(reissueOrderId) : Promise.resolve(null),
    ])
    paymentMethods.value = response.paymentMethods
    reissueOrder.value = restoredOrder
    paymentOpen.value = true
  } catch (caught) {
    paymentError.value = localizedError(caught, 'errors.balanceUnavailable')
  } finally {
    paymentLoading.value = false
  }
}

watch(() => props.openTopUp, (requested) => {
  if (!requested) return
  if (openTopUpPayment()) emit('topUpRequestConsumed')
}, { immediate: true })

watch(() => props.reissueOrderId, (orderId) => {
  if (!orderId) return
  if (openTopUpPayment(orderId)) emit('reissueRequestConsumed')
}, { immediate: true })
</script>

<template>
  <section class="home-balance">
    <div class="home-balance__copy">
      <p>{{ $t('billing.availableBalance') }}</p>
      <strong>{{ formatMoney(balance) }}</strong>
    </div>
    <div class="home-balance__action">
      <UButton
        class="button button--light home-balance__button"
        icon="i-ph-plus-bold"
        :label="$t('billing.addBalance')"
        :loading="paymentLoading"
        data-haptic
        @click="openTopUpPayment()"
      />
    </div>
    <InlineNotice v-if="paymentError" class="home-balance__notice" tone="warning">{{ paymentError }}</InlineNotice>
    <BalancePaymentSheet v-model:open="paymentOpen" :methods="paymentMethods" :reissue-order="reissueOrder" @paid="emit('paid')" />
  </section>
</template>
