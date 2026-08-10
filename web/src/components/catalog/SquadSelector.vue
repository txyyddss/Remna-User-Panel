<script setup lang="ts">
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
      <h2>{{ $t('catalog.optionalSquads') }}</h2>
      <p>{{ $t('catalog.optionalSquadsHint') }}</p>
    </div>
    <div v-if="squads.length" v-auto-animate class="squad-grid">
      <label v-for="squad in squads" :key="squad.id" class="squad-option" :class="{ 'squad-option--selected': isSelected(squad.id), 'squad-option--included': isIncluded(squad.id) }">
        <span class="squad-option__icon"><UIcon name="i-ph-globe-hemisphere-west" /></span>
        <span class="squad-option__copy">
          <strong>{{ squad.name }}</strong>
          <MarkdownContent :source="squad.description" compact />
          <StatusBadge v-if="isIncluded(squad.id)" tone="neutral" :label="$t('catalog.included')" />
          <span v-else>{{ formatMoney(squad.price) }}</span>
        </span>
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
