<script setup lang="ts">
import { useI18n } from '@/i18n'

const { t } = useI18n()
</script>

<template>
  <main class="auth-screen browser-capability-gate" role="alert">
    <div class="auth-screen__copy">
      <p class="eyebrow">{{ t('auth.telegramAccess') }}</p>
      <h1>{{ t('auth.browserUnsupportedTitle') }}</h1>
      <p>{{ t('auth.browserUnsupportedDescription') }}</p>
    </div>
  </main>
</template>
