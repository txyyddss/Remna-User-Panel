<script setup lang="ts">
import { PhArrowClockwise } from '@phosphor-icons/vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useActivity } from '@/composables/useActivity'
import { formatMoney } from '@/utils/format'
import BetGamesPanel from './BetGamesPanel.vue'
import DailyCheckInCard from './DailyCheckInCard.vue'
import LuckyDrawPanel from './LuckyDrawPanel.vue'
import GroupMessageRewardPanel from './GroupMessageRewardPanel.vue'
import ActivityResultDialog from './ActivityResultDialog.vue'

const { overview, result, loading, busy, error, load, checkIn, placeBet, draw, clearResult } = useActivity()
</script>

<template>
  <div class="page page--activity">
    <header class="page-header">
      <p class="eyebrow">{{ $t('activity.eyebrow') }}</p>
      <h1>{{ $t('activity.title') }}</h1>
      <p>{{ $t('activity.copy') }}</p>
    </header>
    <template v-if="loading">
      <SkeletonBlock height="9rem" />
      <div class="activity-layout"><SkeletonBlock height="24rem" /><SkeletonBlock height="12rem" /></div>
    </template>
    <template v-else-if="overview">
      <div class="page-toolbar">
        <p>{{ $t('activity.balance', { amount: formatMoney(overview.balance) }) }}</p>
        <button class="text-button" type="button" :aria-label="$t('common.refresh')" @click="load({ quiet: true })"><PhArrowClockwise :size="17" /> {{ $t('common.refresh') }}</button>
      </div>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <div class="activity-layout">
        <DailyCheckInCard
          :checked-in="overview.checkedInToday"
          :reward-min-txb-minor="overview.dailyRewardMinTxbMinor"
          :reward-max-txb-minor="overview.dailyRewardMaxTxbMinor"
          :time-zone="overview.timeZone"
          :busy="busy === 'check-in'"
          @check-in="checkIn"
        />
        <GroupMessageRewardPanel :reward="overview.groupMessageReward" />
        <BetGamesPanel :games="overview.games" :busy="busy === 'bet'" @bet="placeBet" />
        <LuckyDrawPanel :draws="overview.draws" :busy="busy === 'draw'" @draw="draw" />
      </div>
      <ActivityResultDialog :result="result" @close="clearResult" />
    </template>
    <div v-else class="error-state">
      <h1>{{ $t('errors.activityUnavailable') }}</h1><p>{{ error }}</p>
      <button class="button button--primary" type="button" @click="load()">{{ $t('common.tryAgain') }}</button>
    </div>
  </div>
</template>

<style scoped>
.activity-layout { display: grid; gap: 0.9rem; }
</style>
