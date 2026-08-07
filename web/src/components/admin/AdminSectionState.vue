<script setup lang="ts">
import { PhArrowClockwise } from '@phosphor-icons/vue'

import SkeletonBlock from '@/components/common/SkeletonBlock.vue'

defineProps<{
  loading: boolean
  error?: string | null
}>()

defineEmits<{ retry: [] }>()
</script>

<template>
  <div v-if="loading" class="admin-loading">
    <SkeletonBlock height="5rem" />
    <SkeletonBlock height="5rem" />
    <SkeletonBlock height="5rem" />
  </div>
  <div v-else-if="error" class="error-state error-state--compact">
    <h2>This section is unavailable.</h2>
    <p>{{ error }}</p>
    <button class="button button--secondary" type="button" @click="$emit('retry')">
      <PhArrowClockwise :size="18" /> Retry
    </button>
  </div>
  <slot v-else />
</template>
