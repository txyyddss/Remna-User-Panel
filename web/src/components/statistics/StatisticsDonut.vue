<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare } from '@/api/types'
import { ringSegments } from './statisticsGeometry'
import { chartSegments, formatStatisticPercent } from './statisticsFormat'

const props = defineProps<{
  items: readonly NamedShare[]
  centerLabel: string
  centerValue: string
  chartLabel: string
}>()

const segments = computed(() => chartSegments(props.items))
const rings = computed(() => ringSegments(segments.value))
</script>

<template>
  <div class="statistics-donut-layout">
    <div class="statistics-donut">
      <svg viewBox="0 0 120 120" role="img" :aria-label="chartLabel">
        <title>{{ chartLabel }}</title>
        <circle class="statistics-ring-track" cx="60" cy="60" r="47" pathLength="100" />
        <circle
          v-for="ring in rings"
          :key="ring.id"
          class="statistics-ring-segment"
          cx="60"
          cy="60"
          r="47"
          pathLength="100"
          :stroke="ring.color"
          :stroke-dasharray="ring.dasharray"
          :stroke-dashoffset="ring.dashoffset"
        />
      </svg>
      <div class="statistics-donut__center" aria-hidden="true">
        <strong>{{ centerValue }}</strong>
        <span>{{ centerLabel }}</span>
      </div>
    </div>
    <ul v-if="segments.length" class="statistics-legend">
      <li v-for="segment in segments" :key="segment.id">
        <span class="statistics-legend__swatch" :style="{ backgroundColor: segment.color }" />
        <span class="statistics-legend__label" :title="segment.label">{{ segment.label }}</span>
        <strong>{{ formatStatisticPercent(segment.percentage) }}</strong>
      </li>
    </ul>
    <p v-else class="statistics-empty">{{ $t('statistics.noData') }}</p>
  </div>
</template>
