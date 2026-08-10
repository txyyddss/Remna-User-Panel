<script setup lang="ts">
import { computed } from 'vue'
import type { DeepReadonly } from 'vue'

import type { CouponGrant } from '@/api/features'
import type { Combo, Money, PurchaseQuote, SquadProduct } from '@/api/types'
import { useI18n } from '@/i18n'
import { formatMoney } from '@/utils/format'

const open = defineModel<boolean>('open', { required: true })
const couponGrantId = defineModel<string | null>('couponGrantId', { required: true })

const props = defineProps<{
  combo?: Combo
  squads: readonly SquadProduct[]
  coupons: readonly CouponGrant[]
  balance: Money
  quote: DeepReadonly<PurchaseQuote> | null
  quoting: boolean
  purchasing: boolean
  needsBalance: boolean
  error?: string | null
}>()

defineEmits<{ confirm: [] }>()

const { t } = useI18n()
const couponItems = computed(() => [
  { label: t('catalog.noCoupon'), value: null },
  ...props.coupons.map((grant) => ({
    label: `${grant.coupon.name} · ${grant.coupon.code}`,
    value: grant.id,
  })),
])
</script>

<template>
  <UModal
    v-model:open="open"
    :title="$t('catalog.reviewPurchase')"
    :description="$t('catalog.checkoutDescription')"
  >
    <template #body>
      <div v-if="combo" class="checkout-summary">
        <div class="checkout-summary__hero">
          <span class="feature-icon"><UIcon name="i-ph-check-bold" /></span>
          <span>
            <strong>{{ combo.name }}</strong>
            <small>{{ $t('catalog.termSummary', { days: combo.validityDays, reset: combo.resetStrategy.toLowerCase() }) }}</small>
          </span>
          <strong>{{ formatMoney(combo.price) }}</strong>
        </div>
        <div v-if="squads.length" class="checkout-lines">
          <div v-for="squad in squads" :key="squad.id">
            <span>{{ squad.name }}</span>
            <strong>{{ formatMoney(squad.price) }}</strong>
          </div>
        </div>
        <UFormField
          v-if="coupons.length"
          class="checkout-coupon"
          :label="$t('catalog.coupon')"
          :description="$t('catalog.couponHint')"
        >
          <USelect v-model="couponGrantId" :items="couponItems" value-key="value" />
        </UFormField>
        <div class="checkout-balance">
          <span><UIcon name="i-ph-wallet" /> {{ $t('catalog.currentBalance') }}</span>
          <strong>{{ formatMoney(balance) }}</strong>
        </div>
        <div v-if="quote" class="checkout-quote" role="status">
          <span><UIcon name="i-ph-info" /> {{ $t('catalog.effectiveDate') }}</span>
          <strong>{{ new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(quote.effectiveAt)) }}</strong>
          <small>{{ quote.queued ? $t('catalog.queuedEffectiveHint') : $t('catalog.immediateEffectiveHint') }}</small>
        </div>
        <div v-if="quote" class="checkout-total">
          <span>{{ $t('catalog.serverTotal') }}</span>
          <strong>{{ formatMoney(quote.netPrice) }}</strong>
        </div>
      </div>

      <UAlert v-if="error" color="warning" variant="soft" icon="i-ph-info-fill" :description="error" />

      <UButton
        v-if="needsBalance"
        block
        to="/balance"
        trailing-icon="i-ph-arrow-right"
        :label="$t('catalog.addBalance')"
      />
      <UButton
        v-else
        block
        :disabled="purchasing || quoting || !combo || !quote"
        :loading="purchasing || quoting"
        trailing-icon="i-ph-arrow-right"
        :label="quoting ? $t('catalog.quoting') : purchasing ? $t('catalog.confirming') : $t('catalog.confirmPurchase')"
        @click="$emit('confirm')"
      />
    </template>
  </UModal>
</template>

<style scoped>
.checkout-coupon { padding-top: 0.75rem; border-top: 1px solid var(--line); }
.checkout-quote, .checkout-total { display: grid; gap: 0.3rem; padding-top: 0.75rem; border-top: 1px solid var(--line); }
.checkout-quote > span { display: flex; align-items: center; gap: 0.4rem; color: var(--warning); font-size: 0.72rem; font-weight: 700; }
.checkout-quote strong { font-size: 0.86rem; }
.checkout-quote small { color: var(--text-faint); font-size: 0.66rem; line-height: 1.45; }
.checkout-total { grid-template-columns: 1fr auto; align-items: baseline; }
.checkout-total span { color: var(--text-muted); font-size: 0.75rem; }
.checkout-total strong { font-size: 1.05rem; }
</style>
