<script setup lang="ts">
withDefaults(defineProps<{ height?: string }>(), { height: '5rem' })
</script>

<template>
  <div class="skeleton-block" :style="{ height }" aria-hidden="true" />
</template>
