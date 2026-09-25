<script setup lang="ts">
import { computed, nextTick, onMounted, onScopeDispose, shallowRef, useTemplateRef, watch } from 'vue'
import { gsap } from 'gsap'

import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { previewPrizes, selectedPreviewIndex, type DrawPresenterProps } from '../selection'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const visible = computed(() => previewPrizes(props.prizes, 0, props.result))
const { reducedMotion } = useMotionPreferences()
const reel = useTemplateRef<globalThis.HTMLElement>('reel')
const pulled = shallowRef(false)
const rowHeight = 52
const rows = computed(() => Array.from({ length: 5 }, (_, cycle) =>
  visible.value.map((prize) => ({ key: `${cycle}:${prize.id}`, label: prize.name })),
).flat())
let loop: gsap.core.Tween | undefined
let stopping: gsap.core.Tween | undefined
let autoStop: ReturnType<typeof globalThis.setTimeout> | undefined
let release: ReturnType<typeof globalThis.setTimeout> | undefined
let settled = false

function start(): void {
  if (!reel.value || !visible.value.length || reducedMotion.value) return
  loop = gsap.to(reel.value, {
    y: -visible.value.length * rowHeight,
    duration: Math.max(0.9, visible.value.length * 0.15),
    repeat: -1, ease: 'none',
  })
}

function settle(): void {
  if (settled || !props.result) return
  settled = true
  if (autoStop) globalThis.clearTimeout(autoStop)
  loop?.kill()
  const target = selectedPreviewIndex(visible.value, props.result)
  if (!reel.value || target < 0 || reducedMotion.value) { emit('finished'); return }
  const count = visible.value.length
  const current = -Number(gsap.getProperty(reel.value, 'y')) / rowHeight
  const end = (Math.ceil(current / count) + 2) * count + target
  stopping = gsap.to(reel.value, {
    y: -end * rowHeight, duration: 2.1, ease: 'power3.out',
    onComplete: () => emit('finished'),
  })
}

function pull(): void {
  if (settled) return
  pulled.value = true
  if (release) globalThis.clearTimeout(release)
  release = globalThis.setTimeout(() => { pulled.value = false; loop?.timeScale(1) }, 430)
  if (props.result) settle()
  else loop?.timeScale(2.4)
}

onMounted(() => {
  if (props.result) settle()
  else start()
})
watch(() => props.result, async (value) => {
  if (!value) return
  await nextTick()
  if (!settled) autoStop = globalThis.setTimeout(settle, 1500)
})
watch(reducedMotion, (value) => { if (value) { loop?.kill(); if (props.result) settle() } })
onScopeDispose(() => {
  if (autoStop) globalThis.clearTimeout(autoStop)
  if (release) globalThis.clearTimeout(release)
  loop?.kill()
  stopping?.kill()
})
</script>

<template>
  <div class="draw-slot" role="group" :aria-label="$t('activity.drawStyle.slot')">
    <div class="draw-slot__machine">
      <div class="draw-slot__window" aria-hidden="true">
        <div ref="reel" class="draw-slot__reel">
          <div v-for="row in rows" :key="row.key" class="draw-slot__row">{{ row.label }}</div>
        </div>
      </div>
      <span class="draw-slot__marker" aria-hidden="true" />
    </div>
    <UButton class="draw-slot__lever" color="neutral" variant="outline" :aria-label="$t('activity.pullLever')" @click="pull">
      <span class="draw-slot__handle" :class="{ 'draw-slot__handle--pulled': pulled }" aria-hidden="true"><span /></span>
      <span>{{ $t('activity.pullLever') }}</span>
    </UButton>
    <p class="draw-slot__hint" role="status">{{ $t(result ? 'activity.slotReadyHint' : 'activity.slotRollingHint') }}</p>
  </div>
</template>

<style scoped>
.draw-slot { display: grid; justify-items: center; align-content: center; gap: 0.7rem; min-height: 15rem; }
.draw-slot__machine { position: relative; width: min(100%, 17.5rem); padding: 0.65rem; border: 1px solid var(--line-strong); border-radius: var(--radius-panel); background: var(--surface-raised); }
.draw-slot__window { position: relative; height: 9.75rem; overflow: hidden; border: 1px solid var(--line-strong); border-radius: var(--radius-control); background: var(--canvas); }
.draw-slot__window::before, .draw-slot__window::after { position: absolute; right: 0; left: 0; z-index: 1; height: 1px; background: var(--accent); content: ''; }
.draw-slot__window::before { top: 3.25rem; }
.draw-slot__window::after { bottom: 3.25rem; }
.draw-slot__reel { padding-top: 3.25rem; will-change: transform; }
.draw-slot__row { display: grid; place-items: center; height: 3.25rem; padding: 0 0.6rem; color: var(--text); font-size: 0.88rem; text-align: center; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.draw-slot__marker { position: absolute; top: 50%; right: 0.35rem; width: 0; height: 0; border-block: 0.35rem solid transparent; border-right: 0.4rem solid var(--accent); transform: translateY(-50%); }
.draw-slot__lever { min-height: 3rem; gap: 0.65rem; }
.draw-slot__handle { display: grid; align-items: start; width: 1rem; height: 1.55rem; border-left: 2px solid currentColor; transform: rotate(-22deg); transform-origin: 50% 90%; transition: transform 180ms ease; }
.draw-slot__handle span { width: 0.75rem; height: 0.75rem; margin-left: -0.42rem; border-radius: 50%; background: currentColor; }
.draw-slot__handle--pulled { transform: rotate(28deg); }
.draw-slot__hint { margin: 0; color: var(--text-muted); font-size: 0.75rem; text-align: center; }
@media (prefers-reduced-motion: reduce) { .draw-slot__reel { will-change: auto; } .draw-slot__handle { transition: none; } }
</style>
