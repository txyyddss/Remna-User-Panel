<script setup lang="ts">
import { nextTick, onScopeDispose, shallowRef, useTemplateRef, watch } from 'vue'

import type { DrawPresenterProps } from '../selection'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const canvas = useTemplateRef<globalThis.HTMLCanvasElement>('cover')
const ready = shallowRef(false)
const scratching = shallowRef(false)
let frame: number | undefined
let completed = false

function paint(): void {
  const element = canvas.value
  if (!element || !props.result) return
  element.width = 280
  element.height = 150
  const context = element.getContext('2d')
  if (!context) { emit('finished'); return }
  context.globalCompositeOperation = 'source-over'
  context.fillStyle = '#394039'
  context.fillRect(0, 0, element.width, element.height)
  context.globalCompositeOperation = 'destination-out'
  context.lineCap = 'round'
  context.lineWidth = 10
  ready.value = true
}

function startReveal(): void {
  const element = canvas.value
  if (!element || !props.result || scratching.value || completed) return
  const context = element.getContext('2d')
  if (!context) { emit('finished'); return }
  const cover = element
  const drawingContext = context
  scratching.value = true
  let startedAt: number | undefined
  function draw(now: number): void {
    if (startedAt === undefined) startedAt = now
    const progress = Math.min((now - startedAt) / 1600, 1)
    for (let row = 0; row < 12; row += 1) {
      const portion = Math.min(Math.max(progress * 12 - row, 0), 1)
      if (!portion) continue
      const y = 8 + row * 12
      drawingContext.beginPath()
      drawingContext.moveTo(row % 2 ? 280 : 0, y)
      drawingContext.lineTo(row % 2 ? 280 * (1 - portion) : 280 * portion, y)
      drawingContext.stroke()
    }
    if (progress < 1) { frame = globalThis.requestAnimationFrame(draw); return }
    drawingContext.clearRect(0, 0, cover.width, cover.height)
    completed = true
    emit('finished')
  }
  frame = globalThis.requestAnimationFrame(draw)
}

watch(() => props.result, async (value) => {
  if (!value) return
  ready.value = false
  await nextTick()
  paint()
}, { immediate: true })
watch(() => props.revealRequested, async (value) => {
  if (!value) return
  await nextTick()
  if (!ready.value) paint()
  startReveal()
})
onScopeDispose(() => { if (frame !== undefined) globalThis.cancelAnimationFrame(frame) })
</script>

<template>
  <div class="draw-scratch" role="group" :aria-label="$t('activity.drawStyle.scratch')">
    <UButton v-if="result" class="draw-scratch__surface" color="neutral" variant="ghost" :disabled="scratching" :aria-label="$t('activity.scratchHint')" @click="startReveal">
      <strong>{{ result.prizeName }}</strong>
      <span v-if="!ready" class="draw-scratch__placeholder" aria-hidden="true" />
      <canvas ref="cover" :class="{ 'draw-scratch__cover--ready': ready }" aria-hidden="true" />
    </UButton>
    <div v-else class="draw-scratch__waiting"><UIcon name="i-ph-ticket" aria-hidden="true" /></div>
    <p>{{ result ? $t('activity.scratchHint') : $t('activity.drawing') }}</p>
  </div>
</template>

<style scoped>
.draw-scratch { display: grid; place-items: center; align-content: center; gap: 0.6rem; min-height: 15rem; }
.draw-scratch__surface { position: relative; display: grid; place-items: center; width: min(100%, 17.5rem); height: 9.375rem; padding: 1rem; border: 1px solid var(--accent); border-radius: var(--radius-control); background: var(--accent-soft); font: inherit; text-align: center; overflow-wrap: anywhere; cursor: pointer; }
.draw-scratch__surface:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.draw-scratch__surface:disabled { cursor: default; }
.draw-scratch__surface strong { color: var(--text); font-size: 1.1rem; }
.draw-scratch__placeholder, .draw-scratch__surface canvas { position: absolute; inset: 0; width: 100%; height: 100%; border-radius: var(--radius-control); }
.draw-scratch__placeholder { background: var(--line-strong); }
.draw-scratch__surface canvas { visibility: hidden; pointer-events: none; }
.draw-scratch__surface .draw-scratch__cover--ready { visibility: visible; }
.draw-scratch__waiting { color: var(--accent); font-size: 4rem; }
.draw-scratch p { margin: 0; color: var(--text-muted); font-size: 0.78rem; text-align: center; }
</style>
