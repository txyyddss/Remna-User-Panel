<script setup lang="ts">
import { PhArrowClockwise, PhTelegramLogo } from '@phosphor-icons/vue'

defineProps<{ message: string }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <main class="auth-screen">
    <div class="brand-mark brand-mark--large" aria-hidden="true">
      <PhTelegramLogo :size="28" weight="fill" />
    </div>
    <div class="auth-screen__copy">
      <p class="eyebrow">Telegram access</p>
      <h1>Open TX Carpool in Telegram.</h1>
      <p>{{ message }}</p>
    </div>
    <button class="button button--primary" type="button" @click="$emit('retry')">
      <PhArrowClockwise :size="19" />
      Try again
    </button>
  </main>
</template>
