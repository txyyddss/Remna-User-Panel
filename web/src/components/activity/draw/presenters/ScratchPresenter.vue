<script setup lang="ts">
import { nextTick, onUnmounted, shallowRef, useTemplateRef, watch } from 'vue'

import type { DrawPresenterProps } from '../selection'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const canvas = useTemplateRef<globalThis.HTMLCanvasElement>('cover')
const ready = shallowRef(false)
let drawing = false
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
  context.lineJoin = 'round'
  context.lineWidth = 30
  ready.value = true
}

function position(event: globalThis.PointerEvent): [number, number] {
  const element = canvas.value!
  const bounds = element.getBoundingClientRect()
  return [(event.clientX - bounds.left) * element.width / bounds.width, (event.clientY - bounds.top) * element.height / bounds.height]
}

function erase(event: globalThis.PointerEvent): void {
  if (!drawing || completed || !props.result) return
  const context = canvas.value?.getContext('2d')
  if (!context) return
  const [x, y] = position(event)
  context.lineTo(x, y)
  context.stroke()
}

function down(event: globalThis.PointerEvent): void {
  if (!canvas.value || !props.result) return
  drawing = true
  canvas.value.setPointerCapture(event.pointerId)
  const context = canvas.value.getContext('2d')
  if (!context) return
  const [x, y] = position(event)
  context.beginPath()
  context.moveTo(x, y)
  context.lineTo(x + 0.1, y + 0.1)
  context.stroke()
}

function up(): void {
  if (!drawing) return
  drawing = false
  const context = canvas.value?.getContext('2d')
  if (!context || !canvas.value) return
  const pixels = context.getImageData(0, 0, canvas.value.width, canvas.value.height).data
  let cleared = 0
  let sampled = 0
  for (let index = 3; index < pixels.length; index += 64) {
    sampled += 1
    if (pixels[index]! < 30) cleared += 1
  }
  if (cleared / sampled >= 0.45) {
    completed = true
    emit('finished')
  }
}

watch(() => props.result, async (value) => {
  if (!value) return
  ready.value = false
  await nextTick()
  paint()
}, { immediate: true })
onUnmounted(() => { drawing = false })
</script>

<template>
  <div class="draw-scratch" role="group" :aria-label="$t('activity.drawStyle.scratch')">
    <div v-if="result" class="draw-scratch__surface">
      <strong>{{ result.prizeName }}</strong>
      <div v-if="!ready" class="draw-scratch__placeholder" aria-hidden="true" />
      <canvas ref="cover" :class="{ 'draw-scratch__cover--ready': ready }" :aria-label="$t('activity.scratchHint')" @pointerdown="down" @pointermove="erase" @pointerup="up" @pointercancel="up" />
    </div>
    <div v-else class="draw-scratch__waiting"><UIcon name="i-ph-ticket" aria-hidden="true" /></div>
    <p>{{ result ? $t('activity.scratchHint') : $t('activity.drawing') }}</p>
  </div>
</template>

<style scoped>
.draw-scratch { display: grid; place-items: center; align-content: center; gap: 0.6rem; min-height: 15rem; }
.draw-scratch__surface { position: relative; display: grid; place-items: center; width: min(100%, 17.5rem); height: 9.375rem; padding: 1rem; border: 1px solid var(--accent); border-radius: var(--radius-control); background: var(--accent-soft); text-align: center; overflow-wrap: anywhere; }
.draw-scratch__surface strong { color: var(--text); font-size: 1.1rem; }
.draw-scratch__placeholder, .draw-scratch__surface canvas { position: absolute; inset: 0; width: 100%; height: 100%; border-radius: var(--radius-control); }
.draw-scratch__placeholder { background: var(--line-strong); }
.draw-scratch__surface canvas { visibility: hidden; touch-action: none; cursor: crosshair; }
.draw-scratch__surface .draw-scratch__cover--ready { visibility: visible; }
.draw-scratch__waiting { color: var(--accent); font-size: 4rem; }
.draw-scratch p { margin: 0; color: var(--text-muted); font-size: 0.78rem; text-align: center; }
</style>
