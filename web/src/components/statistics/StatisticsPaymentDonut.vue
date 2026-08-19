<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare } from '@/api/types'
import { useI18n } from '@/i18n'
import StatisticsDonut from './StatisticsDonut.vue'
import { formatStatisticNumber, shareTotal } from './statisticsFormat'

const props = defineProps<{ items: readonly NamedShare[] }>()
const { t } = useI18n()

const paymentItems = computed<NamedShare[]>(() => props.items.map((item) => {
  const [provider, status] = item.id.split(':')
  return {
    ...item,
    label: t('statistics.paymentLegend', {
      provider: t(`statistics.provider.${provider}`),
      status: t(`statistics.paymentStatus.${status}`),
    }),
  }
}))
const total = computed(() => shareTotal(props.items))
const hasPayments = computed(() => total.value > 0)
</script>

<template>
  <StatisticsDonut
    v-if="hasPayments"
    :items="paymentItems"
    :center-label="$t('statistics.paymentStates')"
    :center-value="formatStatisticNumber(total)"
    :chart-label="$t('statistics.paymentChartLabel')"
  />
  <div v-else class="statistics-empty statistics-empty--panel">
    <UIcon name="i-ph-wallet" aria-hidden="true" />
    <span>{{ $t('statistics.noPayments') }}</span>
  </div>
</template>
