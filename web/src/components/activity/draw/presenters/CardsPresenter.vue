<script setup lang="ts">
import { onUnmounted, watch } from 'vue'
import { motion } from 'motion-v'
import type { DrawPresenterProps } from '../selection'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
let timer: ReturnType<typeof globalThis.setTimeout> | undefined
watch(() => props.result, (value) => {
  if (timer) globalThis.clearTimeout(timer)
  if (value) timer = globalThis.setTimeout(() => emit('finished'), 850)
}, { immediate: true })
onUnmounted(() => { if (timer) globalThis.clearTimeout(timer) })
</script>

<template>
  <div class="draw-cards" role="status">
    <motion.div v-for="index in 3" :key="index" class="draw-cards__card" :animate="{ rotateY: result && index === 2 ? 180 : 0, y: result && index === 2 ? -8 : 0 }" :transition="{ duration: 0.7, ease: 'easeOut' }">
      <span class="draw-cards__front"><UIcon name="i-ph-sparkle" aria-hidden="true" /></span>
      <strong v-if="result && index === 2" class="draw-cards__back">{{ result.prizeName }}</strong>
    </motion.div>
  </div>
</template>

<style scoped>
.draw-cards { display: flex; justify-content: center; align-items: center; gap: 0.5rem; min-height: 15rem; perspective: 40rem; }
.draw-cards__card { position: relative; display: grid; place-items: center; width: min(27%, 6rem); min-height: 8rem; border: 1px solid var(--line-strong); border-radius: var(--radius-control); background: var(--surface-raised); color: var(--accent); transform-style: preserve-3d; text-align: center; overflow-wrap: anywhere; }
.draw-cards__front, .draw-cards__back { position: absolute; inset: 0; display: grid; place-items: center; padding: 0.45rem; backface-visibility: hidden; }
.draw-cards__front { font-size: 1.8rem; }
.draw-cards__back { color: var(--text); font-size: 0.82rem; transform: rotateY(180deg); }
</style>
