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
import type { Combo, Money, SquadProduct } from '@/api/types'
import { formatMoney } from '@/utils/format'

const open = defineModel<boolean>('open', { required: true })
const couponGrantId = defineModel<string | null>('couponGrantId', { required: true })

defineProps<{
  combo?: Combo
  squads: readonly SquadProduct[]
  coupons: readonly CouponGrant[]
  balance: Money
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
            <DialogTitle class="dialog-title">Review purchase</DialogTitle>
            <DialogDescription class="dialog-description">The server confirms the final TXB total before deduction.</DialogDescription>
          </div>
          <DialogClose class="icon-button" aria-label="Close checkout"><PhX :size="20" /></DialogClose>
        </header>

        <div v-if="combo" class="checkout-summary">
          <div class="checkout-summary__hero">
            <span class="feature-icon"><PhCheck :size="21" weight="bold" /></span>
            <span>
              <strong>{{ combo.name }}</strong>
              <small>{{ combo.validityDays }} days, {{ combo.resetStrategy.toLowerCase() }} traffic reset</small>
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
            <span>Coupon</span>
            <select v-model="couponGrantId" class="compact-select">
              <option :value="null">No coupon</option>
              <option v-for="grant in coupons" :key="grant.id" :value="grant.id">{{ grant.coupon.name }} · {{ grant.coupon.code }}</option>
            </select>
            <small>One eligible discount can be applied. The server confirms the final total.</small>
          </label>
          <div class="checkout-balance">
            <span><PhWallet :size="18" /> Current balance</span>
            <strong>{{ formatMoney(balance) }}</strong>
          </div>
        </div>

        <div v-if="error" class="notice notice--warning" role="alert">
          <PhInfo :size="20" weight="fill" />
          <p>{{ error }}</p>
        </div>

        <RouterLink v-if="needsBalance" class="button button--primary button--wide" to="/balance">
          Add balance
          <PhArrowRight :size="19" />
        </RouterLink>
        <button v-else class="button button--primary button--wide" type="button" :disabled="purchasing || !combo" @click="$emit('confirm')">
          {{ purchasing ? 'Confirming with server' : 'Confirm purchase' }}
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
</style>
