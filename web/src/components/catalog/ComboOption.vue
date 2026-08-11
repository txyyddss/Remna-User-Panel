<script setup lang="ts">
import type { Combo } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { formatBytes, formatMoney } from '@/utils/format'

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
    </span>
    <span class="combo-option__price">
      <strong>{{ formatMoney(combo.price) }}</strong>
      <small>{{ $t('catalog.perTerm') }}</small>
    </span>
  </UButton>
</template>
