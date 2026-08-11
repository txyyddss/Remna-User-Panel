<script setup lang="ts">
import type { DeepReadonly } from 'vue'

import type { Combo, PurchaseQuote, SquadProduct } from '@/api/types'
import type { CouponGrant } from '@/api/features'
import { formatMoney } from '@/utils/format'

defineProps<{
  combo?: Combo
  squads: readonly SquadProduct[]
  coupon: CouponGrant | null
  quote: DeepReadonly<PurchaseQuote> | null
  quoting: boolean
  error?: string | null
}>()
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
        <strong>{{ squads.map((squad) => squad.name).join(', ') }}</strong>
      </div>
      <div class="catalog-checkout__line">
        <span>{{ $t('catalog.coupon') }}</span>
        <strong>{{ coupon?.coupon.name ?? $t('catalog.noCoupon') }}</strong>
      </div>
      <div class="catalog-checkout__total">
        <span>{{ $t('catalog.serverTotal') }}</span>
        <strong>{{ quote ? formatMoney(quote.netPrice) : $t('catalog.quoting') }}</strong>
        <small v-if="quote">{{ quote.queued ? $t('catalog.queuedEffectiveHint') : $t('catalog.immediateEffectiveHint') }}</small>
      </div>
    </div>
    <UAlert v-if="error" color="warning" variant="soft" icon="i-ph-warning-circle" :description="error" />
    <USkeleton v-if="quoting" class="h-16" />
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
</style>
