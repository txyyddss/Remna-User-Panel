<script setup lang="ts">
import { computed, nextTick, onMounted, onScopeDispose, shallowRef, useTemplateRef, watch } from 'vue'
import { gsap } from 'gsap'

import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { previewPrizes, selectedPreviewIndex, type DrawPresenterProps } from '../selection'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: []; started: [] }>()
const visible = computed(() => previewPrizes(props.prizes, 0, props.result))
const { reducedMotion } = useMotionPreferences()
const reel = useTemplateRef<globalThis.HTMLElement>('reel')
const pulled = shallowRef(false)
let engaged = false
const rowHeight = 52
const rows = computed(() => Array.from({ length: 5 }, (_, cycle) =>
  visible.value.map((prize) => ({ key: `${cycle}:${prize.id}`, label: prize.name })),
).flat())
let loop: gsap.core.Tween | undefined
let stopping: gsap.core.Tween | undefined
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
  if (!engaged) { engaged = true; emit('started') }
  pulled.value = true
  if (release) globalThis.clearTimeout(release)
  release = globalThis.setTimeout(() => { pulled.value = false }, 430)
  if (props.result) settle()
  else loop?.timeScale(2.4)
}

onMounted(start)
watch(() => props.result, async (value) => {
  if (!value) return
  await nextTick()
  if (engaged) settle()
})
watch(reducedMotion, (value) => { if (value) { loop?.kill(); if (props.result && engaged) settle() } })
onScopeDispose(() => {
  if (release) globalThis.clearTimeout(release)
  loop?.kill()
  stopping?.kill()
})
</script>

<template>
  <div class="draw-slot" role="group" :aria-label="$t('activity.drawStyle.slot')">
    <div class="draw-slot__assembly">
      <UButton class="draw-slot__lever" color="neutral" variant="ghost" :disabled="settled" :aria-label="$t('activity.pullLever')" @click="pull">
        <span class="draw-slot__handle" :class="{ 'draw-slot__handle--pulled': pulled }" aria-hidden="true"><span /></span>
        <span class="draw-slot__pivot" aria-hidden="true" />
      </UButton>
      <div class="draw-slot__machine">
        <div class="draw-slot__window" aria-hidden="true">
          <div ref="reel" class="draw-slot__reel">
            <div v-for="row in rows" :key="row.key" class="draw-slot__row">{{ row.label }}</div>
          </div>
        </div>
        <span class="draw-slot__marker" aria-hidden="true" />
      </div>
    </div>
    <p v-if="!result" class="draw-slot__hint" role="status">{{ $t('activity.slotRollingHint') }}</p>
  </div>
</template>

<style scoped>
.draw-slot { display: grid; justify-items: center; align-content: center; gap: 0.7rem; min-height: 15rem; }
.draw-slot__assembly { position: relative; width: min(100%, 20rem); padding-left: 2.6rem; }
.draw-slot__machine { position: relative; width: 100%; padding: 0.65rem; border: 1px solid var(--line-strong); border-radius: var(--radius-panel); background: var(--surface-raised); }
.draw-slot__window { position: relative; height: 9.75rem; overflow: hidden; border: 1px solid var(--line-strong); border-radius: var(--radius-control); background: var(--canvas); }
.draw-slot__window::before, .draw-slot__window::after { position: absolute; right: 0; left: 0; z-index: 1; height: 1px; background: var(--accent); content: ''; }
.draw-slot__window::before { top: 3.25rem; }
.draw-slot__window::after { bottom: 3.25rem; }
.draw-slot__reel { padding-top: 3.25rem; will-change: transform; }
.draw-slot__row { display: grid; place-items: center; height: 3.25rem; padding: 0 0.6rem; color: var(--text); font-size: 0.88rem; text-align: center; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.draw-slot__marker { position: absolute; top: 50%; right: 0.35rem; width: 0; height: 0; border-block: 0.35rem solid transparent; border-right: 0.4rem solid var(--accent); transform: translateY(-50%); }
.draw-slot__lever { position: absolute; top: 50%; left: 0; z-index: 2; width: 3.1rem; height: 5.5rem; padding: 0; border: 0; background: transparent; color: var(--accent); cursor: pointer; transform: translateY(-50%); }
.draw-slot__lever::after { position: absolute; top: 50%; right: 0; width: 0.85rem; height: 1.65rem; border: 1px solid var(--line-strong); border-right: 0; border-radius: 0.4rem 0 0 0.4rem; background: var(--surface-raised); content: ''; transform: translateY(-50%); }
.draw-slot__lever:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.draw-slot__lever:disabled { cursor: default; }
.draw-slot__handle { position: absolute; bottom: 2.65rem; left: 1.05rem; z-index: 1; width: 0.22rem; height: 2.2rem; border-radius: 0.2rem; background: currentColor; transform: rotate(-25deg); transform-origin: 50% 100%; transition: transform 320ms ease; }
.draw-slot__handle span { position: absolute; top: -0.55rem; left: -0.34rem; width: 0.9rem; height: 0.9rem; border: 1px solid var(--surface-raised); border-radius: 50%; background: currentColor; }
.draw-slot__handle--pulled { transform: rotate(35deg); }
.draw-slot__pivot { position: absolute; bottom: 2.1rem; left: 0.82rem; z-index: 3; width: 0.7rem; height: 0.7rem; border: 1px solid var(--line-strong); border-radius: 50%; background: var(--canvas); }
.draw-slot__hint { margin: 0; color: var(--text-muted); font-size: 0.75rem; text-align: center; }
@media (prefers-reduced-motion: reduce) { .draw-slot__reel { will-change: auto; } .draw-slot__handle { transition: none; } }
</style>
