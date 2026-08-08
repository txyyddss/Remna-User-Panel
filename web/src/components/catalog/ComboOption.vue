<script setup lang="ts">
import { PhCheck, PhClock, PhGauge, PhStack } from '@phosphor-icons/vue'

import type { Combo } from '@/api/types'
import { formatBytes, formatMoney } from '@/utils/format'
import StatusBadge from '@/components/common/StatusBadge.vue'
import MarkdownContent from '@/components/common/MarkdownContent.vue'

defineProps<{
  combo: Combo
  selected: boolean
}>()

defineEmits<{ select: [id: string] }>()
</script>

<template>
  <button
    class="combo-option"
    :class="{ 'combo-option--selected': selected }"
    type="button"
    :aria-pressed="selected"
    @click="$emit('select', combo.id)"
  >
    <span class="combo-option__top">
      <span>
        <StatusBadge v-if="selected" tone="success" :label="$t('catalog.selectedBadge')" />
        <strong>{{ combo.name }}</strong>
      </span>
      <span v-if="selected" class="selection-mark"><PhCheck :size="17" weight="bold" /></span>
    </span>
    <MarkdownContent class="combo-option__description" :source="combo.description" compact />
    <span class="combo-option__metrics">
      <span><PhGauge :size="17" />{{ formatBytes(combo.trafficLimitBytes) }}</span>
      <span><PhClock :size="17" />{{ $t('catalog.days', { count: combo.validityDays }) }}</span>
      <span><PhStack :size="17" />{{ $t('catalog.squads', { count: combo.includedSquads.length }) }}</span>
    </span>
    <span class="combo-option__price">
      <strong>{{ formatMoney(combo.price) }}</strong>
      <small>{{ $t('catalog.perTerm') }}</small>
    </span>
  </button>
</template>
