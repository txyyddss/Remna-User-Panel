<script setup lang="ts">
import type { NodeCompensationEvent } from '@/api/contracts/compensation'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'
import { durationParts, multiplierFactor } from './format'

const props = defineProps<{ event: NodeCompensationEvent }>()
defineEmits<{ review: [event: NodeCompensationEvent] }>()
const { t } = useI18n()

function durationText(): string {
  if (props.event.observedDurationSeconds === null) return t('adminCompensation.ongoing')
  const value = durationParts(props.event.observedDurationSeconds)
  return t('adminCompensation.duration', value)
}
</script>

<template>
  <article class="compensation-event">
    <div class="compensation-event__top">
      <div><p class="eyebrow">{{ t(`adminCompensation.status.${event.status}`) }}</p><h3>{{ event.nodeName }}</h3></div>
      <UBadge color="neutral" variant="soft" :label="durationText()" />
    </div>
    <p class="compensation-event__time">{{ formatDateTime(event.offlineObservedAt) }}</p>
    <div class="compensation-event__facts">
      <span><small>{{ t('adminCompensation.thresholdShort') }}</small><strong>{{ event.thresholdMinutes }}m</strong></span>
      <span><small>{{ t('adminCompensation.multiplierShort') }}</small><strong>{{ multiplierFactor(event.multiplierBps) }}×</strong></span>
      <span><small>{{ t('adminCompensation.recipientsShort') }}</small><strong>{{ event.frozenRecipientCount }}</strong></span>
    </div>
    <div v-if="event.squads.length" class="compensation-event__squads">
      <UBadge v-for="squad in event.squads" :key="squad.uuid" color="neutral" variant="outline" :label="squad.name" />
    </div>
    <p v-if="event.ineligibleReason" class="compensation-event__note">{{ t(`adminCompensation.reason.${event.ineligibleReason}`) }}</p>
    <p v-else-if="event.operation" class="compensation-event__note">{{ t('adminCompensation.operation', { status: event.operation.status }) }}</p>
    <UButton v-if="event.status === 'pending_review'" block color="warning" variant="soft" icon="i-ph-magnifying-glass" :label="t('adminCompensation.review')" data-haptic="open" @click="$emit('review', event)" />
  </article>
</template>

<style scoped>
.compensation-event { display: grid; gap: 0.8rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.compensation-event h3, .compensation-event p { margin: 0; }
.compensation-event__top { display: flex; align-items: flex-start; justify-content: space-between; gap: 0.75rem; }
.compensation-event__time, .compensation-event__note { color: var(--text-muted); font-size: 0.76rem; line-height: 1.5; }
.compensation-event__facts { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1px; overflow: hidden; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--line); }
.compensation-event__facts span { min-width: 0; padding: 0.7rem; background: var(--surface); }
.compensation-event__facts small, .compensation-event__facts strong { display: block; }
.compensation-event__facts small { color: var(--text-faint); font-size: 0.65rem; }
.compensation-event__facts strong { margin-top: 0.15rem; font-size: 0.88rem; }
.compensation-event__squads { display: flex; flex-wrap: wrap; gap: 0.4rem; }
</style>
