<script setup lang="ts">
import { computed } from 'vue'

import type { LedgerEntry } from '@/api/types'
import { useI18n } from '@/i18n'
import { formatDateTime, formatMoney } from '@/utils/format'

const props = defineProps<{ entries: readonly LedgerEntry[] }>()
const { t } = useI18n()

const contextualKinds = new Set([
  'admin_adjustment', 'purchase_debit', 'activity_bet_stake', 'activity_bet_payout',
  'activity_draw_fee', 'activity_draw_reward', 'coupon_balance_add',
  'coupon_balance_multiply', 'admin_entitlement_cancellation', 'payment_reversal',
])

function isCredit(entry: LedgerEntry): boolean {
  return !entry.delta.minor.startsWith('-') && entry.delta.minor !== '0'
}

const visibleEntries = computed(() => props.entries.slice(0, 12))

function entryLabel(entry: LedgerEntry): string {
  const key = `billing.ledgerKind.${entry.kind}`
  const label = t(key)
  return label === key ? t('billing.transaction') : label
}

function entryContext(entry: LedgerEntry): string {
  return contextualKinds.has(entry.kind) ? entry.note : ''
}
</script>

<template>
  <section class="ledger-section">
    <div class="section-heading">
      <h2>{{ $t('billing.recentActivity') }}</h2>
      <span class="section-heading__meta">{{ $t('billing.entryCount', { count: entries.length }) }}</span>
    </div>
    <div v-if="visibleEntries.length" v-auto-animate class="ledger-list">
      <article v-for="entry in visibleEntries" :key="entry.id" class="ledger-row">
        <span class="ledger-row__icon" :class="{ 'ledger-row__icon--credit': isCredit(entry) }">
          <UIcon :name="isCredit(entry) ? 'i-ph-arrow-down-left' : 'i-ph-arrow-up-right'" />
        </span>
        <span class="ledger-row__copy">
          <strong>{{ entryLabel(entry) }}</strong>
          <small>{{ entryContext(entry) ? `${entryContext(entry)} · ` : '' }}{{ formatDateTime(entry.createdAt) }}</small>
        </span>
        <strong class="ledger-row__amount" :class="{ 'ledger-row__amount--credit': isCredit(entry) }">
          {{ isCredit(entry) ? '+' : '' }}{{ formatMoney(entry.delta) }}
        </strong>
      </article>
    </div>
    <div v-else class="empty-inline">
      <span class="feature-icon"><UIcon name="i-ph-receipt" /></span>
      <div>
        <h3>{{ $t('billing.noActivity') }}</h3>
        <p>{{ $t('billing.noActivityHint') }}</p>
      </div>
    </div>
  </section>
</template>
