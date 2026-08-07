<script setup lang="ts">
import { PhCirclesFour } from '@phosphor-icons/vue'
</script>

<template>
  <main class="auth-screen" aria-live="polite" aria-busy="true">
    <div class="brand-mark brand-mark--large" aria-hidden="true">
      <PhCirclesFour :size="28" weight="fill" />
    </div>
    <p class="auth-screen__label">Securing your ride</p>
    <div class="auth-screen__skeleton" aria-hidden="true">
      <span />
      <span />
      <span />
    </div>
  </main>
</template>
