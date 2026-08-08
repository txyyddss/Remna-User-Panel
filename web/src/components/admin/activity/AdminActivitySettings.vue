<script setup lang="ts">
import { reactive, watch } from 'vue'

import type { ActivitySettings } from '@/api/features'
import TxbAmountField from '@/components/common/TxbAmountField.vue'

const props = defineProps<{ settings: ActivitySettings | null; busy: boolean }>()
const emit = defineEmits<{ save: [value: { timezone: string; dailyRewardTxb: string; groupMessageThreshold: number; groupMessageRewardTxb: string }] }>()
const draft = reactive({ timezone: 'Asia/Shanghai', dailyRewardTxb: '0.00', groupMessageThreshold: 0, groupMessageRewardTxb: '0.00' })

watch(() => props.settings, (settings) => {
  if (!settings) return
  draft.timezone = settings.timezone
  draft.dailyRewardTxb = settings.dailyRewardTxb
  draft.groupMessageThreshold = settings.groupMessageThreshold
  draft.groupMessageRewardTxb = settings.groupMessageRewardTxb
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
    <TxbAmountField id="daily-check-in-reward" v-model="draft.dailyRewardTxb" :label="$t('adminActivity.dailyReward')" min-minor="0" required />
    <label><span>{{ $t('adminActivity.groupMessageThreshold') }}</span><input v-model.number="draft.groupMessageThreshold" type="number" min="0" step="1" inputmode="numeric" /><small class="field-hint">{{ $t('adminActivity.groupMessageThresholdHint') }}</small></label>
    <TxbAmountField id="group-message-reward" v-model="draft.groupMessageRewardTxb" :label="$t('adminActivity.groupMessageReward')" min-minor="0" required />
    <button class="button button--secondary" type="submit" :disabled="busy">{{ busy ? $t('common.saving') : $t('adminActivity.saveSettings') }}</button>
  </form>
</template>

<style scoped>
.activity-settings { display: grid; gap: 0.8rem; margin: 0 1rem 1rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.activity-settings h3, .activity-settings p { margin: 0; }
.activity-settings p { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.8rem; }
.activity-settings .button { justify-self: start; }
@media (min-width: 760px) { .activity-settings { grid-template-columns: minmax(180px, 0.8fr) repeat(2, minmax(190px, 1fr)) auto; align-items: end; } }
</style>
