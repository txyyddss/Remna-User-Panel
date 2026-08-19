<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useStatistics } from '@/composables/useStatistics'
import { formatDateTime } from '@/utils/format'
import StatisticsDistribution from './StatisticsDistribution.vue'
import StatisticsFreshness from './StatisticsFreshness.vue'
import StatisticsNodes from './StatisticsNodes.vue'
import StatisticsOverview from './StatisticsOverview.vue'
import StatisticsShareCharts from './StatisticsShareCharts.vue'
import StatisticsTrafficChart from './StatisticsTrafficChart.vue'

const { snapshot, nodeSnapshot, loading, refreshing, nodesLoading, error, nodesError, load } = useStatistics()
</script>

<template>
  <div class="page page--statistics">
    <header class="statistics-heading">
      <div>
        <p class="eyebrow">{{ $t('dashboard.aroundTx') }}</p>
        <h1>{{ $t('statistics.title') }}</h1>
        <p v-if="snapshot">{{ $t('statistics.updatedAt', { date: formatDateTime(snapshot.generatedAt) }) }}</p>
      </div>
      <UTooltip :text="$t('statistics.refresh')">
        <UButton
          type="button"
          color="neutral"
          variant="outline"
          square
          class="statistics-heading__refresh"
          icon="i-ph-arrows-clockwise"
          :loading="refreshing"
          :aria-label="$t('statistics.refresh')"
          @click="load({ quiet: true })"
        />
      </UTooltip>
    </header>

    <template v-if="loading && !snapshot">
      <SkeletonBlock height="16rem" />
      <div class="statistics-skeleton-grid"><SkeletonBlock height="13rem" /><SkeletonBlock height="13rem" /></div>
      <SkeletonBlock height="18rem" />
    </template>

    <template v-else-if="snapshot">
      <InlineNotice v-if="snapshot.stalePartitions.length" tone="warning">{{ $t('statistics.stalePartitions') }}</InlineNotice>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <InlineNotice v-if="nodesError" tone="warning">{{ nodesError }}</InlineNotice>
      <StatisticsFreshness :snapshot="snapshot" />
      <StatisticsOverview :snapshot="snapshot" />
      <StatisticsNodes :snapshot="nodeSnapshot" :loading="nodesLoading" />
      <StatisticsTrafficChart :remote="snapshot.remote" />
      <StatisticsShareCharts :database="snapshot.database" />
      <StatisticsDistribution :database="snapshot.database" />
    </template>

    <div v-else class="error-state">
      <h1>{{ $t('statistics.unavailable') }}</h1>
      <p>{{ error ?? $t('statistics.loadFailed') }}</p>
      <UButton icon="i-ph-arrows-clockwise" :label="$t('common.tryAgain')" @click="load()" />
    </div>
  </div>
</template>
