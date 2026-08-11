<script setup lang="ts">
import type { DeepReadonly } from 'vue'
import { useRouter } from 'vue-router'

import type { Combo, Purchase, PurchaseQuote } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { formatMoney } from '@/utils/format'

const router = useRouter()

defineProps<{
  combo?: Combo
  quote: DeepReadonly<PurchaseQuote> | null
  purchase: DeepReadonly<Purchase> | null
  purchasing: boolean
  needsBalance: boolean
  error?: string | null
}>()

defineEmits<{ confirm: [] }>()

function goHome(): void {
  void router.push('/home')
}

function goToBalance(): void {
  void router.push({ path: '/home', query: { topUp: '1' } })
}
</script>

<template>
  <section class="catalog-payment-step">
    <template v-if="purchase">
      <InlineNotice tone="success" :title="$t('catalog.purchaseConfirmed')">{{ $t('catalog.purchaseScheduled', { name: purchase.comboName }) }}</InlineNotice>
      <UButton block trailing-icon="i-ph-house" :label="$t('catalog.returnHome')" data-haptic @click="goHome" />
    </template>
    <template v-else>
      <div class="section-heading section-heading--stacked">
        <h2>{{ $t('catalog.steps.payment') }}</h2>
        <p>{{ $t('catalog.paymentHint') }}</p>
      </div>
      <div class="catalog-payment-total">
        <span>{{ $t('catalog.serverTotal') }}</span>
        <strong>{{ quote ? formatMoney(quote.netPrice) : $t('common.notAvailable') }}</strong>
      </div>
      <UAlert v-if="error" color="warning" variant="soft" icon="i-ph-warning-circle" :description="error" />
      <UButton v-if="needsBalance" block trailing-icon="i-ph-plus" :label="$t('catalog.addBalance')" data-haptic @click="goToBalance" />
      <UButton v-else block :disabled="purchasing || !quote || !combo" :loading="purchasing" trailing-icon="i-ph-check" :label="purchasing ? $t('catalog.confirming') : $t('catalog.confirmPurchase')" data-haptic @click="$emit('confirm')" />
    </template>
  </section>
</template>

<style scoped>
.catalog-payment-step { display: grid; gap: 0.85rem; }
.catalog-payment-total { display: grid; gap: 0.25rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.catalog-payment-total > span, .catalog-payment-total small { color: var(--text-faint); font-size: 0.72rem; }
.catalog-payment-total strong { color: var(--accent); font-family: var(--font-mono); font-size: 1.65rem; letter-spacing: -0.04em; }
</style>
