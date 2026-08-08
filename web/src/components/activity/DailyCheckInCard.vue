<script setup lang="ts">
import { PhCalendarCheck, PhCheckCircle } from '@phosphor-icons/vue'

import { txbInputFromMinor } from '@/utils/format'

defineProps<{
  checkedIn: boolean
  rewardTxbMinor: string
  timeZone: string
  busy: boolean
}>()

defineEmits<{ checkIn: [] }>()
</script>

<template>
  <section class="section-block activity-card">
    <span class="feature-icon"><PhCalendarCheck :size="23" /></span>
    <div class="activity-card__copy">
      <h2>Daily check-in</h2>
      <p>{{ checkedIn ? 'Today is complete.' : `Claim ${txbInputFromMinor(rewardTxbMinor)} TXB once today.` }}</p>
      <small>Resets at midnight in {{ timeZone }}.</small>
    </div>
    <button class="button button--primary" type="button" :disabled="checkedIn || busy" @click="$emit('checkIn')">
      <PhCheckCircle :size="18" weight="fill" />
      {{ checkedIn ? 'Checked in' : busy ? 'Checking in' : 'Check in' }}
    </button>
  </section>
</template>

<style scoped>
.activity-card {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: 0.8rem;
}

.activity-card__copy h2,
.activity-card__copy p { margin: 0; }
.activity-card__copy h2 { font-size: 1.05rem; }
.activity-card__copy p { margin-top: 0.35rem; color: var(--text-muted); font-size: 0.82rem; }
.activity-card__copy small { display: block; margin-top: 0.45rem; color: var(--text-faint); font-size: 0.68rem; }
.activity-card .button { grid-column: 1 / -1; width: 100%; }

@media (min-width: 640px) {
  .activity-card { grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; }
  .activity-card .button { grid-column: auto; width: auto; }
}
</style>
