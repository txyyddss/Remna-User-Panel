<script setup lang="ts">
import type { SquadProduct } from '@/api/types'
import StatisticsGeocheckModal from '@/components/statistics/StatisticsGeocheckModal.vue'
import { useNodeGeocheck } from '@/composables/useNodeGeocheck'
import SquadSelector from './SquadSelector.vue'

defineProps<{
  squads: readonly SquadProduct[]
  selectedIds: readonly string[]
  includedIds: readonly string[]
  featuredIds: readonly string[]
  orderedIds: readonly string[]
}>()

const emit = defineEmits<{ toggle: [id: string] }>()
const geocheck = useNodeGeocheck()
</script>

<template>
  <SquadSelector
    :squads="squads"
    :selected-ids="selectedIds"
    :included-ids="includedIds"
    :featured-ids="featuredIds"
    :ordered-ids="orderedIds"
    @toggle="emit('toggle', $event)"
    @open-geocheck="geocheck.show"
  />
  <StatisticsGeocheckModal v-model:open="geocheck.isOpen.value" :node="geocheck.selectedNode.value" :result="geocheck.result.value" :loading="geocheck.loading.value" :error="geocheck.error.value" />
</template>
