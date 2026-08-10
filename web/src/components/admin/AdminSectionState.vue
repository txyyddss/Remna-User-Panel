<script setup lang="ts">
import { PhArrowClockwise } from '@phosphor-icons/vue'

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
    <button class="button button--secondary" type="button" @click="$emit('retry')">
      <PhArrowClockwise :size="18" /> {{ t('adminSection.retry') }}
    </button>
  </div>
  <slot v-else />
</template>
