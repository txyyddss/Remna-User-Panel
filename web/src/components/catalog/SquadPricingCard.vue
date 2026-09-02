<script setup lang="ts">
import { computed } from 'vue'

import type { SquadProduct } from '@/api/types'
import SquadProfileSummary from '@/components/squad-profile/SquadProfileSummary.vue'
import { useI18n } from '@/i18n'
import { formatMoney } from '@/utils/format'
import { selectionHaptic } from '@/utils/telegram'
import SquadNodeBlocks from './SquadNodeBlocks.vue'

const props = defineProps<{
  squad: SquadProduct
  selected: boolean
  included: boolean
  featured: boolean
}>()

const emit = defineEmits<{
  toggle: [id: string]
  openGeocheck: [node: SquadProduct['accessibleNodes'][number]]
}>()
const { t } = useI18n()

const isFull = computed(() => !props.included && props.squad.stockRemaining === 0 && !props.squad.stockHeldByCurrentUser)
const remainingText = computed(() => {
  if (props.squad.stockLimit === null || props.squad.stockLimit === undefined) return t('catalog.unlimitedStock')
  if (props.squad.stockLimit <= 0) return '0%'
  const count = Math.max(0, Math.min(props.squad.stockLimit, props.squad.stockRemaining ?? props.squad.stockLimit))
  return `${Math.round(count * 100 / props.squad.stockLimit)}%`
})

function toggle(): void {
  if (props.included || isFull.value) return
  selectionHaptic()
  emit('toggle', props.squad.id)
}
</script>

<template>
  <article
    class="squad-pricing-card"
    :class="[
      `squad-pricing-card--${squad.profile?.type ?? 'unconfigured'}`,
      { 'squad-pricing-card--selected': selected, 'squad-pricing-card--included': included },
    ]"
    :aria-label="squad.name"
  >
    <SquadProfileSummary
      :name="squad.name"
      :profile="squad.profile"
      :description="squad.description"
      compact
      presentation="member"
    >
      <template #nameTags>
        <UIcon v-if="featured" name="i-ph-crown" :aria-label="$t('catalog.featured')" />
        <UBadge v-if="squad.activationRequired" color="warning" variant="subtle" :label="$t('catalog.activationRequired')" />
        <UBadge v-if="isFull" color="error" variant="subtle" :label="$t('catalog.full')" />
      </template>
      <template #headingMeta>
        <UIcon v-if="selected" name="i-ph-check-bold" aria-hidden="true" />
      </template>
      <template #facts>
        <span class="squad-card__fact">
          <UIcon name="i-ph-stack" aria-hidden="true" />
          <span>{{ $t('catalog.nodeCount') }}</span>
          <strong>{{ squad.accessibleNodes.length }}</strong>
        </span>
        <span class="squad-card__fact">
          <UIcon name="i-ph-gauge" aria-hidden="true" />
          <span>{{ $t('catalog.remaining') }}</span>
          <strong>{{ remainingText }}</strong>
        </span>
      </template>
    </SquadProfileSummary>

    <div v-if="squad.accessibleNodes.length" class="squad-card__nodes">
      <span class="squad-card__nodes-label">{{ $t('catalog.nodes') }}</span>
      <SquadNodeBlocks :nodes="squad.accessibleNodes" @open-geocheck="emit('openGeocheck', $event)" />
    </div>

    <div v-if="!included" class="squad-card__purchase">
      <strong>{{ formatMoney(squad.price) }}</strong>
      <UButton
        type="button"
        block
        size="lg"
        :disabled="isFull"
        :color="selected ? 'success' : 'neutral'"
        :variant="selected ? 'solid' : 'outline'"
        :label="$t(isFull ? 'catalog.full' : selected ? 'catalog.selectedBadge' : 'catalog.selectSquad')"
        :aria-pressed="selected"
        data-haptic="select"
        @click="toggle"
      />
    </div>
  </article>
</template>

<style scoped>
.squad-pricing-card {
  min-width: 0;
  display: grid;
  gap: 0.7rem;
  padding: 0.8rem;
  border: 1px solid var(--line);
  border-radius: var(--radius-control);
  background: color-mix(in srgb, var(--surface-strong) 78%, transparent);
}
.squad-pricing-card--selected {
  border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 9%, var(--surface-strong));
}
.squad-pricing-card--included { opacity: 0.78; }
.squad-card__fact {
  flex: 1 1 calc(50% - 0.35rem);
  max-width: 100%;
  overflow-wrap: anywhere;
}
.squad-card__fact > span { min-width: 0; }
.squad-card__fact > strong {
  margin-left: auto;
  color: var(--text);
  font-family: var(--font-mono);
  white-space: nowrap;
}
.squad-card__nodes {
  display: grid;
  gap: 0.35rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--line);
}
.squad-card__nodes-label {
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.squad-card__purchase { display: grid; gap: 0.6rem; }
.squad-card__purchase > strong { font-family: var(--font-mono); font-size: 1.1rem; }
@media (max-width: 380px) {
  .squad-card__fact { flex-basis: 100%; }
}
</style>
