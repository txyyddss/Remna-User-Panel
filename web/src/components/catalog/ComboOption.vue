<script setup lang="ts">
import { PhCheck, PhClock, PhGauge, PhStack } from '@phosphor-icons/vue'

import type { Combo } from '@/api/types'
import { formatBytes, formatMoney } from '@/utils/format'
import StatusBadge from '@/components/common/StatusBadge.vue'

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
        <StatusBadge v-if="selected" tone="success" label="Selected" />
        <strong>{{ combo.name }}</strong>
      </span>
      <span v-if="selected" class="selection-mark"><PhCheck :size="17" weight="bold" /></span>
    </span>
    <span class="combo-option__description">{{ combo.description }}</span>
    <span class="combo-option__metrics">
      <span><PhGauge :size="17" />{{ formatBytes(combo.trafficLimitBytes) }}</span>
      <span><PhClock :size="17" />{{ combo.validityDays }} days</span>
      <span><PhStack :size="17" />{{ combo.includedSquads.length }} squads</span>
    </span>
    <span class="combo-option__price">
      <strong>{{ formatMoney(combo.price) }}</strong>
      <small>per term</small>
    </span>
  </button>
</template>
