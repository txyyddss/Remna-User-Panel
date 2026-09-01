<script setup lang="ts">
import type { Combo } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatBytes, formatMoney } from '@/utils/format'
import { selectionHaptic } from '@/utils/telegram'

const props = defineProps<{
  combo: Combo
  selected: boolean
}>()

const emit = defineEmits<{ select: [id: string] }>()
const { t } = useI18n()

function trafficTerm(): string {
  const key = props.combo.resetStrategy === 'DAY' ? 'day' : props.combo.resetStrategy === 'WEEK' ? 'week' : 'month'
  return `${formatBytes(props.combo.trafficLimitBytes)}/${t(`catalog.resetUnit.${key}`)}`
}

function selectCombo(): void {
  if (props.selected) return
  selectionHaptic()
  emit('select', props.combo.id)
}
</script>

<template>
  <UButton
    class="combo-option"
    :class="{ 'combo-option--selected': selected }"
    color="neutral"
    variant="ghost"
    :aria-pressed="selected"
    @click="selectCombo"
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
      <span><UIcon name="i-ph-gauge" />{{ trafficTerm() }}</span>
      <span><UIcon name="i-ph-stack" />{{ $t('catalog.includedSquadCount', { count: combo.includedSquads.length }) }}</span>
      <span><UIcon name="i-ph-chart-line-up" />{{ $t('catalog.rolloverThreshold', { threshold: (combo.rolloverMinRemainingBps / 100).toFixed(2) }) }}</span>
    </span>
    <span class="combo-option__price">
      <strong>{{ formatMoney(combo.price) }}</strong>
      <small>{{ $t('catalog.perDays', { count: combo.validityDays }) }}</small>
    </span>
  </UButton>
</template>
