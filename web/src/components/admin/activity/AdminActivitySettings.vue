<script setup lang="ts">
import { reactive, watch } from 'vue'

import type { ActivitySettings } from '@/api/features'
import { useI18n } from '@/i18n'
const props = defineProps<{ settings: ActivitySettings | null; busy: boolean }>()
const emit = defineEmits<{ save: [value: { timezone: string; groupMessageThreshold: number }] }>()
const draft = reactive({ timezone: 'Asia/Shanghai', groupMessageThreshold: 0 })
const { t } = useI18n()

watch(() => props.settings, (settings) => {
  if (!settings) return
  draft.timezone = settings.timezone
  draft.groupMessageThreshold = settings.groupMessageThreshold
}, { immediate: true })
</script>

<template>
  <form class="activity-settings" @submit.prevent="emit('save', { ...draft })">
    <div>
      <h3>{{ t('adminActivity.dailyCheckIn') }}</h3>
      <p>{{ t('adminActivity.dailyCheckInHint') }}</p>
    </div>
    <UFormField name="activity-timezone" :label="t('adminActivity.calendarTimezone')" :hint="t('adminActivity.timezoneHint')" required><UInput v-model.trim="draft.timezone" class="w-full" :maxlength="80" /></UFormField>
    <UFormField name="group-threshold" :label="t('adminActivity.groupMessageThreshold')" :hint="t('adminActivity.groupMessageThresholdHint')"><UInput v-model.number="draft.groupMessageThreshold" class="w-full" type="number" :min="0" :step="1" inputmode="numeric" /></UFormField>
    <UButton class="activity-settings__save" type="submit" color="neutral" variant="outline" icon="i-ph-floppy-disk" :loading="busy" :disabled="busy" :label="busy ? t('common.saving') : t('adminActivity.saveSettings')" />
  </form>
</template>

<style scoped>
.activity-settings { display: grid; gap: 0.8rem; margin: 0 1rem 1rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.activity-settings h3, .activity-settings p { margin: 0; }
.activity-settings p { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.8rem; }
.activity-settings__save { justify-self: start; }
@media (min-width: 760px) { .activity-settings { grid-template-columns: minmax(180px, 0.8fr) repeat(2, minmax(190px, 1fr)); align-items: end; } }
</style>
