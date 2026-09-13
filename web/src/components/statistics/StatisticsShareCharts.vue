<script setup lang="ts">
import { motion } from 'motion-v'

import type { StatisticsSnapshot } from '@/api/types'
import StatisticsMembershipDonut from './StatisticsMembershipDonut.vue'
import StatisticsPaymentDonut from './StatisticsPaymentDonut.vue'
import { useMotionPreferences } from '@/composables/useMotionPreferences'

const props = defineProps<{ database: StatisticsSnapshot['database'] }>()
const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <section class="statistics-section">
    <div class="statistics-section__heading"><h2>{{ $t('statistics.distributions') }}</h2></div>
    <div class="statistics-share-grid">
      <motion.article
        class="statistics-panel"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 4 }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: reducedMotion ? 0.08 : 0.18, ease: 'easeOut' }"
      >
        <h3>{{ $t('statistics.members') }}</h3>
        <StatisticsMembershipDonut :database="props.database" />
      </motion.article>
      <motion.article
        class="statistics-panel"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 4 }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: reducedMotion ? 0.08 : 0.18, delay: reducedMotion ? 0 : 0.03, ease: 'easeOut' }"
      >
        <h3>{{ $t('statistics.paymentStates') }}</h3>
        <StatisticsPaymentDonut :items="props.database.paymentStatuses" />
      </motion.article>
    </div>
  </section>
</template>
