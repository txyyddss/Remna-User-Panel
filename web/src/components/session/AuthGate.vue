<script setup lang="ts">
import { computed } from 'vue'

import LanguageControl from '@/components/layout/LanguageControl.vue'
import { isTelegramWebAppDetected } from '@/utils/telegram'

const props = defineProps<{ message: string }>()
defineEmits<{ retry: [] }>()
const authRequestFailed = computed(() => isTelegramWebAppDetected() && props.message !== '')
</script>

<template>
  <main class="auth-screen">
    <div class="auth-screen__copy">
      <p class="eyebrow">{{ $t('auth.telegramAccess') }}</p>
      <h1>{{ authRequestFailed ? $t('auth.authenticationFailed') : $t('auth.openInTelegram') }}</h1>
      <p>{{ message }}</p>
    </div>
    <UButton
      :label="$t('auth.tryAgain')"
      icon="i-ph-arrow-clockwise"
      data-haptic="retry"
      @click="$emit('retry')"
    />
    <footer class="auth-screen__locale"><LanguageControl /></footer>
  </main>
</template>
