<script setup lang="ts">
import { watch } from 'vue'
import { motion } from 'motion-v'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import type { DrawPresenterProps } from '../selection'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const { reducedMotion } = useMotionPreferences()
watch(() => props.result, (value) => { if (value) emit('finished') }, { immediate: true })
</script>

<template>
  <div class="draw-simple" role="status">
    <motion.div class="draw-simple__icon" :animate="{ scale: reducedMotion ? 1 : result ? [0.9, 1.08, 1] : [0.98, 1.02, 0.98] }" :transition="{ duration: result ? 0.35 : 1.3, repeat: reducedMotion || result ? 0 : Infinity }">
      <UIcon name="i-ph-gift" aria-hidden="true" />
    </motion.div>
    <p>{{ $t('activity.drawing') }}</p>
  </div>
</template>

<style scoped>
.draw-simple { display: grid; place-items: center; gap: 0.7rem; min-height: 15rem; color: var(--text-muted); }
.draw-simple__icon { display: grid; place-items: center; width: 5rem; height: 5rem; border: 1px solid var(--line-strong); border-radius: var(--radius-panel); color: var(--accent); font-size: 2.5rem; }
.draw-simple p { margin: 0; font-size: 0.82rem; }
</style>
