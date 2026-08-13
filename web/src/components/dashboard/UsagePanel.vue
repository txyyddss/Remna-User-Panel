<script setup lang="ts">
import { computed, shallowRef } from 'vue'

import type { CatalogNode, RFC3339, UsageStatistics } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import { formatBytes, formatDateTime } from '@/utils/format'
import TrafficUsageBar from './TrafficUsageBar.vue'

const props = defineProps<{
  statistics: UsageStatistics
  ratio: number
  stale: boolean
  fetchedAt: RFC3339
  catalogNodes: readonly CatalogNode[]
}>()

const percentage = computed(() => Math.min(100, Math.max(0, Math.round(props.ratio * 100))))
const tone = computed(() => percentage.value >= 90 ? 'danger' : percentage.value >= 75 ? 'warning' : 'safe')
const meterStyle = computed(() => ({ '--usage': `${percentage.value}%` }))
const useTrafficMultiplier = shallowRef(true)
</script>

<template>
  <section class="section-block home-usage" :class="`home-usage--${tone}`">
    <div class="section-heading">
      <h2>{{ $t('dashboard.usage') }}</h2>
      <div class="home-usage__heading-actions">
        <SwitchField
          id="usage-traffic-multiplier"
          v-model="useTrafficMultiplier"
          :label="$t('dashboard.useTrafficMultiplier')"
          data-haptic
        />
      </div>
    </div>

    <InlineNotice v-if="stale" tone="warning">
      {{ $t('dashboard.lastKnownData') }} · {{ $t('dashboard.remnawaveUnavailable') }} {{ $t('dashboard.showingDataFrom', { date: formatDateTime(fetchedAt) }) }}
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
    <TrafficUsageBar
      :nodes="statistics.topNodes"
      :total-bytes="statistics.usedTrafficBytes"
      :catalog-nodes="catalogNodes"
      :use-multiplier="useTrafficMultiplier"
    />
  </section>
</template>
