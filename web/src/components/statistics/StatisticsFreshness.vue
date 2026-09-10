<script setup lang="ts">
import { computed } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { StatisticsSnapshot } from '@/api/types'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ snapshot: StatisticsSnapshot }>()
const { reducedMotion } = useMotionPreferences()

const partitions = computed(() => [
  {
    id: 'remote',
    labelKey: 'statistics.remoteUpdatedAt',
    generatedAt: props.snapshot.remoteGeneratedAt,
  },
  {
    id: 'database',
    labelKey: 'statistics.databaseUpdatedAt',
    generatedAt: props.snapshot.databaseGeneratedAt,
  },
].map((partition) => ({
  ...partition,
  stale: props.snapshot.stalePartitions.includes(partition.id),
})))
</script>

<template>
  <section
    v-for="partition in partitions"
    :key="partition.id"
    class="statistics-section statistics-freshness"
    :aria-labelledby="`statistics-freshness-${partition.id}`"
  >
    <div class="statistics-section__heading">
      <h2 :id="`statistics-freshness-${partition.id}`">{{ $t(partition.labelKey) }}</h2>
    </div>
    <AnimatePresence mode="wait" :initial="false">
      <motion.div
        :key="`${partition.generatedAt}:${partition.stale}`"
        class="statistics-freshness__value"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, scale: 0.985 }"
        :animate="{ opacity: 1, scale: 1 }"
        :exit="{ opacity: 0 }"
        :transition="{ duration: reducedMotion ? 0.08 : 0.24, ease: 'easeOut' }"
      >
        <time :datetime="partition.generatedAt">{{ formatDateTime(partition.generatedAt) }}</time>
        <span :class="{ 'statistics-freshness__stale': partition.stale }">
          {{ $t(partition.stale ? 'statistics.partitionStale' : 'statistics.partitionCurrent') }}
        </span>
      </motion.div>
    </AnimatePresence>
  </section>
</template>
