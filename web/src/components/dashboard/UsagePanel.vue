<script setup lang="ts">
import { computed } from 'vue'
import { PhChartDonut, PhDatabase } from '@phosphor-icons/vue'

import type { RFC3339, UsageStatistics } from '@/api/types'
import { formatBytes, formatDateTime } from '@/utils/format'
import InlineNotice from '@/components/common/InlineNotice.vue'

const props = defineProps<{
  statistics: UsageStatistics
  ratio: number
  stale: boolean
  warning?: string
  fetchedAt: RFC3339
}>()

const percentage = computed(() => Math.round(props.ratio * 100))
const ringStyle = computed(() => ({ '--usage': `${percentage.value * 3.6}deg` }))
</script>

<template>
  <section class="section-block usage-section">
    <div class="section-heading">
      <h2>{{ $t('dashboard.usage') }}</h2>
      <span class="section-heading__meta">{{ stale ? $t('dashboard.lastKnownData') : $t('dashboard.livePeriod') }}</span>
    </div>

    <InlineNotice v-if="stale" tone="warning">
      {{ warning || $t('dashboard.remnawaveUnavailable') }} {{ $t('dashboard.showingDataFrom', { date: formatDateTime(fetchedAt) }) }}
    </InlineNotice>

    <div class="usage-layout">
      <div class="usage-ring" :style="ringStyle">
        <div class="usage-ring__inner">
          <strong>{{ percentage }}%</strong>
          <span>{{ $t('dashboard.used') }}</span>
        </div>
      </div>
      <dl class="usage-totals">
        <div>
          <dt><PhChartDonut :size="18" /> {{ $t('dashboard.thisTerm') }}</dt>
          <dd>{{ formatBytes(statistics.usedTrafficBytes) }}</dd>
          <small>{{ $t('dashboard.ofLimit', { amount: formatBytes(statistics.trafficLimitBytes) }) }}</small>
        </div>
        <div>
          <dt><PhDatabase :size="18" /> {{ $t('dashboard.lifetime') }}</dt>
          <dd>{{ formatBytes(statistics.lifetimeTrafficBytes) }}</dd>
          <small>{{ $t('dashboard.acrossSubscriptions') }}</small>
        </div>
      </dl>
    </div>

    <div v-if="statistics.topNodes.length" class="node-list">
      <h3>{{ $t('dashboard.topNodes') }}</h3>
      <div v-for="node in statistics.topNodes.slice(0, 4)" :key="node.name" class="node-row">
        <span>{{ node.name }}</span>
        <strong>{{ formatBytes(node.totalBytes) }}</strong>
      </div>
    </div>
  </section>
</template>
