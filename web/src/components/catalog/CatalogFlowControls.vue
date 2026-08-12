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
    <UButton v-if="showBack" color="neutral" variant="ghost" leading-icon="i-ph-arrow-left" :label="$t('catalog.back')" data-haptic @click="$emit('back')" />
    <UButton class="catalog-flow-controls__next" :disabled="nextDisabled" :loading="loading" trailing-icon="i-ph-arrow-right" :label="nextLabel" data-haptic @click="$emit('next')" />
  </footer>
</template>

<style scoped>
.catalog-flow-controls { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 0.55rem; margin-top: 1rem; padding: 0.65rem 0 0; border-top: 1px solid var(--line); }
.catalog-flow-controls__next { width: 100%; }
</style>
