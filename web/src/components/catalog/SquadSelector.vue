<script setup lang="ts">
import { CheckboxIndicator, CheckboxRoot } from 'reka-ui'
import { PhCheck, PhGlobeHemisphereWest } from '@phosphor-icons/vue'

import type { SquadProduct } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { formatMoney } from '@/utils/format'

const props = defineProps<{
  squads: readonly SquadProduct[]
  selectedIds: readonly string[]
  includedIds: readonly string[]
}>()

const emit = defineEmits<{ toggle: [id: string] }>()

function isSelected(id: string): boolean {
  return props.selectedIds.includes(id)
}

function isIncluded(id: string): boolean {
  return props.includedIds.includes(id)
}
</script>

<template>
  <section class="squad-selector">
    <div class="section-heading section-heading--stacked">
      <h2>Optional squads</h2>
      <p>Add regions to the same traffic budget and subscription term.</p>
    </div>
    <div v-if="squads.length" class="squad-grid">
      <label v-for="squad in squads" :key="squad.id" class="squad-option" :class="{ 'squad-option--selected': isSelected(squad.id), 'squad-option--included': isIncluded(squad.id) }">
        <span class="squad-option__icon"><PhGlobeHemisphereWest :size="21" /></span>
        <span class="squad-option__copy">
          <strong>{{ squad.name }}</strong>
          <MarkdownContent :source="squad.description" compact />
          <StatusBadge v-if="isIncluded(squad.id)" tone="neutral" label="Included" />
          <span v-else>{{ formatMoney(squad.price) }}</span>
        </span>
        <CheckboxRoot
          class="checkbox-control"
          :model-value="isSelected(squad.id)"
          :disabled="isIncluded(squad.id)"
          @update:model-value="emit('toggle', squad.id)"
        >
          <CheckboxIndicator class="checkbox-indicator"><PhCheck :size="16" weight="bold" /></CheckboxIndicator>
        </CheckboxRoot>
      </label>
    </div>
    <div v-else class="empty-inline">
      <div>
        <h3>No optional squads today</h3>
        <p>Your combo still includes its standard regions.</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.squad-option--included {
  color: var(--text-faint);
  border-color: var(--line);
  background: var(--canvas-soft);
  cursor: default;
}
</style>
