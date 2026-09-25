<script setup lang="ts">
import { computed } from 'vue'

import type { ActivityResult } from '@/api/features'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { hasPositiveDrawReward } from './selection'

const props = defineProps<{ result: ActivityResult }>()
const { reducedMotion } = useMotionPreferences()
const visible = computed(() => !reducedMotion.value && hasPositiveDrawReward(props.result))
const sparks = Array.from({ length: 24 }, (_, index) => ({
  id: index,
  angle: `${index * 15}deg`,
  distance: `${5.5 + index % 4}rem`,
  delay: `${index % 4 * 30}ms`,
}))
</script>

<template>
  <div v-if="visible" class="draw-celebration" aria-hidden="true">
    <span
      v-for="spark in sparks"
      :key="spark.id"
      class="draw-celebration__spark"
      :style="{ '--angle': spark.angle, '--distance': spark.distance, '--delay': spark.delay }"
    />
  </div>
</template>

<style scoped>
.draw-celebration { position: absolute; inset: 0; z-index: 2; overflow: hidden; pointer-events: none; }
.draw-celebration__spark { position: absolute; top: 24%; left: 50%; width: 0.22rem; height: 0.95rem; border-radius: 0.08rem; background: var(--warning); transform-origin: 50% 0; animation: draw-spark 1.25s cubic-bezier(0.16, 1, 0.3, 1) var(--delay) both; }
.draw-celebration__spark:nth-child(3n) { background: var(--text-muted); }
.draw-celebration__spark:nth-child(3n + 1) { background: var(--text); }
@keyframes draw-spark {
  0% { opacity: 0; transform: rotate(var(--angle)) translateY(0) scale(0.2); }
  18% { opacity: 1; }
  100% { opacity: 0; transform: rotate(var(--angle)) translateY(var(--distance)) scale(0.8); }
}
@media (prefers-reduced-motion: reduce) { .draw-celebration { display: none; } }
</style>
