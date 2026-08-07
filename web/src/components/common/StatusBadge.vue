<script setup lang="ts">
withDefaults(defineProps<{
  tone?: 'neutral' | 'success' | 'warning' | 'danger'
  label: string
}>(), { tone: 'neutral' })
</script>

<template>
  <span class="status-badge" :class="`status-badge--${tone}`">{{ label }}</span>
</template>
