<script setup lang="ts">
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useI18n } from '@/i18n'

defineProps<{
  loading: boolean
  error?: string | null
}>()

defineEmits<{ retry: [] }>()
const { t } = useI18n()
</script>

<template>
  <div v-if="loading" class="admin-loading">
    <SkeletonBlock height="5rem" />
    <SkeletonBlock height="5rem" />
    <SkeletonBlock height="5rem" />
  </div>
  <div v-else-if="error" class="error-state error-state--compact">
    <h2>{{ t('adminSection.unavailable') }}</h2>
    <p>{{ error }}</p>
    <UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="t('adminSection.retry')" @click="$emit('retry')" />
  </div>
  <slot v-else />
</template>
