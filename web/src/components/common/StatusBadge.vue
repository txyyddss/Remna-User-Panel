<script setup lang="ts">
import { computed } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import { useMotionPreferences } from '@/composables/useMotionPreferences'

const props = withDefaults(defineProps<{
  tone?: 'neutral' | 'success' | 'warning' | 'danger'
  label: string
}>(), { tone: 'neutral' })

const color = computed(() => ({
  neutral: 'neutral',
  success: 'success',
  warning: 'warning',
  danger: 'error',
} as const)[props.tone])
const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <AnimatePresence mode="wait" :initial="false">
    <motion.span key="status" class="status-badge-motion" :initial="{ opacity: 0, scale: reducedMotion ? 1 : 0.94 }" :animate="{ opacity: 1, scale: 1 }" :exit="{ opacity: 0 }" :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }">
      <UBadge class="status-badge" :class="`status-badge--${tone}`" :color="color" variant="soft" :label="label" />
    </motion.span>
  </AnimatePresence>
</template>

<style scoped>
.status-badge-motion { display: inline-flex; }
</style>
