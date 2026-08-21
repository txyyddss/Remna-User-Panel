<script setup lang="ts">
import { computed } from 'vue'

import type { SquadProduct } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SquadProfileSummary from '@/components/squad-profile/SquadProfileSummary.vue'
import { formatMoney } from '@/utils/format'
import SquadNodeBlocks from './SquadNodeBlocks.vue'

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

function occupancyPercentage(squad: SquadProduct): number | null {
  if (squad.stockLimit === null || squad.stockLimit === undefined) return null
  if (squad.stockLimit <= 0) return 100
  const remaining = Math.max(0, Math.min(squad.stockLimit, squad.stockRemaining ?? squad.stockLimit))
  const occupied = squad.stockLimit - remaining
  if (occupied <= 0) return 0
  if (remaining <= 0) return 100
  return Math.max(1, Math.min(99, Math.round(occupied * 100 / squad.stockLimit)))
}

const squadRows = computed(() => props.squads.map((squad) => ({
  squad,
  occupancyPercentage: occupancyPercentage(squad),
})))
</script>

<template>
  <section class="squad-selector">
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('catalog.optionalSquads') }}</h2>
      <p>{{ $t('catalog.optionalSquadsHint') }}</p>
    </div>
    <div v-if="squads.length" v-auto-animate class="squad-grid">
      <label v-for="row in squadRows" :key="row.squad.id" class="squad-option" :class="[profileClass(row.squad), { 'squad-option--selected': isSelected(row.squad.id), 'squad-option--included': isIncluded(row.squad.id), 'squad-option--full': isFullPaidAddon(row.squad) }]" :data-haptic="isIncluded(row.squad.id) || isFullPaidAddon(row.squad) ? undefined : 'light'">
        <div class="squad-option__copy">
          <SquadProfileSummary :name="row.squad.name" :profile="row.squad.profile" :description="row.squad.description" presentation="member" compact>
            <template v-if="row.occupancyPercentage !== null" #facts>
              <span><UIcon name="i-ph-users-three" />{{ row.occupancyPercentage }}%</span>
            </template>
          </SquadProfileSummary>
          <SquadNodeBlocks :nodes="row.squad.accessibleNodes" />
          <StatusBadge v-if="isIncluded(row.squad.id)" tone="neutral" :label="$t('catalog.included')" />
          <StatusBadge v-else-if="row.squad.activationRequired" tone="warning" :label="$t('catalog.activationRequired')" />
          <StatusBadge v-else-if="isFullPaidAddon(row.squad)" tone="neutral" :label="$t('catalog.full')" />
          <span v-else>{{ formatMoney(row.squad.price) }}</span>
        </div>
        <UCheckbox
          :model-value="isSelected(row.squad.id)"
          :disabled="isIncluded(row.squad.id) || isFullPaidAddon(row.squad)"
          :aria-label="row.squad.name"
          @update:model-value="emit('toggle', row.squad.id)"
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
