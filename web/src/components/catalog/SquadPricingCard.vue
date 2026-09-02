<script setup lang="ts">
import { computed } from 'vue'

import type { SquadProduct } from '@/api/types'
import SquadProfileFacts from '@/components/squad-profile/SquadProfileFacts.vue'
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
const remainingBadgeText = computed(() => props.squad.stockLimit === null || props.squad.stockLimit === undefined ? '∞' : remainingText.value)

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
      {
        'squad-pricing-card--selected': selected,
        'squad-pricing-card--included': included,
        'squad-pricing-card--nonselectable': included || isFull,
      },
    ]"
    :aria-label="squad.name"
    :aria-pressed="!included && !isFull ? selected : undefined"
    :role="!included && !isFull ? 'button' : undefined"
    :tabindex="!included && !isFull ? 0 : undefined"
    @click="toggle"
    @keydown.enter.prevent="toggle"
    @keydown.space.prevent="toggle"
  >
    <div class="squad-card__header">
      <SquadProfileSummary
        :name="squad.name"
        :profile="squad.profile"
        :description="squad.visible ? squad.description : undefined"
        compact
        presentation="member"
        :show-facts="false"
      >
        <template #namePrefix>
          <UIcon v-if="featured" class="squad-card__featured" name="i-ph-crown" :aria-label="$t('catalog.featured')" />
        </template>
        <template #nameTags>
          <UIcon v-if="!squad.visible" name="i-ph-lock-key" :aria-label="$t('catalog.hidden')" />
          <UBadge v-if="squad.activationRequired" color="warning" variant="subtle" :label="$t('catalog.activationRequired')" />
          <UBadge v-if="isFull" color="error" variant="subtle" :label="$t('catalog.full')" />
          <span v-if="!included && !isFull" class="squad-card__remaining" :aria-label="$t('catalog.remaining')">
            <UIcon name="i-ph-gauge" aria-hidden="true" />
            <strong>{{ remainingBadgeText }}</strong>
          </span>
        </template>
        <template #headingMeta>
          <UIcon v-if="selected" name="i-ph-check-bold" aria-hidden="true" />
        </template>
      </SquadProfileSummary>
      <div v-if="!included" class="squad-card__price">
        <strong>{{ formatMoney(squad.price) }}</strong>
      </div>
    </div>

    <div v-if="squad.accessibleNodes.length" class="squad-card__nodes" @click.stop>
      <div class="squad-card__nodes-heading">
        <span class="squad-card__nodes-label">{{ $t('catalog.nodes') }}</span>
        <SquadProfileFacts class="squad-card__tags" :profile="squad.profile" presentation="member" />
      </div>
      <SquadNodeBlocks :nodes="squad.accessibleNodes" @open-geocheck="emit('openGeocheck', $event)" />
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
  cursor: pointer;
}
.squad-pricing-card:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.squad-pricing-card--nonselectable { cursor: default; }
.squad-pricing-card--selected {
  border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 9%, var(--surface-strong));
}
.squad-pricing-card--included { opacity: 0.78; }
.squad-card__header { min-width: 0; display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 0.65rem; align-items: start; }
.squad-card__price { min-width: 0; padding-top: 0.05rem; text-align: right; }
.squad-card__price > strong { display: block; white-space: nowrap; font-family: var(--font-mono); font-size: 1.1rem; }
.squad-card__remaining { display: inline-flex; align-items: center; gap: 0.2rem; padding: 0.18rem 0.35rem; border: 1px solid var(--line); border-radius: 999px; color: var(--text-muted); font-size: 0.65rem; line-height: 1.2; }
.squad-card__remaining strong { color: var(--text); font-family: var(--font-mono); font-weight: 700; }
.squad-card__nodes {
  display: grid;
  gap: 0.35rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--line);
}
.squad-card__nodes-heading { min-width: 0; display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: start; gap: 0.55rem; }
.squad-card__tags { justify-content: flex-end; margin-top: -0.25rem; }
.squad-card__nodes-label {
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
</style>
