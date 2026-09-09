<script setup lang="ts">
import { motion } from 'motion-v'

import { useMotionPreferences } from '@/composables/useMotionPreferences'

defineProps<{
  showBack: boolean
  nextDisabled: boolean
  loading: boolean
  nextLabel: string
}>()

defineEmits<{ back: []; next: [] }>()
const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <motion.footer layout class="catalog-flow-controls">
    <motion.div v-if="showBack" layout :initial="reducedMotion ? false : { opacity: 0, x: -6 }" :animate="{ opacity: 1, x: 0 }" :exit="{ opacity: 0, x: reducedMotion ? 0 : -6 }" :transition="{ duration: reducedMotion ? 0.08 : 0.15, ease: 'easeOut' }">
      <UButton color="neutral" variant="ghost" leading-icon="i-ph-arrow-left" :label="$t('catalog.back')" data-haptic="navigate" @click="$emit('back')" />
    </motion.div>
    <UButton class="catalog-flow-controls__next" :disabled="nextDisabled" :loading="loading" trailing-icon="i-ph-arrow-right" :label="nextLabel" data-haptic="navigate" @click="$emit('next')" />
  </motion.footer>
</template>

<style scoped>
.catalog-flow-controls { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 0.55rem; margin-top: 1rem; padding: 0.65rem 0 0; border-top: 1px solid var(--line); }
.catalog-flow-controls__next { width: 100%; }
</style>
