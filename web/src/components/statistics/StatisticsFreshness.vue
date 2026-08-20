<script setup lang="ts">
import { computed } from 'vue'

import type { StatisticsSnapshot } from '@/api/types'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ snapshot: StatisticsSnapshot }>()

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
    <div class="statistics-freshness__value">
      <time :datetime="partition.generatedAt">{{ formatDateTime(partition.generatedAt) }}</time>
      <span :class="{ 'statistics-freshness__stale': partition.stale }">
        {{ $t(partition.stale ? 'statistics.partitionStale' : 'statistics.partitionCurrent') }}
      </span>
    </div>
  </section>
</template>
