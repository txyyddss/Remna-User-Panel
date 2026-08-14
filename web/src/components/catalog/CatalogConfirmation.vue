<script setup lang="ts">
import type { DeepReadonly } from 'vue'

import type { Purchase } from '@/api/types'
import { useI18n } from '@/i18n'
import { formatBytes, formatDate, formatMoney } from '@/utils/format'

const props = defineProps<{
  purchase: DeepReadonly<Purchase>
}>()

const emit = defineEmits<{
  home: []
}>()

const { t } = useI18n()

function statusLabel(): string {
  return t(`catalog.purchaseStatus.${props.purchase.status}`)
}

function resetLabel(): string {
  return t(`home.reset.${props.purchase.resetStrategy}`)
}
</script>

<template>
  <section class="catalog-confirmation" data-test="catalog-confirmation" role="status" aria-live="polite">
    <div class="catalog-confirmation__hero">
      <div class="catalog-confirmation__icon" aria-hidden="true">
        <UIcon name="i-ph-check-circle-fill" />
      </div>
      <div class="catalog-confirmation__copy">
        <p class="catalog-confirmation__eyebrow">{{ $t('catalog.purchaseConfirmed') }}</p>
        <h1>{{ $t('catalog.purchaseSuccessTitle') }}</h1>
        <p>{{ $t('catalog.purchaseScheduled', { name: purchase.comboName }) }}</p>
      </div>
    </div>

    <div class="catalog-confirmation__summary" :aria-label="$t('catalog.purchaseSummary')">
      <div class="catalog-confirmation__line">
        <span>{{ $t('catalog.coreCombos') }}</span>
        <strong>{{ purchase.comboName }}</strong>
      </div>
      <div class="catalog-confirmation__line">
        <span>{{ $t('catalog.purchaseCharged') }}</span>
        <strong>{{ formatMoney(purchase.price) }}</strong>
      </div>
      <div class="catalog-confirmation__line">
        <span>{{ $t('catalog.purchaseDiscount') }}</span>
        <strong>{{ formatMoney(purchase.couponDiscount) }}</strong>
      </div>
      <div class="catalog-confirmation__line">
        <span>{{ $t('catalog.purchaseStarts') }}</span>
        <strong>{{ formatDate(purchase.validFrom) }}</strong>
      </div>
      <div class="catalog-confirmation__line">
        <span>{{ $t('catalog.purchaseEnds') }}</span>
        <strong>{{ formatDate(purchase.validUntil) }}</strong>
      </div>
      <div class="catalog-confirmation__line">
        <span>{{ $t('catalog.purchaseStatusLabel') }}</span>
        <strong>{{ statusLabel() }}</strong>
      </div>
      <div class="catalog-confirmation__line">
        <span>{{ $t('catalog.purchaseTraffic') }}</span>
        <strong>{{ formatBytes(purchase.trafficLimitBytes) }}</strong>
      </div>
      <div class="catalog-confirmation__line">
        <span>{{ $t('catalog.purchaseReset') }}</span>
        <strong>{{ resetLabel() }}</strong>
      </div>
    </div>

    <UButton
      block
      class="catalog-confirmation__home-action"
      :ui="{ trailingIcon: 'absolute end-3 top-1/2 -translate-y-1/2' }"
      trailing-icon="i-ph-house"
      :label="$t('catalog.returnHome')"
      data-haptic
      @click="emit('home')"
    />
  </section>
</template>

<style scoped>
.catalog-confirmation { display: grid; gap: 1rem; max-width: 42rem; margin: 0 auto; }
.catalog-confirmation__hero { display: grid; justify-items: center; gap: 0.8rem; padding: 1.5rem 1rem 1rem; text-align: center; }
.catalog-confirmation__icon { display: inline-flex; align-items: center; justify-content: center; width: 4.5rem; height: 4.5rem; border: 1px solid var(--line-strong); border-radius: 1.5rem; color: var(--accent); background: var(--accent-soft); font-size: 2.75rem; }
.catalog-confirmation__copy { display: grid; gap: 0.45rem; }
.catalog-confirmation__copy h1, .catalog-confirmation__copy p { margin: 0; }
.catalog-confirmation__copy h1 { font-size: clamp(1.45rem, 5vw, 2rem); letter-spacing: -0.03em; }
.catalog-confirmation__copy p:last-child { color: var(--text-muted); font-size: 0.82rem; line-height: 1.5; }
.catalog-confirmation__eyebrow { color: var(--accent); font-size: 0.7rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; }
.catalog-confirmation__summary { display: grid; gap: 0.65rem; padding: 0.9rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.catalog-confirmation__line { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; padding-bottom: 0.65rem; border-bottom: 1px solid var(--line); }
.catalog-confirmation__line:last-child { padding-bottom: 0; border-bottom: 0; }
.catalog-confirmation__line span { color: var(--text-faint); font-size: 0.7rem; }
.catalog-confirmation__line strong { overflow-wrap: anywhere; color: var(--text); font-family: var(--font-mono); font-size: 0.78rem; text-align: right; }
.catalog-confirmation__home-action { position: relative; justify-content: center; }
</style>
