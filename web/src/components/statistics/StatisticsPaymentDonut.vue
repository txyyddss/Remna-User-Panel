<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare } from '@/api/types'
import { useI18n } from '@/i18n'
import { ringSegments } from './statisticsGeometry'
import { chartSegments, formatStatisticPercent, shareTotal } from './statisticsFormat'

const props = defineProps<{ items: readonly NamedShare[] }>()
const { t } = useI18n()

function providerItems(provider: 'bepusdt' | 'ezpay'): NamedShare[] {
  return props.items
    .filter((item) => item.id.startsWith(`${provider}:`))
    .map((item) => {
      const status = item.id.split(':')[1] ?? ''
      return { ...item, label: t(`statistics.paymentStatus.${status}`) }
    })
}

const bepItems = computed(() => providerItems('bepusdt'))
const ezItems = computed(() => providerItems('ezpay'))
const bepSegments = computed(() => chartSegments(bepItems.value))
const ezSegments = computed(() => chartSegments(ezItems.value))
const bepRings = computed(() => ringSegments(bepSegments.value))
const ezRings = computed(() => ringSegments(ezSegments.value))
const hasPayments = computed(() => shareTotal(props.items) > 0)
const providerLegends = computed(() => [
  { id: 'ezpay', segments: ezSegments.value },
  { id: 'bepusdt', segments: bepSegments.value },
])
</script>

<template>
  <div v-if="hasPayments" class="statistics-payment-layout">
    <div class="statistics-payment-rings">
      <svg viewBox="0 0 120 120" role="img" :aria-label="$t('statistics.paymentChartLabel')">
        <title>{{ $t('statistics.paymentChartLabel') }}</title>
        <circle class="statistics-ring-track statistics-payment-ring--outer" cx="60" cy="60" r="50" pathLength="100" />
        <circle
          v-for="ring in ezRings"
          :key="`ez-${ring.id}`"
          class="statistics-ring-segment statistics-payment-ring--outer"
          cx="60"
          cy="60"
          r="50"
          pathLength="100"
          :stroke="ring.color"
          :stroke-dasharray="ring.dasharray"
          :stroke-dashoffset="ring.dashoffset"
        />
        <circle class="statistics-ring-track statistics-payment-ring--inner" cx="60" cy="60" r="32" pathLength="100" />
        <circle
          v-for="ring in bepRings"
          :key="`bep-${ring.id}`"
          class="statistics-ring-segment statistics-payment-ring--inner"
          cx="60"
          cy="60"
          r="32"
          pathLength="100"
          :stroke="ring.color"
          :stroke-dasharray="ring.dasharray"
          :stroke-dashoffset="ring.dashoffset"
        />
      </svg>
      <UIcon class="statistics-payment-rings__center" name="i-ph-wallet" aria-hidden="true" />
    </div>
    <div class="statistics-payment-legends">
      <div v-for="provider in providerLegends" :key="provider.id">
        <strong>{{ $t(`statistics.provider.${provider.id}`) }}</strong>
        <ul class="statistics-legend">
          <li v-for="segment in provider.segments" :key="segment.id">
            <span class="statistics-legend__swatch" :style="{ backgroundColor: segment.color }" />
            <span class="statistics-legend__label">{{ segment.label }}</span>
            <strong>{{ formatStatisticPercent(segment.percentage) }}</strong>
          </li>
        </ul>
      </div>
    </div>
  </div>
  <div v-else class="statistics-empty statistics-empty--panel">
    <UIcon name="i-ph-wallet" aria-hidden="true" />
    <span>{{ $t('statistics.noPayments') }}</span>
  </div>
</template>
