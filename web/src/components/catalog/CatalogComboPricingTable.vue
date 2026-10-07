<script setup lang="ts">
import type { Combo, SquadProduct } from '@/api/types'
import { usePreferencesStore } from '@/stores/preferences'
import { comboDisplayPrice } from './comboDisplayPrice'
import ComboOption from './ComboOption.vue'

const props = defineProps<{
  combos: readonly Combo[]
  selectedId: string | null
  selectedSquads?: readonly SquadProduct[]
}>()

const emit = defineEmits<{ select: [id: string] }>()
const preferences = usePreferencesStore()
</script>

<template>
  <div class="combo-pricing">
    <div class="combo-grid">
      <ComboOption
        v-for="combo in props.combos"
        :key="combo.id"
        :combo="combo"
        :selected="combo.id === props.selectedId"
        :display-price="comboDisplayPrice(combo, props.selectedSquads ?? [], preferences.includeNodePrices)"
        @select="emit('select', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
.combo-grid {
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 20rem), 1fr));
  align-items: start;
}
</style>
