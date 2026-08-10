<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  tone?: 'neutral' | 'success' | 'warning' | 'danger'
  label: string
}>(), { tone: 'neutral' })

const color = computed(() => ({
  neutral: 'neutral',
  success: 'success',
  warning: 'warning',
  danger: 'error',
} as const)[props.tone])
</script>

<template>
  <UBadge class="status-badge" :class="`status-badge--${tone}`" :color="color" variant="soft" :label="label" />
</template>
