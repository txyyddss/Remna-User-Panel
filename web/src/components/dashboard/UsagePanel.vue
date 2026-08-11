<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import type { Purchase, RFC3339, UsageStatistics } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useDashboardNodeUsage } from '@/composables/useDashboard'
import { formatBytes, formatDateTime } from '@/utils/format'
import TrafficUsageDetails from './TrafficUsageDetails.vue'

const props = defineProps<{
  statistics: UsageStatistics
  ratio: number
  stale: boolean
  fetchedAt: RFC3339
  term?: Purchase | null
}>()

const percentage = computed(() => Math.min(100, Math.max(0, Math.round(props.ratio * 100))))
const tone = computed(() => percentage.value >= 90 ? 'danger' : percentage.value >= 75 ? 'warning' : 'safe')
const meterStyle = computed(() => ({ '--usage': `${percentage.value}%` }))
const detailsOpen = shallowRef(false)
const nodeUsage = useDashboardNodeUsage()

watch(detailsOpen, (open) => {
  if (open) void nodeUsage.loadNodeUsage()
})
</script>

<template>
  <section class="section-block home-usage" :class="`home-usage--${tone}`">
    <div class="section-heading">
      <h2>{{ $t('dashboard.usage') }}</h2>
      <div class="home-usage__heading-actions">
        <span class="section-heading__meta">{{ stale ? $t('dashboard.lastKnownData') : $t('dashboard.livePeriod') }}</span>
        <UPopover v-model:open="detailsOpen" :content="{ side: 'bottom', align: 'end', sideOffset: 10, collisionPadding: 16 }">
          <UButton
            class="home-usage__details-trigger"
            color="neutral"
            variant="ghost"
            icon="i-ph-question"
            :aria-label="$t('home.trafficDetails')"
            data-haptic
          />
          <template #content>
            <TrafficUsageDetails
              :start-date="nodeUsage.nodeUsageStart.value"
              :end-date="nodeUsage.nodeUsageEnd.value"
              :term="term"
              :usage="nodeUsage.nodeUsage.value"
              :loading="nodeUsage.nodeUsageLoading.value"
              :error="nodeUsage.nodeUsageError.value"
              @load="nodeUsage.loadNodeUsage"
              @update:start-date="nodeUsage.setNodeUsageStart"
              @update:end-date="nodeUsage.setNodeUsageEnd"
            />
          </template>
        </UPopover>
      </div>
    </div>

    <InlineNotice v-if="stale" tone="warning">
      {{ $t('dashboard.remnawaveUnavailable') }} {{ $t('dashboard.showingDataFrom', { date: formatDateTime(fetchedAt) }) }}
    </InlineNotice>

    <div class="home-usage__summary">
      <div>
        <span>{{ $t('dashboard.thisTerm') }}</span>
        <strong>{{ formatBytes(statistics.usedTrafficBytes) }}</strong>
        <small>{{ $t('dashboard.ofLimit', { amount: formatBytes(statistics.trafficLimitBytes) }) }}</small>
      </div>
      <strong class="home-usage__percentage">{{ percentage }}%</strong>
    </div>
    <div
      class="home-usage__meter"
      :style="meterStyle"
      role="progressbar"
      :aria-label="$t('dashboard.usage')"
      :aria-valuemax="100"
      :aria-valuemin="0"
      :aria-valuenow="percentage"
    >
      <span />
    </div>
  </section>
</template>
