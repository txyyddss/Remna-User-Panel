<script setup lang="ts">
defineProps<{
  color: string
  label: string
  value: string
}>()
</script>

<template>
  <output class="statistics-chart-detail" aria-live="polite">
    <span class="statistics-legend__swatch" :style="{ backgroundColor: color }" aria-hidden="true" />
    <span :title="label">{{ label }}</span>
    <strong>{{ value }}</strong>
  </output>
</template>
