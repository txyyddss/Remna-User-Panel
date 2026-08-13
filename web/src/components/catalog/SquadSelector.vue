<script setup lang="ts">
import type { SquadProduct } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SquadProfileSummary from '@/components/squad-profile/SquadProfileSummary.vue'
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
      <h2>{{ $t('catalog.optionalSquads') }}</h2>
      <p>{{ $t('catalog.optionalSquadsHint') }}</p>
    </div>
    <div v-if="squads.length" v-auto-animate class="squad-grid">
      <label v-for="squad in squads" :key="squad.id" class="squad-option" :class="{ 'squad-option--selected': isSelected(squad.id), 'squad-option--included': isIncluded(squad.id) }" :data-haptic="isIncluded(squad.id) ? undefined : 'light'">
        <div class="squad-option__copy">
          <SquadProfileSummary :name="squad.name" :profile="squad.profile" :description="squad.description" presentation="member" compact />
          <StatusBadge v-if="isIncluded(squad.id)" tone="neutral" :label="$t('catalog.included')" />
          <span v-else>{{ formatMoney(squad.price) }}</span>
        </div>
        <UCheckbox
          :model-value="isSelected(squad.id)"
          :disabled="isIncluded(squad.id)"
          :aria-label="squad.name"
          @update:model-value="emit('toggle', squad.id)"
        />
      </label>
    </div>
    <div v-else class="empty-inline">
      <div>
        <h3>{{ $t('catalog.noSquads') }}</h3>
        <p>{{ $t('catalog.noSquadsHint') }}</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.squad-option--included { color: var(--text-faint); border-color: var(--line); background: var(--canvas-soft); cursor: default; }
</style>
