<script setup lang="ts">
import type { DeepReadonly } from 'vue'

import type { Combo, PurchaseQuote, SquadProduct } from '@/api/types'
import type { CouponGrant } from '@/api/features'
import { formatBytes, formatDate, formatMoney } from '@/utils/format'
import { useRouter } from 'vue-router'
import { useI18n } from '@/i18n'

const props = defineProps<{
  combo?: Combo
  squads: readonly SquadProduct[]
  coupon: CouponGrant | null
  quote: DeepReadonly<PurchaseQuote> | null
  quoting: boolean
  error?: string | null
  purchasing: boolean
  needsBalance: boolean
}>()
const emit = defineEmits<{ confirm: [] }>()
const router = useRouter()
const { t } = useI18n()
function couponEffect(): string {
  if (!props.coupon) return ''
  if (props.coupon.coupon.discountMode === 'percent') {
    const key = props.coupon.coupon.kind === 'purchase_once' || props.coupon.coupon.kind === 'purchase_recurring'
      ? 'coupons.effectPurchasePercent'
      : 'coupons.effectPercent'
    return t(key, { value: (Number(props.coupon.coupon.valueMinorOrBps) / 100).toFixed(2) })
  }
  return t('coupons.effectFixed', { amount: formatMoney({ currency: 'TXB', minor: props.coupon.coupon.valueMinorOrBps, display: '' }) })
}
function goToBalance(): void { void router.push({ path: '/home', query: { topUp: '1' } }) }
</script>

<template>
  <section class="catalog-checkout">
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('catalog.steps.review') }}</h2>
      <p>{{ $t('catalog.checkoutDescription') }}</p>
    </div>
    <div v-if="combo" class="catalog-checkout__summary">
      <div class="catalog-checkout__line">
        <span>{{ $t('catalog.coreCombos') }}</span>
        <strong>{{ combo.name }}</strong>
      </div>
      <div v-if="squads.length" class="catalog-checkout__line">
        <span>{{ $t('catalog.optionalSquads') }}</span>
        <strong>{{ squads.map((squad) => squad.name).join($t('home.squadSeparator')) }}</strong>
      </div>
      <div class="catalog-checkout__line">
        <span>{{ $t('catalog.coupon') }}</span>
        <strong>{{ coupon?.coupon.name ?? $t('catalog.noCoupon') }}</strong>
        <small v-if="coupon">{{ couponEffect() }}</small>
      </div>
      <div v-if="quote" class="catalog-checkout__line">
        <span>{{ $t('catalog.validity') }}</span>
        <strong>{{ formatDate(quote.effectiveAt) }} {{ $t('common.rangeSeparator') }} {{ formatDate(quote.expiresAt) }}</strong>
      </div>
      <div v-if="quote" class="catalog-checkout__line">
        <span>{{ $t('catalog.accessibleNodes') }}</span>
        <strong>{{ quote.accessibleNodes.length }}</strong>
      </div>
      <div class="catalog-checkout__line">
        <span>{{ $t('catalog.traffic') }}</span>
        <strong>{{ formatBytes(combo.trafficLimitBytes) }}</strong>
      </div>
      <div class="catalog-checkout__line">
        <span>{{ $t('catalog.term') }}</span>
        <strong>{{ $t('catalog.termSummary', { days: combo.validityDays, reset: $t(`home.reset.${combo.resetStrategy}`) }) }}</strong>
      </div>
      <div class="catalog-checkout__line">
        <span>{{ $t('catalog.rollover') }}</span>
        <strong>{{ $t('catalog.rolloverSummary', { threshold: (combo.rolloverMinRemainingBps / 100).toFixed(2), cap: formatMoney(combo.rolloverMax) }) }}</strong>
      </div>
      <div class="catalog-checkout__total">
        <span>{{ $t('catalog.serverTotal') }}</span>
        <strong>{{ quote ? formatMoney(quote.netPrice) : quoting ? $t('catalog.quoting') : $t('common.notAvailable') }}</strong>
        <small v-if="quote">{{ quote.queued ? $t('catalog.queuedEffectiveHint') : $t('catalog.immediateEffectiveHint') }}</small>
      </div>
    </div>
    <UAlert v-if="error" color="warning" variant="soft" icon="i-ph-warning-circle" :description="error" />
    <USkeleton v-if="quoting" class="h-16" />
    <UButton v-if="needsBalance" block trailing-icon="i-ph-plus" :label="$t('catalog.addBalance')" data-haptic @click="goToBalance" />
    <UButton v-else block class="catalog-checkout__confirm" :disabled="purchasing || !quote || quote.accessibleNodes.length === 0" :loading="purchasing" :label="purchasing ? $t('catalog.confirming') : $t('catalog.confirmPurchase')" data-haptic @click="emit('confirm')" />
  </section>
</template>

<style scoped>
.catalog-checkout, .catalog-checkout__summary { display: grid; gap: 0.7rem; }
.catalog-checkout__summary { padding: 0.85rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.catalog-checkout__line { display: grid; gap: 0.22rem; padding-bottom: 0.65rem; border-bottom: 1px solid var(--line); }
.catalog-checkout__line > span, .catalog-checkout__total > span, .catalog-checkout__total small { color: var(--text-faint); font-size: 0.68rem; }
.catalog-checkout__line strong { overflow-wrap: anywhere; font-size: 0.8rem; }
.catalog-checkout__total { display: grid; gap: 0.25rem; }
.catalog-checkout__total strong { color: var(--accent); font-family: var(--font-mono); font-size: 1.3rem; }
.catalog-checkout__confirm { justify-content: center; }
</style>
