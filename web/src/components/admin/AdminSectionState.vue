<script setup lang="ts">
import { AnimatePresence, motion } from 'motion-v'

import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { useI18n } from '@/i18n'

defineProps<{
  loading: boolean
  error?: string | null
}>()

defineEmits<{ retry: [] }>()
const { t } = useI18n()
const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <AnimatePresence mode="wait" :initial="false">
    <motion.div
      v-if="loading"
      key="loading"
      class="admin-loading"
      :initial="{ opacity: 0 }"
      :animate="{ opacity: 1 }"
      :exit="{ opacity: 0 }"
      :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }"
    >
      <SkeletonBlock height="5rem" />
      <SkeletonBlock height="5rem" />
      <SkeletonBlock height="5rem" />
    </motion.div>
    <motion.div
      v-else-if="error"
      :key="`error:${error}`"
      class="error-state error-state--compact"
      :initial="{ opacity: 0 }"
      :animate="{ opacity: 1 }"
      :exit="{ opacity: 0 }"
      :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }"
    >
      <h2>{{ t('adminSection.unavailable') }}</h2>
      <p>{{ error }}</p>
      <UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="t('adminSection.retry')" @click="$emit('retry')" />
    </motion.div>
    <motion.div
      v-else
      key="content"
      :initial="{ opacity: 0 }"
      :animate="{ opacity: 1 }"
      :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }"
    >
      <slot />
    </motion.div>
  </AnimatePresence>
</template>
