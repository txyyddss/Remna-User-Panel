<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import type { DrawPresenterProps } from '../selection'
import { useSelectionTicker } from './useSelectionTicker'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: []; started: [] }>()
const scope = useTemplateRef<globalThis.HTMLElement>('scope')
const { visible, active, running, start } = useSelectionTicker(props, scope, () => emit('started'), () => emit('finished'))
const positions = [
  [0, 0], [1, 0], [2, 0], [2, 1], [2, 2], [1, 2], [0, 2], [0, 1],
] as const
const tiles = computed(() => visible.value.map((prize, index) => {
  const characters = Array.from(prize.name)
  return { ...prize, x: positions[index]![0] * 91 + 8, y: positions[index]![1] * 91 + 8, label: characters.slice(0, 8).join('') + (characters.length > 8 ? '…' : '') }
}))
</script>

<template>
  <div ref="scope" class="draw-grid" role="group" :aria-label="$t('activity.drawStyle.grid')">
    <svg viewBox="0 0 290 290" aria-hidden="true">
      <g v-for="(tile, index) in tiles" :key="tile.id">
        <rect :x="tile.x" :y="tile.y" width="84" height="84" rx="10" :fill="index === active ? 'var(--accent-soft)' : 'var(--surface-raised)'" :stroke="index === active ? 'var(--accent)' : 'var(--line-strong)'" stroke-width="2" />
        <text :x="tile.x + 42" :y="tile.y + 42" text-anchor="middle" dominant-baseline="middle" fill="var(--text)" font-size="11">{{ tile.label }}</text>
      </g>
    </svg>
    <UButton class="draw-grid__start" color="neutral" variant="ghost" :disabled="running" :aria-label="$t('activity.startDraw')" @click="start">
      <UIcon name="i-ph-play-fill" aria-hidden="true" />
      <span>{{ $t('activity.startDraw') }}</span>
    </UButton>
  </div>
</template>

<style scoped>
.draw-grid { position: relative; display: grid; place-items: center; width: min(100%, 18rem); margin: 0 auto; min-height: 15rem; }
.draw-grid svg { display: block; width: 100%; max-height: 18rem; }
.draw-grid__start { position: absolute; top: 50%; left: 50%; display: grid; place-items: center; align-content: center; gap: 0.12rem; width: 29%; height: 29%; padding: 0.2rem; border: 1px solid var(--accent); border-radius: 0.6rem; background: var(--canvas); color: var(--accent); font: inherit; font-size: 0.65rem; cursor: pointer; transform: translate(-50%, -50%); }
.draw-grid__start :deep(svg) { width: 1.1rem; height: 1.1rem; }
.draw-grid__start:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.draw-grid__start:disabled { opacity: 0.65; cursor: default; }
</style>
