<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, shallowRef, useTemplateRef, watch } from 'vue'

import type { Purchase } from '@/api/types'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import RideSummaryFace from './RideSummaryFace.vue'
import RolloverDetailFace from './RolloverDetailFace.vue'
import { useRolloverDetail } from '@/composables/useRolloverDetail'

const props = defineProps<{
  active: Purchase
  squadNames?: readonly string[]
}>()

const flipped = shallowRef(false)
const cardHeight = shallowRef<string | null>(null)
const { detail, loading, error, load, retry, reset } = useRolloverDetail()
const summaryFace = useTemplateRef<globalThis.HTMLElement>('summaryFace')
const rolloverFace = useTemplateRef<globalThis.HTMLElement>('rolloverFace')
const flipInnerStyle = computed(() => cardHeight.value ? { height: cardHeight.value } : undefined)
const ownsTelegramBackButton = computed(() => flipped.value)
let changingFace = false
let resizeObserver: globalThis.ResizeObserver | undefined
let animationFrame: number | undefined
let frameTimer: ReturnType<typeof globalThis.setTimeout> | undefined
let resolveFrame: (() => void) | undefined
let unmounted = false

function activeFace(): globalThis.HTMLElement | null {
  return flipped.value ? rolloverFace.value : summaryFace.value
}

function syncHeight(): void {
  if (unmounted) return
  const height = activeFace()?.offsetHeight ?? 0
  if (height > 0) cardHeight.value = `${height}px`
}

function nextFrame(): Promise<void> {
  return new Promise((resolve) => {
    const finish = (): void => {
      animationFrame = undefined
      frameTimer = undefined
      resolveFrame = undefined
      resolve()
    }
    resolveFrame = finish
    if (typeof globalThis.requestAnimationFrame === 'function') {
      animationFrame = globalThis.requestAnimationFrame(finish)
      return
    }
    frameTimer = globalThis.setTimeout(finish, 0)
  })
}

async function changeFace(next: boolean): Promise<void> {
  if (changingFace || flipped.value === next) return
  changingFace = true
  try {
    syncHeight()
    await nextFrame()
    if (unmounted) return
    flipped.value = next
    await nextTick()
    syncHeight()
  } finally {
    changingFace = false
  }
}

async function showDetail(): Promise<void> {
  if (flipped.value || changingFace) return
  void load(props.active.id)
  await changeFace(true)
}

async function showSummary(): Promise<void> {
  if (!flipped.value || changingFace) return
  await changeFace(false)
}

useTelegramBackButton(ownsTelegramBackButton, showSummary)

watch(() => props.active.id, () => {
  flipped.value = false
  cardHeight.value = null
  reset()
})

watch([detail, loading, error], () => {
  if (flipped.value) void nextTick().then(syncHeight)
})

onMounted(() => {
  syncHeight()
  if (typeof globalThis.ResizeObserver !== 'function') return
  resizeObserver = new globalThis.ResizeObserver(syncHeight)
  if (summaryFace.value) resizeObserver.observe(summaryFace.value)
  if (rolloverFace.value) resizeObserver.observe(rolloverFace.value)
})

onUnmounted(() => {
  unmounted = true
  resizeObserver?.disconnect()
  if (animationFrame !== undefined && typeof globalThis.cancelAnimationFrame === 'function') globalThis.cancelAnimationFrame(animationFrame)
  if (frameTimer !== undefined) globalThis.clearTimeout(frameTimer)
  resolveFrame?.()
})
</script>

<template>
  <div class="home-ride__summary home-ride__flip-card" :class="{ 'is-flipped': flipped }">
    <div class="home-ride__flip-inner" :style="flipInnerStyle">
      <div ref="summaryFace" class="home-ride__flip-face home-ride__flip-face--front">
        <UButton
          id="your-ride-summary"
          type="button"
          class="home-ride__summary-trigger"
          block
          color="neutral"
          variant="ghost"
          aria-controls="your-ride-rollover"
          :aria-expanded="flipped"
          :aria-hidden="flipped"
          :inert="flipped"
          data-haptic="open"
          @click="showDetail"
        >
          <RideSummaryFace :active="active" :squad-names="squadNames" />
        </UButton>
      </div>
      <div id="your-ride-rollover" ref="rolloverFace" class="home-ride__flip-face home-ride__flip-face--back" :aria-hidden="!flipped" :inert="!flipped">
        <RolloverDetailFace :detail="detail" :loading="loading" :error="error" @back="showSummary" @retry="retry" />
      </div>
    </div>
  </div>
</template>
