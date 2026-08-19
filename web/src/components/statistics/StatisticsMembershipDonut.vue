<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare, StatisticsSnapshot } from '@/api/types'
import { useI18n } from '@/i18n'
import StatisticsConcentricDonut from './StatisticsConcentricDonut.vue'
import { formatStatisticNumber, shareTotal } from './statisticsFormat'

const props = defineProps<{ database: StatisticsSnapshot['database'] }>()
const { t } = useI18n()

const subscriptions = computed<NamedShare[]>(() => props.database.subscriptionStates.map((item) => ({
  ...item,
  label: t(`statistics.subscription.${item.id}`),
})))
const activeUsers = computed(() => shareTotal(props.database.comboShares))
const rings = computed(() => [
  { id: 'subscriptions', label: t('statistics.subscriptionStates'), items: subscriptions.value },
  { id: 'combos', label: t('statistics.comboChoices'), items: props.database.comboShares },
] as const)
</script>

<template>
  <StatisticsConcentricDonut
    :rings="rings"
    :center-label="$t('statistics.active')"
    :center-value="formatStatisticNumber(activeUsers)"
    :chart-label="$t('statistics.subscriptionChartLabel')"
  />
</template>
