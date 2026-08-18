<script setup lang="ts">
import { computed } from 'vue'

import type { StatisticsSnapshot } from '@/api/types'
import { statisticsColors, formatShortStatisticDate, formatStatisticBytes, safeStatisticBytes, sumStatisticBytes } from './statisticsFormat'

const props = defineProps<{ remote: StatisticsSnapshot['remote'] }>()

const legend = computed(() => props.remote.trafficSeries.map((series, index) => ({
  id: series.uuid,
  name: series.name,
  color: statisticsColors[index % statisticsColors.length],
  total: sumStatisticBytes(series.dailyBytes),
})))

const days = computed(() => {
  const raw = props.remote.trafficDates.map((date, dayIndex) => {
    const segments = props.remote.trafficSeries.map((series, seriesIndex) => ({
      id: series.uuid,
      name: series.name,
      color: statisticsColors[seriesIndex % statisticsColors.length],
      value: safeStatisticBytes(series.dailyBytes[dayIndex] ?? '0'),
    })).filter((segment) => segment.value > 0n)
    return { date, segments, total: segments.reduce((sum, segment) => sum + segment.value, 0n) }
  })
  const maximum = raw.reduce((current, day) => day.total > current ? day.total : current, 0n)
  return raw.map((day) => ({
    ...day,
    height: maximum === 0n ? 0 : Math.max(5, Number(day.total * 100n / maximum)),
    segments: day.segments.map((segment) => ({
      ...segment,
      share: day.total === 0n ? 0 : Number(segment.value * 10_000n / day.total) / 100,
    })),
  }))
})
const hasTraffic = computed(() => days.value.some((day) => day.total > 0n))
</script>

<template>
  <section class="statistics-section">
    <div class="statistics-section__heading"><h2>{{ $t('statistics.sevenDayTraffic') }}</h2></div>
    <article v-if="hasTraffic" class="statistics-panel statistics-traffic">
      <div class="statistics-traffic__plot" role="group" :aria-label="$t('statistics.trafficChartLabel')">
        <div v-for="day in days" :key="day.date" class="statistics-traffic__day">
          <div class="statistics-traffic__track">
            <div class="statistics-traffic__stack" :style="{ height: `${day.height}%` }">
              <UTooltip v-for="segment in day.segments" :key="segment.id" :content="{ side: 'top' }">
                <span
                  class="statistics-traffic__segment"
                  :style="{ backgroundColor: segment.color, flexBasis: `${segment.share}%` }"
                  :aria-label="$t('statistics.trafficPoint', { node: segment.name, date: day.date, value: formatStatisticBytes(segment.value) })"
                  tabindex="0"
                />
                <template #content>
                  <div class="statistics-chart-tooltip"><strong>{{ segment.name }}</strong><span>{{ day.date }}</span><span>{{ formatStatisticBytes(segment.value) }}</span></div>
                </template>
              </UTooltip>
            </div>
          </div>
          <span>{{ formatShortStatisticDate(day.date) }}</span>
        </div>
      </div>
      <ul class="statistics-traffic__legend">
        <li v-for="series in legend" :key="series.id">
          <span class="statistics-legend__swatch" :style="{ backgroundColor: series.color }" />
          <UTooltip :text="series.name"><span>{{ series.name }}</span></UTooltip>
          <strong>{{ formatStatisticBytes(series.total) }}</strong>
        </li>
      </ul>
    </article>
    <div v-else class="statistics-empty statistics-empty--panel"><UIcon name="i-ph-chart-line" aria-hidden="true" /><span>{{ $t('statistics.noTraffic') }}</span></div>
  </section>
</template>
