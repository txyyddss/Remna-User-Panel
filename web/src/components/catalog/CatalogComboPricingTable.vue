<script setup lang="ts">
import { computed } from 'vue'

import type { Combo } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import { useI18n } from '@/i18n'
import { formatBytes, formatMoney } from '@/utils/format'
import { selectionHaptic } from '@/utils/telegram'

const props = defineProps<{
  combos: readonly Combo[]
  selectedId: string | null
}>()

const emit = defineEmits<{ select: [id: string] }>()
const { t } = useI18n()

const tiers = computed(() => props.combos.map((combo) => ({
  id: combo.id,
  title: combo.name,
  description: combo.description,
  price: formatMoney(combo.price),
  billingCycle: t('catalog.perDays', { count: combo.validityDays }),
  highlight: combo.id === props.selectedId,
})))

const sections = computed(() => [{
  title: t('catalog.planDetails'),
  features: [
    { id: 'traffic', title: t('catalog.trafficTerm'), tiers: values((combo) => `${formatBytes(combo.trafficLimitBytes)}/${resetUnit(combo)}`) },
    { id: 'squads', title: t('catalog.includedSquads'), tiers: values((combo) => t('catalog.includedSquadCount', { count: combo.includedSquads.length })) },
    { id: 'rollover', title: t('catalog.rollover'), tiers: values((combo) => t('catalog.rolloverThreshold', { threshold: (combo.rolloverMinRemainingBps / 100).toFixed(2) })) },
  ],
}])

function values(value: (combo: Combo) => string): Record<string, string> {
  return Object.fromEntries(props.combos.map((combo) => [combo.id, value(combo)]))
}

function comboFor(id: string): Combo | undefined {
  return props.combos.find((combo) => combo.id === id)
}

function resetUnit(combo: Combo): string {
  const key = combo.resetStrategy === 'DAY' ? 'day' : combo.resetStrategy === 'WEEK' ? 'week' : 'month'
  return t(`catalog.resetUnit.${key}`)
}

function selectCombo(id: string): void {
  if (id === props.selectedId) return
  selectionHaptic()
  emit('select', id)
}
</script>

<template>
  <UPricingTable :tiers="tiers" :sections="sections" :caption="$t('catalog.coreCombos')" class="combo-pricing-table">
    <template #tier-description="{ tier }">
      <MarkdownContent v-if="comboFor(tier.id)" class="combo-pricing-table__description" :source="comboFor(tier.id)?.description ?? ''" compact />
    </template>
    <template #tier-button="{ tier }">
      <UButton
        block
        size="lg"
        :color="tier.id === selectedId ? 'success' : 'neutral'"
        :variant="tier.id === selectedId ? 'solid' : 'outline'"
        :label="$t(tier.id === selectedId ? 'catalog.selectedBadge' : 'catalog.selectCombo')"
        :aria-pressed="tier.id === selectedId"
        data-haptic="select"
        @click="selectCombo(tier.id)"
      />
    </template>
  </UPricingTable>
</template>

<style scoped>
.combo-pricing-table :deep([data-slot='tierWrapper']) { min-width: 12rem; }
.combo-pricing-table :deep([data-slot='tierDescription']) { color: var(--text-muted); }
.combo-pricing-table :deep([data-slot='tierBillingCycle']) { overflow: visible; text-overflow: clip; white-space: nowrap; }
.combo-pricing-table__description { margin: 0; }
</style>
