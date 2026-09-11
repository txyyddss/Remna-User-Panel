<script setup lang="ts">
import { AnimatePresence, motion } from 'motion-v'

import type { SessionStatus } from '@/stores/session'
import { motionDurations } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'

import LoadingScreen from './LoadingScreen.vue'
import { useLoadingSequence } from './useLoadingSequence'

const props = defineProps<{ status: SessionStatus }>()
const { showing } = useLoadingSequence(() => props.status)
const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <div class="session-entrance">
    <AnimatePresence :initial="false" mode="sync">
      <motion.div
        v-if="!showing"
        key="session-page"
        class="session-entrance__page"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0.85, y: 4 }"
        :animate="{ opacity: 1, y: 0 }"
        :exit="{ opacity: 0 }"
        :transition="{ duration: reducedMotion ? 0.08 : motionDurations.normal, ease: 'easeOut' }"
      >
        <slot />
      </motion.div>
      <motion.div
        v-else
        key="session-loading"
        class="session-entrance__cover"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 1 }"
        :animate="{ opacity: 1 }"
        :exit="{ opacity: 0 }"
        :transition="{ duration: reducedMotion ? 0.08 : 0.12, ease: 'easeOut' }"
      >
        <LoadingScreen />
      </motion.div>
    </AnimatePresence>
  </div>
</template>
