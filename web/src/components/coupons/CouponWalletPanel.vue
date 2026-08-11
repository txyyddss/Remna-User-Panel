<script setup lang="ts">
import { computed, shallowRef } from 'vue'

import type { CouponGrant } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useCoupons } from '@/composables/useCoupons'
import CouponGrantList from './CouponGrantList.vue'

const { grants, loading, redeeming, discarding, error, message, redeem, discard } = useCoupons()
const code = shallowRef('')
const selectedGrant = shallowRef<CouponGrant | null>(null)
const discardingGrantId = computed(() => discarding.value ? selectedGrant.value?.id ?? null : null)

async function submit(): Promise<void> {
  if (await redeem(code.value)) code.value = ''
}

function setDiscardOpen(open: boolean): void {
  if (!open) selectedGrant.value = null
}

async function confirmDiscard(): Promise<void> {
  if (!selectedGrant.value) return
  if (await discard(selectedGrant.value.id)) selectedGrant.value = null
}
</script>

<template>
  <section class="section-block coupon-wallet">
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('coupons.wallet') }}</h2>
      <p>{{ $t('coupons.walletHint') }}</p>
    </div>
    <form class="coupon-redeem" @submit.prevent="submit">
      <span id="coupon-code-label" class="sr-only">{{ $t('coupons.code') }}</span>
      <UInput
        id="coupon-code"
        v-model.trim="code"
        icon="i-ph-ticket"
        :placeholder="$t('coupons.codePlaceholder')"
        autocomplete="off"
        aria-labelledby="coupon-code-label"
      />
      <UButton
        type="submit"
        color="neutral"
        variant="outline"
        trailing-icon="i-ph-arrow-right"
        :disabled="redeeming || !code"
        :loading="redeeming"
        :label="redeeming ? $t('coupons.redeeming') : $t('coupons.redeem')"
      />
    </form>
    <InlineNotice v-if="message" tone="success">{{ message }}</InlineNotice>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <USkeleton v-if="loading" class="h-10" />
    <CouponGrantList v-else-if="grants.length" :grants="grants" :discarding-id="discardingGrantId" @discard="selectedGrant = $event" />
    <div v-else class="empty-inline"><div><h3>{{ $t('coupons.empty') }}</h3><p>{{ $t('coupons.emptyHint') }}</p></div></div>

    <UModal
      :open="selectedGrant !== null"
      :title="$t('coupons.discardTitle', { name: selectedGrant?.coupon.name ?? '' })"
      :description="$t('coupons.discardDescription')"
      :dismissible="!discarding"
      :ui="{ footer: 'justify-end' }"
      @update:open="setDiscardOpen"
    >
      <template #body>
        <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      </template>
      <template #footer="{ close }">
        <UButton color="neutral" variant="outline" :label="$t('common.cancel')" :disabled="discarding" @click="close" />
        <UButton color="error" icon="i-ph-trash" :loading="discarding" :disabled="discarding" :label="$t('coupons.discardConfirm')" @click="confirmDiscard" />
      </template>
    </UModal>
  </section>
</template>

<style scoped>
.coupon-wallet { display: grid; gap: 0.8rem; }
.coupon-redeem { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 0.55rem; }
</style>
