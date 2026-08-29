<script setup lang="ts">
import { computed } from 'vue'

import type { SquadProduct } from '@/api/types'
import SquadProfileSummary from '@/components/squad-profile/SquadProfileSummary.vue'
import { formatMoney } from '@/utils/format'
import { selectionHaptic } from '@/utils/telegram'
import SquadNodeBlocks from './SquadNodeBlocks.vue'

const props = defineProps<{
  squads: readonly SquadProduct[]
  selectedIds: readonly string[]
  includedIds: readonly string[]
  featuredIds: readonly string[]
  orderedIds: readonly string[]
  hiddenIds?: readonly string[]
}>()

const emit = defineEmits<{
  toggle: [id: string]
  openGeocheck: [node: SquadProduct['accessibleNodes'][number]]
}>()

function isSelected(id: string): boolean {
  return props.selectedIds.includes(id)
}

function isIncluded(id: string): boolean {
  return props.includedIds.includes(id)
}

function isFeatured(id: string): boolean {
  return props.featuredIds.includes(id)
}

function isFullPaidAddon(squad: SquadProduct): boolean {
  return !isIncluded(squad.id) && squad.stockRemaining === 0 && !squad.stockHeldByCurrentUser
}

function profileClass(squad: SquadProduct): string {
  return squad.profile ? `squad-option--${squad.profile.type}` : ''
}

function toggleSquad(squad: SquadProduct): void {
  if (isIncluded(squad.id) || isFullPaidAddon(squad)) return
  selectionHaptic()
  emit('toggle', squad.id)
}

function onCardKeydown(event: globalThis.KeyboardEvent, squad: SquadProduct): void {
  if (event.target !== event.currentTarget || (event.key !== 'Enter' && event.key !== ' ')) return
  event.preventDefault()
  toggleSquad(squad)
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

const squadOrder = computed(() => new Map(props.orderedIds.map((id, index) => [id, index])))
const squadRows = computed(() => props.squads
  .filter((squad) => !props.hiddenIds?.includes(squad.id))
  .map((squad, index) => ({ squad, index, occupancyPercentage: occupancyPercentage(squad) }))
  .sort((left, right) => {
    const leftOrder = squadOrder.value.get(left.squad.id) ?? props.orderedIds.length + left.index
    const rightOrder = squadOrder.value.get(right.squad.id) ?? props.orderedIds.length + right.index
    return leftOrder - rightOrder
  }))
</script>

<template>
  <section class="squad-selector">
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('catalog.optionalSquads') }}</h2>
      <p>{{ $t('catalog.optionalSquadsHint') }}</p>
    </div>
    <div v-if="squadRows.length" v-auto-animate class="squad-grid">
      <article v-for="row in squadRows" :key="row.squad.id" class="squad-option" :class="[profileClass(row.squad), { 'squad-option--featured': isFeatured(row.squad.id), 'squad-option--selected': isSelected(row.squad.id), 'squad-option--included': isIncluded(row.squad.id), 'squad-option--full': isFullPaidAddon(row.squad) }]" :tabindex="isIncluded(row.squad.id) || isFullPaidAddon(row.squad) ? undefined : 0" @click="toggleSquad(row.squad)" @keydown="onCardKeydown($event, row.squad)">
        <div class="squad-option__copy">
          <SquadProfileSummary :name="row.squad.name" :profile="row.squad.profile" :description="row.squad.description" presentation="member" compact>
            <template v-if="isIncluded(row.squad.id) || isFeatured(row.squad.id) || isFullPaidAddon(row.squad)" #nameTags>
              <span v-if="isIncluded(row.squad.id)" class="squad-option__included-tag">{{ $t('catalog.included') }}</span>
              <span v-if="isFeatured(row.squad.id)" class="squad-option__featured-tag">{{ $t('catalog.featured') }}</span>
              <span v-if="isFullPaidAddon(row.squad)" class="squad-option__full-tag">{{ $t('catalog.full') }}</span>
            </template>
            <template v-if="row.occupancyPercentage !== null" #headingMeta>
              <span class="squad-option__occupancy"><UIcon name="i-ph-users-three" />{{ row.occupancyPercentage }}%</span>
            </template>
            <template v-if="row.squad.activationRequired" #facts>
              <span class="squad-option__activation-tag"><UIcon name="i-ph-lock-key" />{{ $t('catalog.activationRequired') }}</span>
            </template>
          </SquadProfileSummary>
          <SquadNodeBlocks :nodes="row.squad.accessibleNodes" @open-geocheck="emit('openGeocheck', $event)" />
          <span v-if="!isIncluded(row.squad.id)">{{ formatMoney(row.squad.price) }}</span>
        </div>
        <UCheckbox
          :model-value="isSelected(row.squad.id)"
          :disabled="isIncluded(row.squad.id) || isFullPaidAddon(row.squad)"
          :aria-label="row.squad.name"
          @click.stop
          @update:model-value="toggleSquad(row.squad)"
        />
      </article>
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
.squad-option__included-tag, .squad-option__featured-tag, .squad-option__full-tag { display: inline-flex; flex: 0 0 auto; align-items: center; padding: 0.16rem 0.3rem; border-radius: 4px; color: var(--canvas); font-size: 0.54rem; font-weight: 800; line-height: 1; white-space: nowrap; }
.squad-option__included-tag { background: var(--success); }
.squad-option__featured-tag { background: var(--warning); }
.squad-option__full-tag { background: var(--danger); }
.squad-option__occupancy { display: inline-flex; flex: 0 0 auto; align-items: center; gap: 0.22rem; color: var(--text-muted); font-family: var(--font-mono); font-size: 0.68rem; }
</style>
