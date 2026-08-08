<script setup lang="ts">
import { PhArrowClockwise, PhTelegramLogo } from '@phosphor-icons/vue'
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'

defineProps<{ message: string }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <main class="auth-screen">
    <LanguageSwitcher />
    <div class="brand-mark brand-mark--large" aria-hidden="true">
      <PhTelegramLogo :size="28" weight="fill" />
    </div>
    <div class="auth-screen__copy">
      <p class="eyebrow">{{ $t('auth.telegramAccess') }}</p>
      <h1>{{ $t('auth.openInTelegram') }}</h1>
      <p>{{ message }}</p>
    </div>
    <button class="button button--primary" type="button" @click="$emit('retry')">
      <PhArrowClockwise :size="19" />
      {{ $t('auth.tryAgain') }}
    </button>
  </main>
</template>
