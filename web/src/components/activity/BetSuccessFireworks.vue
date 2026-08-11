<script setup lang="ts">
const bursts = [
  { id: 'center', x: '50%', y: '38%', delay: '0ms' },
  { id: 'left', x: '30%', y: '55%', delay: '120ms' },
  { id: 'right', x: '70%', y: '48%', delay: '220ms' },
] as const

const sparks = Array.from({ length: 12 }, (_, index) => ({
  id: index,
  angle: `${index * 30}deg`,
  distance: `${2.4 + (index % 4) * 0.55}rem`,
  delay: `${(index % 4) * 45}ms`,
}))
</script>

<template>
  <div class="bet-fireworks" aria-hidden="true">
    <span
      v-for="burst in bursts"
      :key="burst.id"
      class="bet-fireworks__burst"
      :style="{ '--x': burst.x, '--y': burst.y, '--burst-delay': burst.delay }"
    >
      <span
        v-for="spark in sparks"
        :key="spark.id"
        class="bet-fireworks__spark"
        :style="{ '--angle': spark.angle, '--distance': spark.distance, '--spark-delay': spark.delay }"
      />
    </span>
  </div>
</template>

<style scoped>
.bet-fireworks { position: absolute; inset: 0; z-index: 2; overflow: hidden; pointer-events: none; }
.bet-fireworks__burst { position: absolute; left: var(--x); top: var(--y); width: 0.35rem; height: 0.35rem; animation: bet-fireworks-bloom 760ms cubic-bezier(0.16, 1, 0.3, 1) both; animation-delay: var(--burst-delay); }
.bet-fireworks__spark { position: absolute; left: 50%; bottom: 50%; width: 2px; height: 0.9rem; border-radius: 1px; background: var(--accent); transform-origin: 50% 100%; animation: bet-fireworks-spark 820ms cubic-bezier(0.16, 1, 0.3, 1) both; animation-delay: var(--spark-delay); }
.bet-fireworks__spark:nth-child(3n) { background: var(--success); }
.bet-fireworks__spark:nth-child(3n + 1) { background: var(--accent-strong); }
.bet-fireworks__spark:nth-child(3n + 2) { background: var(--warning); }

@keyframes bet-fireworks-bloom {
  from { opacity: 0; transform: scale(0.35); }
  35% { opacity: 1; transform: scale(1); }
  to { opacity: 0; transform: scale(0.85); }
}

@keyframes bet-fireworks-spark {
  from { opacity: 1; transform: rotate(var(--angle)) translateY(0) scaleY(0.55); }
  to { opacity: 0; transform: rotate(var(--angle)) translateY(var(--distance)) scaleY(1); }
}

@media (prefers-reduced-motion: reduce) {
  .bet-fireworks { display: none; }
}
</style>
