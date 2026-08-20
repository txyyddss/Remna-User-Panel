<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare } from '@/api/types'
import { pieSlices } from './statisticsGeometry'
import StatisticsChartDetail from './StatisticsChartDetail.vue'
import { chartSegments, formatStatisticNumber, formatStatisticPercent } from './statisticsFormat'
import { useStatisticsChartSelection } from './useStatisticsChartSelection'

const props = withDefaults(defineProps<{
  items: readonly NamedShare[]
  chartLabel: string
  sliceLabels?: boolean
}>(), { sliceLabels: false })

const segments = computed(() => chartSegments(props.items).map((segment) => ({
  ...segment,
  interactionId: segment.id,
})))
const slices = computed(() => pieSlices(segments.value))
const { activeItem, hasActive, activate, deactivate, select, isActive, isSelected } = useStatisticsChartSelection(segments)
</script>

<template>
  <div class="statistics-pie-layout" :class="{ 'statistics-pie-layout--labels': sliceLabels }">
    <svg
      v-if="slices.length"
      class="statistics-pie"
      viewBox="0 0 200 200"
      role="group"
      :aria-label="chartLabel"
    >
      <title>{{ chartLabel }}</title>
      <g>
        <path
          v-for="slice in slices"
          :key="slice.id"
          class="statistics-chart-segment"
          :class="{ 'statistics-chart-segment--muted': hasActive && !isActive(slice.interactionId) }"
          :d="slice.path"
          :fill="slice.color"
          role="button"
          tabindex="0"
          data-haptic
          :aria-label="$t('statistics.chartSeries', { series: slice.label, segment: $t('statistics.chartPointValue', { value: formatStatisticNumber(slice.value), percent: formatStatisticPercent(slice.percentage) }) })"
          :aria-pressed="isSelected(slice.interactionId)"
          @pointerenter="activate(slice.interactionId)"
          @pointerleave="deactivate(slice.interactionId)"
          @focus="activate(slice.interactionId)"
          @blur="deactivate(slice.interactionId)"
          @click="select(slice.interactionId)"
          @keydown.enter.prevent="select(slice.interactionId)"
          @keydown.space.prevent="select(slice.interactionId)"
        />
        <template v-if="sliceLabels">
          <template v-for="slice in slices" :key="`label-${slice.id}`">
            <path v-if="slice.showLabel" class="statistics-pie__line" :d="slice.labelLine" aria-hidden="true" />
            <text
              v-if="slice.showLabel"
              class="statistics-pie__label"
              :x="slice.labelX"
              :y="slice.labelY"
              :text-anchor="slice.labelAnchor"
              dominant-baseline="middle"
              aria-hidden="true"
            >{{ formatStatisticPercent(slice.percentage) }}</text>
          </template>
        </template>
      </g>
    </svg>
    <p v-else class="statistics-empty">{{ $t('statistics.noData') }}</p>
    <ul v-if="segments.length" class="statistics-legend">
      <li
        v-for="segment in segments"
        :key="segment.id"
        class="statistics-chart-legend-item"
        role="button"
        tabindex="0"
        data-haptic
        :aria-pressed="isSelected(segment.interactionId)"
        @pointerenter="activate(segment.interactionId)"
        @pointerleave="deactivate(segment.interactionId)"
        @focus="activate(segment.interactionId)"
        @blur="deactivate(segment.interactionId)"
        @click="select(segment.interactionId)"
        @keydown.enter.prevent="select(segment.interactionId)"
        @keydown.space.prevent="select(segment.interactionId)"
      >
        <span class="statistics-legend__swatch" :style="{ backgroundColor: segment.color }" />
        <span class="statistics-legend__label" :title="segment.label">{{ segment.label }}</span>
        <strong>{{ formatStatisticPercent(segment.percentage) }}</strong>
      </li>
    </ul>
    <StatisticsChartDetail
      v-if="activeItem"
      :color="activeItem.color"
      :label="activeItem.label"
      :value="$t('statistics.chartPointValue', { value: formatStatisticNumber(activeItem.value), percent: formatStatisticPercent(activeItem.percentage) })"
    />
  </div>
</template>
