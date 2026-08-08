<script setup lang="ts">
import { PhArrowClockwise } from '@phosphor-icons/vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useActivity } from '@/composables/useActivity'
import { formatMoney } from '@/utils/format'
import BetGamesPanel from './BetGamesPanel.vue'
import DailyCheckInCard from './DailyCheckInCard.vue'
import LuckyDrawPanel from './LuckyDrawPanel.vue'

const { overview, result, loading, busy, error, load, checkIn, placeBet, draw } = useActivity()
</script>

<template>
  <div class="page page--activity">
    <header class="page-header">
      <p class="eyebrow">Community activity</p>
      <h1>Play with clear odds.</h1>
      <p>Every fee, chance, and return is shown before you confirm.</p>
    </header>
    <template v-if="loading">
      <SkeletonBlock height="9rem" />
      <div class="activity-layout"><SkeletonBlock height="24rem" /><SkeletonBlock height="12rem" /></div>
    </template>
    <template v-else-if="overview">
      <div class="page-toolbar">
        <p>Balance {{ formatMoney(overview.balance) }}</p>
        <button class="text-button" type="button" @click="load({ quiet: true })"><PhArrowClockwise :size="17" /> Refresh</button>
      </div>
      <InlineNotice v-if="result" :tone="result.outcome === 'loss' ? 'warning' : 'success'" title="Result recorded">{{ result.message }}</InlineNotice>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <div class="activity-layout">
        <DailyCheckInCard
          :checked-in="overview.checkedInToday"
          :reward-txb-minor="overview.dailyRewardTxbMinor"
          :time-zone="overview.timeZone"
          :busy="busy === 'check-in'"
          @check-in="checkIn"
        />
        <BetGamesPanel :games="overview.games" :busy="busy === 'bet'" @bet="placeBet" />
        <LuckyDrawPanel :draws="overview.draws" :busy="busy === 'draw'" @draw="draw" />
      </div>
    </template>
    <div v-else class="error-state">
      <h1>Activity is unavailable.</h1><p>{{ error }}</p>
      <button class="button button--primary" type="button" @click="load()">Try again</button>
    </div>
  </div>
</template>

<style scoped>
.activity-layout { display: grid; gap: 0.9rem; }
</style>
