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
  billingCycle: t('catalog.perTerm'),
  highlight: combo.id === props.selectedId,
})))

const sections = computed(() => [{
  title: t('catalog.planDetails'),
  features: [
    { id: 'traffic', title: t('catalog.traffic'), tiers: values((combo) => formatBytes(combo.trafficLimitBytes)) },
    { id: 'validity', title: t('catalog.validity'), tiers: values((combo) => t('catalog.days', { count: combo.validityDays })) },
    { id: 'squads', title: t('catalog.squads'), tiers: values((combo) => t('catalog.squads', { count: combo.includedSquads.length })) },
    { id: 'reset', title: t('catalog.term'), tiers: values((combo) => t(`home.reset.${combo.resetStrategy}`)) },
    { id: 'rollover', title: t('catalog.rollover'), tiers: values((combo) => t('catalog.rolloverThreshold', { threshold: (combo.rolloverMinRemainingBps / 100).toFixed(2) })) },
  ],
}])

function values(value: (combo: Combo) => string): Record<string, string> {
  return Object.fromEntries(props.combos.map((combo) => [combo.id, value(combo)]))
}

function comboFor(id: string): Combo | undefined {
  return props.combos.find((combo) => combo.id === id)
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
    <template #tier-badge="{ tier }">
      <UBadge v-if="tier.id === selectedId" color="success" variant="subtle" :label="$t('catalog.selectedBadge')" />
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
.combo-pricing-table__description { margin: 0; }
</style>
