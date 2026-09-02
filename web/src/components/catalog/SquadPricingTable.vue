<script setup lang="ts">
import { computed } from 'vue'

import type { SquadProduct } from '@/api/types'
import SquadPricingCard from './SquadPricingCard.vue'

type SquadType = NonNullable<SquadProduct['profile']>['type']

const props = defineProps<{
  profileType: SquadType
  squads: readonly SquadProduct[]
  selectedIds: readonly string[]
  includedIds: readonly string[]
  featuredIds: readonly string[]
}>()

const emit = defineEmits<{
  toggle: [id: string]
  openGeocheck: [node: SquadProduct['accessibleNodes'][number]]
}>()

const typeKey = computed(() => props.profileType === 'china_optimized'
  ? 'chinaOptimized'
  : props.profileType === 'international_network' ? 'internationalNetwork' : 'broadband')
</script>

<template>
  <section class="squad-pricing-table" :class="`squad-pricing-table--${profileType}`">
    <h3>{{ $t(`squadProfile.types.${typeKey}`) }}</h3>
    <div class="squad-pricing-grid">
      <SquadPricingCard
        v-for="squad in props.squads"
        :key="squad.id"
        :squad="squad"
        :selected="props.selectedIds.includes(squad.id)"
        :included="props.includedIds.includes(squad.id)"
        :featured="props.featuredIds.includes(squad.id)"
        @toggle="emit('toggle', $event)"
        @open-geocheck="emit('openGeocheck', $event)"
      />
    </div>
  </section>
</template>

<style scoped>
.squad-pricing-table { min-width: 0; display: grid; gap: 0.55rem; }
.squad-pricing-table > h3 { margin: 0; font-size: 0.92rem; }
.squad-pricing-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 18rem), 1fr));
  align-items: start;
  gap: 0.65rem;
}
@media (min-width: 900px) {
  .squad-pricing-grid { grid-template-columns: repeat(auto-fit, minmax(19rem, 1fr)); }
}
</style>
