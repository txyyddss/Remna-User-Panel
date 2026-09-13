<script setup lang="ts">
import { motion } from 'motion-v'

import { useI18n } from '@/i18n'
import { useMotionPreferences } from '@/composables/useMotionPreferences'

const { t } = useI18n()
const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <motion.main
    class="auth-screen browser-capability-gate"
    role="alert"
    :initial="{ opacity: 0 }"
    :animate="{ opacity: 1 }"
    :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }"
  >
    <div class="auth-screen__copy">
      <p class="eyebrow">{{ t('auth.telegramAccess') }}</p>
      <h1>{{ t('auth.browserUnsupportedTitle') }}</h1>
      <p>{{ t('auth.browserUnsupportedDescription') }}</p>
    </div>
  </motion.main>
</template>
