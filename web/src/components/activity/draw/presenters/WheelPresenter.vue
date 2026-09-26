<script setup lang="ts">
import { computed, onScopeDispose, shallowRef, useTemplateRef, watch } from 'vue'
import { gsap } from 'gsap'

import { selectedPreviewIndex, type DrawPresenterProps } from '../selection'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { usePreview } from './usePreview'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: []; started: [] }>()
const visible = usePreview(props)
const { reducedMotion } = useMotionPreferences()
const wheel = useTemplateRef<globalThis.SVGGElement>('wheel')
let loop: gsap.core.Tween | undefined
let stopping: gsap.core.Tween | undefined
let slowTimer: ReturnType<typeof globalThis.setTimeout> | undefined
let slowRequested = false
const started = shallowRef(false)
const slowing = shallowRef(false)

function point(angle: number, radius: number): [number, number] {
  const radians = (angle - 90) * Math.PI / 180
  return [150 + radius * Math.cos(radians), 150 + radius * Math.sin(radians)]
}
function segment(index: number, count: number): string {
  const sweep = 360 / count
  const [x1, y1] = point(index * sweep, 136)
  const [x2, y2] = point((index + 1) * sweep, 136)
  return 'M 150 150 L ' + x1 + ' ' + y1 + ' A 136 136 0 ' + (sweep > 180 ? 1 : 0) + ' 1 ' + x2 + ' ' + y2 + ' Z'
}
const sectors = computed(() => visible.value.map((prize, index) => {
  const [x, y] = point((index + 0.5) * 360 / visible.value.length, 89)
  const characters = Array.from(prize.name)
  return { ...prize, path: segment(index, visible.value.length), x, y, label: characters.slice(0, 9).join('') + (characters.length > 9 ? '…' : '') }
}))

function settle(): void {
  if (!wheel.value || slowing.value || !props.result) return
  slowing.value = true
  if (slowTimer) globalThis.clearTimeout(slowTimer)
  loop?.kill()
  const index = selectedPreviewIndex(visible.value, props.result)
  if (index < 0) { emit('finished'); return }
  const current = Number(gsap.getProperty(wheel.value, 'rotation')) || 0
  const angle = visible.value.length === 1 ? 0 : (index + 0.5) * 360 / visible.value.length
  const target = Math.ceil((current + 540 + angle) / 360) * 360 - angle
  stopping = gsap.to(wheel.value, {
    rotation: target, duration: 4.5, ease: 'power2.out', onComplete: () => emit('finished'),
  })
}
function scheduleSlowdown(): void {
  if (!started.value || !props.result || slowing.value || slowTimer) return
  if (slowRequested) { settle(); return }
  slowTimer = globalThis.setTimeout(() => { slowTimer = undefined; settle() }, 2600)
}
function pressCenter(): void {
  if (slowing.value || !wheel.value) return
  if (started.value) {
    slowRequested = true
    if (props.result) settle()
    else loop?.timeScale(0.7)
    return
  }
  started.value = true
  emit('started')
  if (reducedMotion.value) { if (props.result) emit('finished'); return }
  loop = gsap.to(wheel.value, { rotation: 360, duration: 3.5, repeat: -1, ease: 'none', transformOrigin: '50% 50%' })
  scheduleSlowdown()
}
watch(() => props.result, (value) => {
  if (!value) return
  if (started.value && reducedMotion.value) { emit('finished'); return }
  scheduleSlowdown()
})
watch(reducedMotion, (value) => { if (value) loop?.kill() })
onScopeDispose(() => {
  if (slowTimer) globalThis.clearTimeout(slowTimer)
  loop?.kill()
  stopping?.kill()
})
</script>

<template>
  <div class="draw-wheel" role="group" :aria-label="$t('activity.drawStyle.wheel')">
    <svg viewBox="0 0 300 300" aria-hidden="true">
      <g ref="wheel">
        <circle cx="150" cy="150" r="138" fill="var(--surface-raised)" stroke="var(--line-strong)" stroke-width="2" />
        <template v-if="sectors.length === 1">
          <circle cx="150" cy="150" r="136" fill="var(--accent-soft)" stroke="var(--line-strong)" stroke-width="1" />
          <text x="150" y="91" text-anchor="middle" fill="var(--text)" font-size="13">{{ sectors[0]?.label }}</text>
        </template>
        <g v-for="(sector, index) in sectors" v-else :key="sector.id">
          <path :d="sector.path" :fill="index % 2 ? 'var(--surface-raised)' : 'var(--accent-soft)'" stroke="var(--line-strong)" stroke-width="1" />
          <text :x="sector.x" :y="sector.y" text-anchor="middle" dominant-baseline="middle" fill="var(--text)" font-size="10">{{ sector.label }}</text>
        </g>
      </g>
      <path d="M 139 4 L 161 4 L 150 32 Z" fill="var(--accent)" />
    </svg>
    <UButton class="draw-wheel__start" color="neutral" variant="ghost" :disabled="slowing" :aria-label="$t(started ? 'activity.slowWheel' : 'activity.spinWheel')" @click="pressCenter">
      {{ $t(started ? 'activity.slowWheel' : 'activity.spinWheel') }}
    </UButton>
  </div>
</template>

<style scoped>
.draw-wheel { position: relative; display: grid; place-items: center; width: min(100%, 18rem); margin: 0 auto; min-height: 15rem; }
.draw-wheel svg { display: block; width: 100%; max-height: 18rem; }
.draw-wheel__start { position: absolute; top: 50%; left: 50%; display: grid; place-items: center; width: 4.5rem; height: 4.5rem; padding: 0.3rem; border: 2px solid var(--accent); border-radius: 50%; background: var(--canvas); color: var(--accent); font: inherit; font-size: 0.69rem; font-weight: 700; line-height: 1.15; text-align: center; cursor: pointer; transform: translate(-50%, -50%); }
.draw-wheel__start:focus-visible { outline: 2px solid var(--accent); outline-offset: 3px; }
.draw-wheel__start:disabled { cursor: default; opacity: 0.7; }
</style>
