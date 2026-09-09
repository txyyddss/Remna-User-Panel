<script setup lang="ts">
import { onMounted } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { OnboardingWelcomeMessage } from '@/api/features'
import LanguageControl from '@/components/layout/LanguageControl.vue'
import { useIntroSequence } from '@/composables/useIntroSequence'
import { motionDurations } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'

const emit = defineEmits<{ complete: [] }>()
const props = defineProps<{ messages: readonly OnboardingWelcomeMessage[] }>()
const { index, message, progress, start, skip } = useIntroSequence({
  messages: () => props.messages,
  onComplete: () => emit('complete'),
})
const { reducedMotion, offset } = useMotionPreferences()

onMounted(start)
</script>

<template>
  <section class="intro-sequence">
    <div class="intro-sequence__center">
      <AnimatePresence mode="wait" :initial="false">
        <motion.h1
          :key="index"
          :initial="{ opacity: 0, y: offset(12) }"
          :animate="{ opacity: 1, y: 0 }"
          :exit="{ opacity: 0, y: reducedMotion ? 0 : 10 }"
          :transition="{ duration: reducedMotion ? 0.08 : motionDurations.normal, ease: 'easeOut' }"
        >
          {{ message }}
        </motion.h1>
      </AnimatePresence>
    </div>
    <footer class="intro-sequence__footer">
      <LanguageControl />
      <UButton color="neutral" variant="ghost" trailing-icon="i-ph-arrow-right" :label="$t('common.skip')" data-haptic="navigate" @click="skip" />
    </footer>
    <UProgress class="intro-sequence__progress" :model-value="progress" :max="100" aria-hidden="true" />
  </section>
</template>
