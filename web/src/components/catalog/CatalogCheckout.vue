<script setup lang="ts">
import {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui'
import { PhArrowRight, PhCheck, PhInfo, PhWallet, PhX } from '@phosphor-icons/vue'
import { RouterLink } from 'vue-router'

import type { CouponGrant } from '@/api/features'
import type { Combo, Money, PurchaseQuote, SquadProduct } from '@/api/types'
import { formatMoney } from '@/utils/format'

const open = defineModel<boolean>('open', { required: true })
const couponGrantId = defineModel<string | null>('couponGrantId', { required: true })

defineProps<{
  combo?: Combo
  squads: readonly SquadProduct[]
  coupons: readonly CouponGrant[]
  balance: Money
  quote: PurchaseQuote | null
  quoting: boolean
  purchasing: boolean
  needsBalance: boolean
  error?: string | null
}>()

defineEmits<{ confirm: [] }>()
</script>

<template>
  <DialogRoot v-model:open="open">
    <DialogPortal to="#overlays">
      <DialogOverlay class="dialog-overlay" />
      <DialogContent class="dialog-content checkout-sheet">
        <div class="dialog-handle" aria-hidden="true" />
        <header class="dialog-header">
          <div>
            <DialogTitle class="dialog-title">{{ $t('catalog.reviewPurchase') }}</DialogTitle>
            <DialogDescription class="dialog-description">{{ $t('catalog.checkoutDescription') }}</DialogDescription>
          </div>
          <DialogClose class="icon-button" :aria-label="$t('catalog.closeCheckout')"><PhX :size="20" /></DialogClose>
        </header>

        <div v-if="combo" class="checkout-summary">
          <div class="checkout-summary__hero">
            <span class="feature-icon"><PhCheck :size="21" weight="bold" /></span>
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
          <label v-if="coupons.length" class="checkout-coupon">
            <span>{{ $t('catalog.coupon') }}</span>
            <select v-model="couponGrantId" class="compact-select">
              <option :value="null">{{ $t('catalog.noCoupon') }}</option>
              <option v-for="grant in coupons" :key="grant.id" :value="grant.id">{{ grant.coupon.name }} · {{ grant.coupon.code }}</option>
            </select>
            <small>{{ $t('catalog.couponHint') }}</small>
          </label>
          <div class="checkout-balance">
            <span><PhWallet :size="18" /> {{ $t('catalog.currentBalance') }}</span>
            <strong>{{ formatMoney(balance) }}</strong>
          </div>
          <div v-if="quote" class="checkout-quote" role="status">
            <span><PhInfo :size="18" /> {{ $t('catalog.effectiveDate') }}</span>
            <strong>{{ new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(quote.effectiveAt)) }}</strong>
            <small>{{ quote.queued ? $t('catalog.queuedEffectiveHint') : $t('catalog.immediateEffectiveHint') }}</small>
          </div>
          <div v-if="quote" class="checkout-total">
            <span>{{ $t('catalog.serverTotal') }}</span>
            <strong>{{ formatMoney(quote.netPrice) }}</strong>
          </div>
        </div>

        <div v-if="error" class="notice notice--warning" role="alert">
          <PhInfo :size="20" weight="fill" />
          <p>{{ error }}</p>
        </div>

        <RouterLink v-if="needsBalance" class="button button--primary button--wide" to="/balance">
          {{ $t('catalog.addBalance') }}
          <PhArrowRight :size="19" />
        </RouterLink>
        <button v-else class="button button--primary button--wide" type="button" :disabled="purchasing || quoting || !combo || !quote" @click="$emit('confirm')">
          {{ quoting ? $t('catalog.quoting') : purchasing ? $t('catalog.confirming') : $t('catalog.confirmPurchase') }}
          <PhArrowRight :size="19" />
        </button>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<style scoped>
.checkout-coupon {
  display: grid;
  gap: 0.4rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--line);
}

.checkout-coupon > span {
  color: var(--text-muted);
  font-size: 0.75rem;
  font-weight: 700;
}

.checkout-coupon small {
  color: var(--text-faint);
  font-size: 0.66rem;
  line-height: 1.4;
}

.checkout-quote, .checkout-total { display: grid; gap: 0.3rem; padding-top: 0.75rem; border-top: 1px solid var(--line); }
.checkout-quote > span { display: flex; align-items: center; gap: 0.4rem; color: var(--warning); font-size: 0.72rem; font-weight: 700; }
.checkout-quote strong { font-size: 0.86rem; }
.checkout-quote small { color: var(--text-faint); font-size: 0.66rem; line-height: 1.45; }
.checkout-total { grid-template-columns: 1fr auto; align-items: baseline; }
.checkout-total span { color: var(--text-muted); font-size: 0.75rem; }
.checkout-total strong { font-size: 1.05rem; }
</style>
