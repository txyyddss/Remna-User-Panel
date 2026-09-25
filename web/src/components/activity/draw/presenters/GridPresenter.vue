<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import type { DrawPresenterProps } from '../selection'
import { useSelectionTicker } from './useSelectionTicker'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const scope = useTemplateRef<globalThis.HTMLElement>('scope')
const { visible, active } = useSelectionTicker(props, scope, () => emit('finished'))
const positions = [
  [0, 0], [1, 0], [2, 0], [2, 1], [2, 2], [1, 2], [0, 2], [0, 1],
] as const
const tiles = computed(() => visible.value.map((prize, index) => {
  const characters = Array.from(prize.name)
  return { ...prize, x: positions[index]![0] * 91 + 8, y: positions[index]![1] * 91 + 8, label: characters.slice(0, 8).join('') + (characters.length > 8 ? '…' : '') }
}))
</script>

<template>
  <div ref="scope" class="draw-grid" role="img" :aria-label="$t('activity.drawStyle.grid')">
    <svg viewBox="0 0 290 290" aria-hidden="true">
      <g v-for="(tile, index) in tiles" :key="tile.id">
        <rect :x="tile.x" :y="tile.y" width="84" height="84" rx="10" :fill="index === active ? 'var(--accent-soft)' : 'var(--surface-raised)'" :stroke="index === active ? 'var(--accent)' : 'var(--line-strong)'" stroke-width="2" />
        <text :x="tile.x + 42" :y="tile.y + 42" text-anchor="middle" dominant-baseline="middle" fill="var(--text)" font-size="11">{{ tile.label }}</text>
      </g>
      <rect x="99" y="99" width="84" height="84" rx="10" fill="var(--canvas)" stroke="var(--line-strong)" />
      <path d="M 141 116 L 151 136 L 172 145 L 151 153 L 141 174 L 133 153 L 112 145 L 133 136 Z" fill="var(--accent)" />
    </svg>
  </div>
</template>

<style scoped>
.draw-grid { display: grid; place-items: center; width: min(100%, 18rem); margin: 0 auto; min-height: 15rem; }
.draw-grid svg { display: block; width: 100%; max-height: 18rem; }
</style>
