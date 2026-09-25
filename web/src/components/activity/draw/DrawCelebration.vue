<script setup lang="ts">
import { onUnmounted, watch } from 'vue'
import { Vue3Lottie } from 'vue3-lottie'

import type { ActivityResult } from '@/api/features'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import artwork from './assets/celebration.json'
import { hasPositiveDrawReward } from './selection'

const props = defineProps<{ result: ActivityResult }>()
const { reducedMotion } = useMotionPreferences()
let active = true
let stopConfetti: (() => void) | undefined

watch(() => props.result.id, async () => {
  if (reducedMotion.value || !hasPositiveDrawReward(props.result)) return
  try {
    const { default: confetti } = await import('canvas-confetti')
    stopConfetti = confetti.reset
    if (active) await confetti({
      particleCount: 34, spread: 52, origin: { y: 0.42 },
      colors: ['#a6d9bb', '#87d6a2', '#d8c18f'],
      disableForReducedMotion: true,
    })
  } catch {
    // Visual effects must never hide a recorded result.
  }
}, { immediate: true })
onUnmounted(() => {
  active = false
  stopConfetti?.()
})
</script>

<template>
  <Vue3Lottie
    v-if="!reducedMotion && hasPositiveDrawReward(result)"
    class="draw-celebration"
    :animation-data="artwork"
    :loop="1"
    :height="78"
    :width="78"
    aria-hidden="true"
  />
</template>

<style scoped>
.draw-celebration { position: absolute; top: -0.4rem; right: 0; pointer-events: none; }
</style>
