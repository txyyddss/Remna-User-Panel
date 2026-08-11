<script setup lang="ts">
defineProps<{
  showBack: boolean
  nextDisabled: boolean
  loading: boolean
  nextLabel: string
}>()

defineEmits<{ back: []; next: [] }>()
</script>

<template>
  <footer class="catalog-flow-controls">
    <UButton v-if="showBack" color="neutral" variant="ghost" leading-icon="i-ph-arrow-left" :label="$t('catalog.back')" @click="$emit('back')" />
    <UButton class="catalog-flow-controls__next" :disabled="nextDisabled" :loading="loading" trailing-icon="i-ph-arrow-right" :label="nextLabel" @click="$emit('next')" />
  </footer>
</template>

<style scoped>
.catalog-flow-controls { position: sticky; bottom: calc(74px + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom))); z-index: 15; display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 0.55rem; margin-top: 1rem; padding: 0.65rem; border: 1px solid var(--line-strong); border-radius: var(--radius-panel); background: rgb(18 21 18 / 0.94); backdrop-filter: blur(18px); }
.catalog-flow-controls__next { width: 100%; }
</style>
