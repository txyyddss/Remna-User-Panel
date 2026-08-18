<script setup lang="ts">
import { computed } from 'vue'

import type { NamedShare, StatisticsSnapshot } from '@/api/types'
import { useI18n } from '@/i18n'
import StatisticsPaymentDonut from './StatisticsPaymentDonut.vue'
import StatisticsPie from './StatisticsPie.vue'

const props = defineProps<{ database: StatisticsSnapshot['database'] }>()
const { t } = useI18n()

const subscriptions = computed<NamedShare[]>(() => props.database.subscriptionStates.map((item) => ({
  ...item,
  label: t(`statistics.subscription.${item.id}`),
})))
</script>

<template>
  <section class="statistics-section">
    <div class="statistics-section__heading"><h2>{{ $t('statistics.distributions') }}</h2></div>
    <div class="statistics-share-grid">
      <article class="statistics-panel">
        <h3>{{ $t('statistics.subscriptionStates') }}</h3>
        <StatisticsPie
          :items="subscriptions"
          :chart-label="$t('statistics.subscriptionChartLabel')"
          slice-labels
        />
      </article>
      <article class="statistics-panel">
        <h3>{{ $t('statistics.comboChoices') }}</h3>
        <StatisticsPie
          :items="database.comboShares"
          :chart-label="$t('statistics.comboChartLabel')"
        />
      </article>
      <article class="statistics-panel">
        <h3>{{ $t('statistics.paymentStates') }}</h3>
        <StatisticsPaymentDonut :items="database.paymentStatuses" />
      </article>
    </div>
  </section>
</template>
