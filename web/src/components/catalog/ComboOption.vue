<script setup lang="ts">
import type { Combo } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SquadProfileSummary from '@/components/squad-profile/SquadProfileSummary.vue'
import { formatBytes, formatMoney } from '@/utils/format'
import SquadNodeBlocks from './SquadNodeBlocks.vue'

defineProps<{
  combo: Combo
  selected: boolean
}>()

defineEmits<{ select: [id: string] }>()
</script>

<template>
  <UButton
    class="combo-option"
    :class="{ 'combo-option--selected': selected }"
    color="neutral"
    variant="ghost"
    :aria-pressed="selected"
    data-haptic
    @click="$emit('select', combo.id)"
  >
    <span class="combo-option__top">
      <span>
        <StatusBadge v-if="selected" tone="success" :label="$t('catalog.selectedBadge')" />
        <strong>{{ combo.name }}</strong>
      </span>
      <span v-if="selected" class="selection-mark"><UIcon name="i-ph-check-bold" /></span>
    </span>
    <MarkdownContent class="combo-option__description" :source="combo.description" compact />
    <span class="combo-option__metrics">
      <span><UIcon name="i-ph-gauge" />{{ formatBytes(combo.trafficLimitBytes) }}</span>
      <span><UIcon name="i-ph-clock" />{{ $t('catalog.days', { count: combo.validityDays }) }}</span>
      <span><UIcon name="i-ph-stack" />{{ $t('catalog.squads', { count: combo.includedSquads.length }) }}</span>
      <span><UIcon name="i-ph-arrow-clockwise" />{{ $t(`home.reset.${combo.resetStrategy}`) }}</span>
      <span><UIcon name="i-ph-chart-line-up" />{{ $t('catalog.rolloverThreshold', { threshold: (combo.rolloverMinRemainingBps / 100).toFixed(2) }) }}</span>
    </span>
    <div v-if="combo.includedSquads.length" class="combo-option__squads">
      <div v-for="squad in combo.includedSquads" :key="squad.id">
        <SquadProfileSummary :name="squad.name" :profile="squad.profile" :description="squad.description" presentation="member" compact />
        <SquadNodeBlocks :nodes="squad.accessibleNodes" />
      </div>
    </div>
    <span class="combo-option__price">
      <strong>{{ formatMoney(combo.price) }}</strong>
      <small>{{ $t('catalog.perTerm') }}</small>
    </span>
  </UButton>
</template>
