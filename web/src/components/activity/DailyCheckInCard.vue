<script setup lang="ts">
import { txbInputFromMinor } from '@/utils/format'

defineProps<{
  checkedIn: boolean
  rewardMinTxbMinor: string
  rewardMaxTxbMinor: string
  timeZone: string
  busy: boolean
}>()

defineEmits<{ checkIn: [] }>()
</script>

<template>
  <section class="section-block activity-card">
    <span class="feature-icon"><UIcon name="i-ph-calendar-check" /></span>
    <div class="activity-card__copy">
      <h2>{{ $t('activity.dailyCheckIn') }}</h2>
      <p>
        {{ checkedIn ? $t('activity.todayComplete') : rewardMinTxbMinor === rewardMaxTxbMinor
          ? $t('activity.claimToday', { amount: txbInputFromMinor(rewardMinTxbMinor) })
          : $t('activity.claimRangeToday', { minimum: txbInputFromMinor(rewardMinTxbMinor), maximum: txbInputFromMinor(rewardMaxTxbMinor) }) }}
      </p>
      <small>{{ $t('activity.resetsAt', { timezone: timeZone }) }}</small>
    </div>
    <UButton
      icon="i-ph-check-circle-fill"
      :disabled="checkedIn || busy"
      :loading="busy"
      :label="checkedIn ? $t('activity.checkedIn') : busy ? $t('activity.checkingIn') : $t('activity.checkIn')"
      @click="$emit('checkIn')"
    />
  </section>
</template>

<style scoped>
.activity-card { display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: start; gap: 0.8rem; }
.activity-card__copy h2, .activity-card__copy p { margin: 0; }
.activity-card__copy h2 { font-size: 1.05rem; }
.activity-card__copy p { margin-top: 0.35rem; color: var(--text-muted); font-size: 0.82rem; }
.activity-card__copy small { display: block; margin-top: 0.45rem; color: var(--text-faint); font-size: 0.68rem; }
.activity-card :deep(button) { grid-column: 1 / -1; width: 100%; justify-content: center; text-align: center; }
@media (min-width: 640px) { .activity-card { grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; } .activity-card :deep(button) { grid-column: auto; width: auto; } }
</style>
