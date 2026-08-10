<script setup lang="ts">
import { reactive, watch } from 'vue'

import type { ActivitySettings } from '@/api/features'
const props = defineProps<{ settings: ActivitySettings | null; busy: boolean }>()
const emit = defineEmits<{ save: [value: { timezone: string; groupMessageThreshold: number }] }>()
const draft = reactive({ timezone: 'Asia/Shanghai', groupMessageThreshold: 0 })

watch(() => props.settings, (settings) => {
  if (!settings) return
  draft.timezone = settings.timezone
  draft.groupMessageThreshold = settings.groupMessageThreshold
}, { immediate: true })
</script>

<template>
  <form class="activity-settings" @submit.prevent="emit('save', { ...draft })">
    <div>
      <h3>{{ $t('adminActivity.dailyCheckIn') }}</h3>
      <p>{{ $t('adminActivity.dailyCheckInHint') }}</p>
    </div>
    <label>
      <span>{{ $t('adminActivity.calendarTimezone') }}</span>
      <input v-model.trim="draft.timezone" required maxlength="80" placeholder="Asia/Shanghai" />
      <small class="field-hint">{{ $t('adminActivity.timezoneHint') }}</small>
    </label>
    <label><span>{{ $t('adminActivity.groupMessageThreshold') }}</span><input v-model.number="draft.groupMessageThreshold" type="number" min="0" step="1" inputmode="numeric" /><small class="field-hint">{{ $t('adminActivity.groupMessageThresholdHint') }}</small></label>
    <button class="button button--secondary" type="submit" :disabled="busy">{{ busy ? $t('common.saving') : $t('adminActivity.saveSettings') }}</button>
  </form>
</template>

<style scoped>
.activity-settings { display: grid; gap: 0.8rem; margin: 0 1rem 1rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.activity-settings h3, .activity-settings p { margin: 0; }
.activity-settings p { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.8rem; }
.activity-settings .button { justify-self: start; }
@media (min-width: 760px) { .activity-settings { grid-template-columns: minmax(180px, 0.8fr) repeat(2, minmax(190px, 1fr)); align-items: end; } }
</style>
