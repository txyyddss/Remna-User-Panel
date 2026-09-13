<script setup lang="ts">
import { AnimatePresence, motion } from 'motion-v'

import { useMotionPreferences } from '@/composables/useMotionPreferences'

defineProps<{
  color: string
  label: string
  value: string
}>()

const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <AnimatePresence mode="wait" :initial="false">
    <motion.output
      :key="`${label}:${value}`"
      class="statistics-chart-detail"
      aria-live="polite"
      :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 3 }"
      :animate="{ opacity: 1, y: 0 }"
      :exit="{ opacity: 0 }"
      :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }"
    >
      <span class="statistics-legend__swatch" :style="{ backgroundColor: color }" aria-hidden="true" />
      <span :title="label">{{ label }}</span>
      <strong>{{ value }}</strong>
    </motion.output>
  </AnimatePresence>
</template>
