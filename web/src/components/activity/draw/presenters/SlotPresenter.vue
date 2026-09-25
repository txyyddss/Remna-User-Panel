<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import type { DrawPresenterProps } from '../selection'
import { useSelectionTicker } from './useSelectionTicker'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const scope = useTemplateRef<globalThis.HTMLElement>('scope')
const { visible, active } = useSelectionTicker(props, scope, () => emit('finished'))
const rows = computed(() => {
  const count = visible.value.length
  if (!count) return []
  return [-1, 0, 1].map((delta) => {
    const prize = visible.value[(active.value + delta + count) % count]!
    const characters = Array.from(prize.name)
    return { id: delta, y: 62 + (delta + 1) * 58, label: characters.slice(0, 18).join('') + (characters.length > 18 ? '…' : '') }
  })
})
</script>

<template>
  <div ref="scope" class="draw-slot" role="img" :aria-label="$t('activity.drawStyle.slot')">
    <svg viewBox="0 0 300 250" aria-hidden="true">
      <rect x="10" y="10" width="280" height="230" rx="16" fill="var(--surface-raised)" stroke="var(--line-strong)" stroke-width="2" />
      <rect x="20" y="98" width="260" height="54" rx="8" fill="var(--accent-soft)" stroke="var(--accent)" stroke-width="2" />
      <text v-for="row in rows" :key="row.id" x="150" :y="row.y" text-anchor="middle" dominant-baseline="middle" :fill="row.id === 0 ? 'var(--text)' : 'var(--text-muted)'" :font-size="row.id === 0 ? 17 : 13">{{ row.label }}</text>
    </svg>
  </div>
</template>

<style scoped>
.draw-slot { display: grid; place-items: center; width: min(100%, 19rem); margin: 0 auto; min-height: 15rem; }
.draw-slot svg { display: block; width: 100%; }
</style>
