<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare } from '@/api/types'
import { ringSegments } from './statisticsGeometry'
import { chartSegments, formatStatisticPercent } from './statisticsFormat'

interface ConcentricRing {
  id: string
  label: string
  items: readonly NamedShare[]
}

const props = defineProps<{
  rings: readonly [ConcentricRing, ConcentricRing]
  centerLabel: string
  centerValue: string
  chartLabel: string
  centerIcon?: string
}>()

const chartRings = computed(() => props.rings.map((ring) => ({
  ...ring,
  segments: chartSegments(ring.items),
})))
const hasSegments = computed(() => chartRings.value.some((ring) => ring.segments.length > 0))
</script>

<template>
  <div v-if="hasSegments" class="statistics-concentric-layout">
    <div class="statistics-concentric-rings">
      <svg viewBox="0 0 120 120" role="img" :aria-label="chartLabel">
        <title>{{ chartLabel }}</title>
        <template v-for="(ring, index) in chartRings" :key="ring.id">
          <circle
            class="statistics-ring-track"
            :class="index === 0 ? 'statistics-concentric-ring--outer' : 'statistics-concentric-ring--inner'"
            cx="60"
            cy="60"
            :r="index === 0 ? 50 : 32"
            pathLength="100"
          />
          <circle
            v-for="segment in ringSegments(ring.segments)"
            :key="segment.id"
            class="statistics-ring-segment"
            :class="index === 0 ? 'statistics-concentric-ring--outer' : 'statistics-concentric-ring--inner'"
            cx="60"
            cy="60"
            :r="index === 0 ? 50 : 32"
            pathLength="100"
            :stroke="segment.color"
            :stroke-dasharray="segment.dasharray"
            :stroke-dashoffset="segment.dashoffset"
          />
        </template>
      </svg>
      <UIcon v-if="centerIcon" class="statistics-concentric__icon" :name="centerIcon" aria-hidden="true" />
      <div v-else class="statistics-concentric__center" aria-hidden="true">
        <strong>{{ centerValue }}</strong>
        <span>{{ centerLabel }}</span>
      </div>
    </div>
    <div class="statistics-concentric-legends">
      <div v-for="ring in chartRings" :key="ring.id">
        <strong>{{ ring.label }}</strong>
        <ul v-if="ring.segments.length" class="statistics-legend">
          <li v-for="segment in ring.segments" :key="segment.id">
            <span class="statistics-legend__swatch" :style="{ backgroundColor: segment.color }" />
            <span class="statistics-legend__label" :title="segment.label">{{ segment.label }}</span>
            <strong>{{ formatStatisticPercent(segment.percentage) }}</strong>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
