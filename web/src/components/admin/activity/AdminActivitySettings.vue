<script setup lang="ts">
import { reactive, watch } from 'vue'

import type { ActivitySettings } from '@/api/features'
import TxbAmountField from '@/components/common/TxbAmountField.vue'

const props = defineProps<{ settings: ActivitySettings | null; busy: boolean }>()
const emit = defineEmits<{ save: [value: { timezone: string; dailyRewardTxb: string }] }>()
const draft = reactive({ timezone: 'Asia/Shanghai', dailyRewardTxb: '0.00' })

watch(() => props.settings, (settings) => {
  if (!settings) return
  draft.timezone = settings.timezone
  draft.dailyRewardTxb = settings.dailyRewardTxb
}, { immediate: true })
</script>

<template>
  <form class="activity-settings" @submit.prevent="emit('save', { ...draft })">
    <div>
      <h3>Daily check-in</h3>
      <p>One reward per configured local calendar day.</p>
    </div>
    <label>
      <span>Calendar timezone</span>
      <input v-model.trim="draft.timezone" required maxlength="80" placeholder="Asia/Shanghai" />
      <small class="field-hint">Use an IANA timezone name.</small>
    </label>
    <TxbAmountField id="daily-check-in-reward" v-model="draft.dailyRewardTxb" label="Daily reward" min-minor="0" required />
    <button class="button button--secondary" type="submit" :disabled="busy">{{ busy ? 'Saving' : 'Save check-in settings' }}</button>
  </form>
</template>

<style scoped>
.activity-settings { display: grid; gap: 0.8rem; margin: 0 1rem 1rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.activity-settings h3, .activity-settings p { margin: 0; }
.activity-settings p { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.8rem; }
.activity-settings .button { justify-self: start; }
@media (min-width: 760px) { .activity-settings { grid-template-columns: minmax(180px, 0.8fr) repeat(2, minmax(190px, 1fr)) auto; align-items: end; } }
</style>
