<script setup lang="ts">
import { onScopeDispose, shallowRef, watch } from 'vue'

import type { DrawPresenterProps } from '../selection'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const chosen = shallowRef<number | null>(null)
let autoChoose: ReturnType<typeof globalThis.setTimeout> | undefined
let finish: ReturnType<typeof globalThis.setTimeout> | undefined

function choose(index: number): void {
  if (!props.result || chosen.value !== null) return
  chosen.value = index
  if (autoChoose) globalThis.clearTimeout(autoChoose)
  finish = globalThis.setTimeout(() => emit('finished'), 1300)
}

watch(() => props.result, (value) => {
  if (value && chosen.value === null) autoChoose = globalThis.setTimeout(() => choose(1), 3000)
}, { immediate: true })
onScopeDispose(() => {
  if (autoChoose) globalThis.clearTimeout(autoChoose)
  if (finish) globalThis.clearTimeout(finish)
})
</script>

<template>
  <div class="draw-cards" role="group" :aria-label="$t('activity.drawStyle.cards')">
    <div class="draw-cards__deck">
      <UButton
        v-for="index in 3"
        :key="index"
        class="draw-cards__card"
        color="neutral"
        variant="outline"
        :disabled="!result || chosen !== null"
        :aria-label="$t('activity.chooseCard', { number: index })"
        @click="choose(index)"
      >
        <span class="draw-cards__inner" :class="{ 'draw-cards__inner--revealed': chosen === index }">
          <span class="draw-cards__front"><UIcon name="i-ph-sparkle" aria-hidden="true" /><small>{{ index }}</small></span>
          <strong class="draw-cards__back">{{ result?.prizeName }}</strong>
        </span>
      </UButton>
    </div>
    <p role="status">{{ $t(chosen !== null ? 'activity.cardRevealingHint' : result ? 'activity.cardReadyHint' : 'activity.drawing') }}</p>
  </div>
</template>

<style scoped>
.draw-cards { display: grid; justify-items: center; align-content: center; gap: 0.85rem; min-height: 15rem; }
.draw-cards__deck { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.5rem; width: min(100%, 19rem); perspective: 40rem; }
.draw-cards__card { position: relative; display: block; width: 100%; height: 8.4rem; padding: 0; border-radius: var(--radius-control); background: var(--surface-raised); }
.draw-cards__card:disabled { opacity: 1; }
.draw-cards__inner { position: absolute; inset: 0; display: block; transform-style: preserve-3d; transition: transform 760ms cubic-bezier(0.16, 1, 0.3, 1); }
.draw-cards__inner--revealed { transform: rotateY(180deg) translateY(0.25rem); }
.draw-cards__front, .draw-cards__back { position: absolute; inset: 0; display: grid; place-items: center; align-content: center; gap: 0.5rem; padding: 0.4rem; backface-visibility: hidden; }
.draw-cards__front { color: var(--accent); font-size: 1.8rem; }
.draw-cards__front small { color: var(--text-muted); font-size: 0.7rem; }
.draw-cards__back { color: var(--text); font-size: 0.82rem; line-height: 1.25; text-align: center; overflow-wrap: anywhere; transform: rotateY(180deg); }
.draw-cards p { margin: 0; color: var(--text-muted); font-size: 0.78rem; text-align: center; }
@media (prefers-reduced-motion: reduce) { .draw-cards__inner { transition: none; } }
</style>
