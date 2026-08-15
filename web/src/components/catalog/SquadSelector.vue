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

function isFullPaidAddon(squad: SquadProduct): boolean {
  return !isIncluded(squad.id) && squad.stockRemaining === 0
}

function profileClass(squad: SquadProduct): string {
  return squad.profile ? `squad-option--${squad.profile.type}` : ''
}
</script>

<template>
  <section class="squad-selector">
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('catalog.optionalSquads') }}</h2>
      <p>{{ $t('catalog.optionalSquadsHint') }}</p>
    </div>
    <div v-if="squads.length" v-auto-animate class="squad-grid">
      <label v-for="squad in squads" :key="squad.id" class="squad-option" :class="[profileClass(squad), { 'squad-option--selected': isSelected(squad.id), 'squad-option--included': isIncluded(squad.id), 'squad-option--full': isFullPaidAddon(squad) }]" :data-haptic="isIncluded(squad.id) || isFullPaidAddon(squad) ? undefined : 'light'">
        <div class="squad-option__copy">
          <SquadProfileSummary :name="squad.name" :profile="squad.profile" :description="squad.description" presentation="member" compact />
          <StatusBadge v-if="isIncluded(squad.id)" tone="neutral" :label="$t('catalog.included')" />
          <StatusBadge v-else-if="squad.activationRequired" tone="warning" :label="$t('catalog.activationRequired')" />
          <StatusBadge v-else-if="isFullPaidAddon(squad)" tone="neutral" :label="$t('catalog.full')" />
          <span v-else>{{ formatMoney(squad.price) }}</span>
        </div>
        <UCheckbox
          :model-value="isSelected(squad.id)"
          :disabled="isIncluded(squad.id) || isFullPaidAddon(squad)"
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
.squad-option--included { color: var(--text-faint); border-color: var(--squad-profile-tone-line, var(--line)); background: var(--squad-profile-tone-soft, var(--canvas-soft)); cursor: default; }
.squad-option.squad-option--full { border-color: var(--line); color: var(--text-faint); background: var(--surface); cursor: not-allowed; opacity: 0.58; }
</style>
