<script setup lang="ts">
import { computed } from 'vue'

import type { GroupMessageRewardStatus } from '@/api/features'
import { formatMoney } from '@/utils/format'

const props = defineProps<{ reward: GroupMessageRewardStatus }>()
const progress = computed(() => props.reward.threshold > 0 ? Math.min(100, (props.reward.messageCount / props.reward.threshold) * 100) : 0)
const amount = computed(() => formatMoney({ minor: props.reward.rewardMinor, currency: 'TXB', display: '' }))
</script>

<template>
  <section class="section-block group-reward-panel" :aria-label="$t('activity.groupRewardTitle')">
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('activity.groupRewardTitle') }}</h2>
      <p v-if="!reward.enabled">{{ $t('activity.groupRewardDisabled') }}</p>
      <p v-else>{{ $t('activity.groupRewardHint') }}</p>
    </div>
    <template v-if="reward.enabled">
      <div class="group-reward-panel__meta"><strong>{{ $t('activity.groupRewardProgress', { count: reward.messageCount, threshold: reward.threshold }) }}</strong><span>{{ amount }}</span></div>
      <div class="progress-track" role="progressbar" :aria-valuenow="reward.messageCount" :aria-valuemin="0" :aria-valuemax="reward.threshold"><span :style="{ width: `${progress}%` }" /></div>
      <p v-if="reward.rewarded" class="field-hint">{{ $t('activity.groupRewardClaimed') }}</p>
    </template>
  </section>
</template>
