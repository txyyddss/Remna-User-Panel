<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare } from '@/api/types'
import { useI18n } from '@/i18n'
import StatisticsConcentricDonut from './StatisticsConcentricDonut.vue'
import { formatStatisticNumber, shareTotal } from './statisticsFormat'

const props = defineProps<{ items: readonly NamedShare[] }>()
const { t } = useI18n()
const includedStatuses = new Set(['paid', 'expired', 'cancelled'])

function providerItems(provider: 'ezpay' | 'bepusdt'): NamedShare[] {
  return props.items
    .filter((item) => {
      const [itemProvider, status] = item.id.split(':')
      return itemProvider === provider && includedStatuses.has(status ?? '')
    })
    .map((item) => {
      const status = item.id.split(':')[1] ?? ''
      return { ...item, label: t(`statistics.paymentStatus.${status}`) }
    })
}

const ezpayItems = computed(() => providerItems('ezpay'))
const bepusdtItems = computed(() => providerItems('bepusdt'))
const rings = computed(() => [
  { id: 'ezpay', label: t('statistics.provider.ezpay'), items: ezpayItems.value },
  { id: 'bepusdt', label: t('statistics.provider.bepusdt'), items: bepusdtItems.value },
] as const)
const total = computed(() => shareTotal([...ezpayItems.value, ...bepusdtItems.value]))
const hasPayments = computed(() => total.value > 0)
</script>

<template>
  <StatisticsConcentricDonut
    v-if="hasPayments"
    :center-label="$t('statistics.paymentStates')"
    :center-value="formatStatisticNumber(total)"
    :chart-label="$t('statistics.paymentChartLabel')"
    :center-icon="'i-ph-wallet'"
    :rings="rings"
  />
  <div v-else class="statistics-empty statistics-empty--panel">
    <UIcon name="i-ph-wallet" aria-hidden="true" />
    <span>{{ $t('statistics.noPayments') }}</span>
  </div>
</template>
