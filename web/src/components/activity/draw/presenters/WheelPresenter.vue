<script setup lang="ts">
import { computed, onMounted, onScopeDispose, useTemplateRef, watch } from 'vue'
import { gsap } from 'gsap'

import { selectedPreviewIndex, type DrawPresenterProps } from '../selection'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { usePreview } from './usePreview'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const visible = usePreview(props)
const { reducedMotion } = useMotionPreferences()
const wheel = useTemplateRef<globalThis.SVGGElement>('wheel')
let context: gsap.Context | undefined
let loop: gsap.core.Tween | undefined

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
  if (!wheel.value || !context) return
  loop?.kill()
  const index = selectedPreviewIndex(visible.value, props.result)
  if (index < 0) { emit('finished'); return }
  const current = Number(gsap.getProperty(wheel.value, 'rotation')) || 0
  const target = Math.ceil(current / 360) * 360 + 720 - (visible.value.length === 1 ? 0 : (index + 0.5) * 360 / visible.value.length)
  context.add(() => gsap.to(wheel.value, {
    rotation: target, duration: 2.4, ease: 'power4.out', onComplete: () => emit('finished'),
  }))
}
onMounted(() => {
  context = gsap.context(() => undefined, wheel.value ?? undefined)
  if (props.result) { settle(); return }
  if (reducedMotion.value) return
  context.add(() => {
    loop = gsap.to(wheel.value, { rotation: 360, duration: 2.5, repeat: -1, ease: 'none', transformOrigin: '50% 50%' })
  })
})
watch(() => props.result, (value) => { if (value) settle() })
watch(reducedMotion, (value) => { if (value) loop?.kill() })
onScopeDispose(() => context?.revert())
</script>

<template>
  <div class="draw-wheel" role="img" :aria-label="$t('activity.drawStyle.wheel')">
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
        <circle cx="150" cy="150" r="26" fill="var(--canvas)" stroke="var(--accent)" stroke-width="2" />
        <circle cx="150" cy="150" r="6" fill="var(--accent)" />
      </g>
      <path d="M 139 4 L 161 4 L 150 32 Z" fill="var(--accent)" />
    </svg>
  </div>
</template>

<style scoped>
.draw-wheel { display: grid; place-items: center; width: min(100%, 18rem); margin: 0 auto; min-height: 15rem; }
.draw-wheel svg { display: block; width: 100%; max-height: 18rem; }
</style>
