<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare } from '@/api/types'
import { pieSlices } from './statisticsGeometry'
import { chartSegments, formatStatisticPercent } from './statisticsFormat'

const props = withDefaults(defineProps<{
  items: readonly NamedShare[]
  chartLabel: string
  sliceLabels?: boolean
}>(), { sliceLabels: false })

const segments = computed(() => chartSegments(props.items))
const slices = computed(() => pieSlices(segments.value))
</script>

<template>
  <div class="statistics-pie-layout" :class="{ 'statistics-pie-layout--labels': sliceLabels }">
    <svg
      v-if="slices.length"
      class="statistics-pie"
      viewBox="0 0 200 200"
      role="img"
      :aria-label="chartLabel"
    >
      <title>{{ chartLabel }}</title>
      <g aria-hidden="true">
        <path v-for="slice in slices" :key="slice.id" :d="slice.path" :fill="slice.color" />
        <template v-if="sliceLabels">
          <template v-for="slice in slices" :key="`label-${slice.id}`">
            <path v-if="slice.showLabel" class="statistics-pie__line" :d="slice.labelLine" />
            <text
              v-if="slice.showLabel"
              class="statistics-pie__label"
              :x="slice.labelX"
              :y="slice.labelY"
              :text-anchor="slice.labelAnchor"
              dominant-baseline="middle"
            >{{ formatStatisticPercent(slice.percentage) }}</text>
          </template>
        </template>
      </g>
    </svg>
    <p v-else class="statistics-empty">{{ $t('statistics.noData') }}</p>
    <ul v-if="segments.length" class="statistics-legend">
      <li v-for="segment in segments" :key="segment.id">
        <span class="statistics-legend__swatch" :style="{ backgroundColor: segment.color }" />
        <span class="statistics-legend__label" :title="segment.label">{{ segment.label }}</span>
        <strong>{{ formatStatisticPercent(segment.percentage) }}</strong>
      </li>
    </ul>
  </div>
</template>
