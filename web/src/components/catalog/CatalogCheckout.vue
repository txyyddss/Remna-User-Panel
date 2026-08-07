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

import type { Combo, Money, SquadProduct } from '@/api/types'
import { formatMoney } from '@/utils/format'

const open = defineModel<boolean>('open', { required: true })

defineProps<{
  combo?: Combo
  squads: readonly SquadProduct[]
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
