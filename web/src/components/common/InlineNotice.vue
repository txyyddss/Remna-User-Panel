<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  tone?: 'info' | 'warning' | 'success'
  title?: string
}>(), {
  tone: 'info',
  title: undefined,
})

const presentation = computed(() => ({
  info: { icon: 'i-ph-info-fill', color: 'info' },
  warning: { icon: 'i-ph-warning-circle-fill', color: 'warning' },
  success: { icon: 'i-ph-check-circle-fill', color: 'success' },
} as const)[props.tone])
</script>

<template>
  <UAlert
    class="notice"
    :class="`notice--${tone}`"
    :icon="presentation.icon"
    :color="presentation.color"
    variant="soft"
    :title="title"
    role="status"
  >
    <template #description><slot /></template>
  </UAlert>
</template>
