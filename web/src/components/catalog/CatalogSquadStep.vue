<script setup lang="ts">
import { computed, onMounted, onScopeDispose, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { NormalizedDistribution, SquadProduct } from '@/api/types'
import StatisticsGeocheckModal from '@/components/statistics/StatisticsGeocheckModal.vue'
import { useNodeGeocheck } from '@/composables/useNodeGeocheck'
import { featuredCatalogSquadIds } from './catalogFeaturedSquads'
import SquadSelector from './SquadSelector.vue'

const props = defineProps<{
  comboId: string | null
  squads: readonly SquadProduct[]
  selectedIds: readonly string[]
  includedIds: readonly string[]
}>()

const emit = defineEmits<{ toggle: [id: string] }>()
const composition = shallowRef<readonly NormalizedDistribution[]>([])
const geocheck = useNodeGeocheck()
let disposed = false

const featuredIds = computed(() => featuredCatalogSquadIds(
  props.squads,
  props.includedIds,
  composition.value,
  props.comboId,
))

onMounted(async () => {
  try {
    const snapshot = await api.getStatistics()
    if (!disposed) composition.value = snapshot.database.squadByCombo
  } catch { /* Featured analytics must not block catalog selection. */ }
})

onScopeDispose(() => { disposed = true })
</script>

<template>
  <SquadSelector
    :squads="squads"
    :selected-ids="selectedIds"
    :included-ids="includedIds"
    :featured-ids="featuredIds"
    @toggle="emit('toggle', $event)"
    @open-geocheck="geocheck.show"
  />
  <StatisticsGeocheckModal v-model:open="geocheck.isOpen.value" :node="geocheck.selectedNode.value" :result="geocheck.result.value" :loading="geocheck.loading.value" :error="geocheck.error.value" />
</template>
