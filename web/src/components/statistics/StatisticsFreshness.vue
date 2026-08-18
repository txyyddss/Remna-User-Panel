<script setup lang="ts">
import { computed } from 'vue'

import type { StatisticsSnapshot } from '@/api/types'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ snapshot: StatisticsSnapshot }>()

const partitions = computed(() => [
  {
    id: 'remote',
    labelKey: 'statistics.remoteData',
    generatedAt: props.snapshot.remoteGeneratedAt,
  },
  {
    id: 'database',
    labelKey: 'statistics.databaseData',
    generatedAt: props.snapshot.databaseGeneratedAt,
  },
].map((partition) => ({
  ...partition,
  stale: props.snapshot.stalePartitions.includes(partition.id),
})))
</script>

<template>
  <dl class="statistics-freshness" :aria-label="$t('statistics.freshnessLabel')">
    <div v-for="partition in partitions" :key="partition.id">
      <dt>{{ $t(partition.labelKey) }}</dt>
      <dd>
        <time :datetime="partition.generatedAt">{{ formatDateTime(partition.generatedAt) }}</time>
        <span :class="{ 'statistics-freshness__stale': partition.stale }">
          {{ $t(partition.stale ? 'statistics.partitionStale' : 'statistics.partitionCurrent') }}
        </span>
      </dd>
    </div>
  </dl>
</template>
