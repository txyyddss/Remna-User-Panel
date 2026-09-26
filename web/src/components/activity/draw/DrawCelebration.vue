<script setup lang="ts">
import { computed } from 'vue'

import type { ActivityResult } from '@/api/features'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { hasPositiveDrawReward } from './selection'

const props = defineProps<{ result: ActivityResult }>()
const { reducedMotion } = useMotionPreferences()
const visible = computed(() => !reducedMotion.value && hasPositiveDrawReward(props.result))
const bursts = [
  { x: '13%', y: '24%', delay: 0 },
  { x: '35%', y: '58%', delay: 130 },
  { x: '56%', y: '18%', delay: 260 },
  { x: '83%', y: '36%', delay: 100 },
  { x: '72%', y: '73%', delay: 380 },
  { x: '20%', y: '78%', delay: 310 },
]
const sparks = Array.from({ length: 14 }, (_, index) => ({
  id: index,
  angle: `${index * 360 / 14}deg`,
  distance: `${4.3 + index % 3 * 1.1}rem`,
}))
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="draw-celebration" aria-hidden="true">
      <span v-for="(burst, index) in bursts" :key="index" class="draw-celebration__burst" :style="{ '--x': burst.x, '--y': burst.y }">
        <span
          v-for="spark in sparks"
          :key="spark.id"
          class="draw-celebration__spark"
          :style="{ '--angle': spark.angle, '--distance': spark.distance, '--delay': `${burst.delay + spark.id * 18}ms` }"
        />
      </span>
    </div>
  </Teleport>
</template>

<style scoped>
.draw-celebration { position: fixed; inset: 0; z-index: 10000; overflow: hidden; pointer-events: none; }
.draw-celebration__burst { position: absolute; top: var(--y); left: var(--x); }
.draw-celebration__spark { position: absolute; top: 0; left: 0; width: 0.2rem; height: 0.85rem; border-radius: 0.08rem; background: var(--warning); transform-origin: 50% 0; animation: draw-spark 1.4s cubic-bezier(0.16, 1, 0.3, 1) var(--delay) both; }
.draw-celebration__spark:nth-child(3n) { background: var(--text-muted); }
.draw-celebration__spark:nth-child(3n + 1) { background: var(--text); }
@keyframes draw-spark {
  0% { opacity: 0; transform: rotate(var(--angle)) translateY(0) scale(0.25); }
  14% { opacity: 1; }
  100% { opacity: 0; transform: rotate(var(--angle)) translateY(var(--distance)) scale(0.65); }
}
@media (prefers-reduced-motion: reduce) { .draw-celebration { display: none; } }
</style>
