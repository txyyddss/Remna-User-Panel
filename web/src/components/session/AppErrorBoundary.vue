<script setup lang="ts">
import { onErrorCaptured, shallowRef } from 'vue'

import LanguageControl from '@/components/layout/LanguageControl.vue'
import { useI18n } from '@/i18n'

const failed = shallowRef(false)
const { t } = useI18n()

onErrorCaptured(() => {
  failed.value = true
  return false
})

function reload(): void {
  globalThis.location.reload()
}
</script>

<template>
  <main v-if="failed" class="auth-screen" role="alert">
    <div class="auth-screen__copy">
      <p class="eyebrow">{{ t('app.name') }}</p>
      <h1>{{ t('recovery.title') }}</h1>
      <p>{{ t('recovery.description') }}</p>
    </div>
    <UButton icon="i-ph-arrow-clockwise" :label="t('recovery.reload')" data-haptic="retry" @click="reload" />
    <footer class="auth-screen__locale"><LanguageControl /></footer>
  </main>
  <slot v-else />
</template>
