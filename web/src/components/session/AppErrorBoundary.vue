<script setup lang="ts">
import { onErrorCaptured, shallowRef } from 'vue'
import { motion } from 'motion-v'

import LanguageControl from '@/components/layout/LanguageControl.vue'
import { useI18n } from '@/i18n'
import { useMotionPreferences } from '@/composables/useMotionPreferences'

const failed = shallowRef(false)
const { t } = useI18n()
const { reducedMotion } = useMotionPreferences()

onErrorCaptured(() => {
  failed.value = true
  return false
})

function reload(): void {
  globalThis.location.reload()
}
</script>

<template>
  <motion.main
    v-if="failed"
    class="auth-screen"
    role="alert"
    :initial="{ opacity: 0 }"
    :animate="{ opacity: 1 }"
    :transition="{ duration: reducedMotion ? 0.08 : 0.14, ease: 'easeOut' }"
  >
    <div class="auth-screen__copy">
      <p class="eyebrow">{{ t('app.name') }}</p>
      <h1>{{ t('recovery.title') }}</h1>
      <p>{{ t('recovery.description') }}</p>
    </div>
    <UButton icon="i-ph-arrow-clockwise" :label="t('recovery.reload')" data-haptic="retry" @click="reload" />
    <footer class="auth-screen__locale"><LanguageControl /></footer>
  </motion.main>
  <slot v-else />
</template>
