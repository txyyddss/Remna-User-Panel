<script setup lang="ts">
import { computed } from 'vue'
import { PhArrowDownLeft, PhArrowUpRight, PhReceipt } from '@phosphor-icons/vue'

import type { LedgerEntry } from '@/api/types'
import { formatDateTime, formatMoney } from '@/utils/format'

const props = defineProps<{ entries: readonly LedgerEntry[] }>()

function isCredit(entry: LedgerEntry): boolean {
  return !entry.delta.minor.startsWith('-') && entry.delta.minor !== '0'
}

const visibleEntries = computed(() => props.entries.slice(0, 12))
</script>

<template>
  <section class="ledger-section">
    <div class="section-heading">
      <h2>Recent activity</h2>
      <span class="section-heading__meta">{{ entries.length }} entries</span>
    </div>
    <div v-if="visibleEntries.length" class="ledger-list">
      <article v-for="entry in visibleEntries" :key="entry.id" class="ledger-row">
        <span class="ledger-row__icon" :class="{ 'ledger-row__icon--credit': isCredit(entry) }">
          <PhArrowDownLeft v-if="isCredit(entry)" :size="19" />
          <PhArrowUpRight v-else :size="19" />
        </span>
        <span class="ledger-row__copy">
          <strong>{{ entry.note }}</strong>
          <small>{{ formatDateTime(entry.createdAt) }}</small>
        </span>
        <strong class="ledger-row__amount" :class="{ 'ledger-row__amount--credit': isCredit(entry) }">
          {{ isCredit(entry) ? '+' : '' }}{{ formatMoney(entry.delta) }}
        </strong>
      </article>
    </div>
    <div v-else class="empty-inline">
      <span class="feature-icon"><PhReceipt :size="22" /></span>
      <div>
        <h3>No balance activity</h3>
        <p>Your top-ups and purchases will appear here.</p>
      </div>
    </div>
  </section>
</template>
