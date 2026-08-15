<script setup lang="ts">
import { computed } from 'vue'

import type { GroupMessageRewardStatus } from '@/api/features'
import { formatMoney } from '@/utils/format'

const props = defineProps<{ reward: GroupMessageRewardStatus }>()
const progress = computed(() => props.reward.threshold > 0 ? Math.min(100, (props.reward.messageCount / props.reward.threshold) * 100) : 0)
const amount = computed(() => formatMoney({ minor: props.reward.rewardMinor, currency: 'TXB', display: '' }))
const remaining = computed(() => Math.max(props.reward.threshold - props.reward.messageCount, 0))
const statusLabel = computed(() => props.reward.rewarded ? 'activity.groupRewardClaimed' : remaining.value === 0 ? 'activity.groupRewardAvailable' : 'activity.groupRewardInProgress')
</script>

<template>
  <section class="section-block group-reward-panel" :aria-label="$t('activity.groupRewardTitle')">
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('activity.groupRewardTitle') }}</h2>
      <p v-if="!reward.enabled">{{ $t('activity.groupRewardDisabled') }}</p>
      <p v-else>{{ $t('activity.groupRewardHint') }}</p>
    </div>
    <template v-if="reward.enabled">
      <div class="group-reward-panel__card">
        <div class="group-reward-panel__stats">
          <div><span>{{ $t('activity.groupRewardCount') }}</span><strong>{{ reward.messageCount }} / {{ reward.threshold }}</strong></div>
          <div><span>{{ $t('activity.groupRewardRemaining') }}</span><strong>{{ remaining }}</strong></div>
          <div><span>{{ $t('activity.groupRewardAmount') }}</span><strong>{{ amount }}</strong></div>
        </div>
        <div class="progress-track" role="progressbar" :aria-label="$t('activity.groupRewardProgressLabel')" :aria-valuenow="reward.messageCount" :aria-valuemin="0" :aria-valuemax="reward.threshold"><span :style="{ width: `${progress}%` }" /></div>
        <p class="group-reward-panel__status" :class="{ 'group-reward-panel__status--ready': remaining === 0 && !reward.rewarded }">{{ $t(statusLabel) }}</p>
      </div>
    </template>
  </section>
</template>

<style scoped>
.group-reward-panel__card { display: grid; gap: 0.85rem; padding: 1rem; border: 1px solid var(--line); border-radius: 1rem; background: var(--surface); }
.group-reward-panel__stats { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.75rem; }
.group-reward-panel__stats div { display: grid; gap: 0.2rem; }
.group-reward-panel__stats span, .group-reward-panel__status { color: var(--text-faint); font-size: 0.78rem; }
.group-reward-panel__stats strong { color: var(--text-strong); font-size: 1rem; }
.group-reward-panel__status { margin: 0; }
.group-reward-panel__status--ready { color: var(--success); }
@media (max-width: 30rem) { .group-reward-panel__stats { grid-template-columns: 1fr 1fr; } .group-reward-panel__stats div:last-child { grid-column: 1 / -1; } }
</style>
