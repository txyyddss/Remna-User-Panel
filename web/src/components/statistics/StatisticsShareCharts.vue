<script setup lang="ts">
import type { StatisticsSnapshot } from '@/api/types'
import StatisticsMembershipDonut from './StatisticsMembershipDonut.vue'
import StatisticsPaymentDonut from './StatisticsPaymentDonut.vue'

const props = defineProps<{ database: StatisticsSnapshot['database'] }>()
</script>

<template>
  <section class="statistics-section">
    <div class="statistics-section__heading"><h2>{{ $t('statistics.distributions') }}</h2></div>
    <div class="statistics-share-grid">
      <article class="statistics-panel">
        <h3>{{ $t('statistics.members') }}</h3>
        <StatisticsMembershipDonut :database="props.database" />
      </article>
      <article class="statistics-panel">
        <h3>{{ $t('statistics.paymentStates') }}</h3>
        <StatisticsPaymentDonut :items="props.database.paymentStatuses" />
      </article>
    </div>
  </section>
</template>
