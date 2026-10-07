<script setup lang="ts">
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { useI18n } from '@/i18n'

defineProps<{
  loading: boolean
  error?: string | null
}>()

defineEmits<{ retry: [] }>()
const { t } = useI18n()
const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <Transition name="admin-state" :mode="reducedMotion ? undefined : 'out-in'" :css="!reducedMotion">
    <div
      v-if="loading"
      key="loading"
      class="admin-loading"
    >
      <SkeletonBlock height="5rem" />
      <SkeletonBlock height="5rem" />
      <SkeletonBlock height="5rem" />
    </div>
    <div
      v-else-if="error"
      :key="`error:${error}`"
      class="error-state error-state--compact"
    >
      <h2>{{ t('adminSection.unavailable') }}</h2>
      <p>{{ error }}</p>
      <UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="t('adminSection.retry')" @click="$emit('retry')" />
    </div>
    <div
      v-else
      key="content"
    >
      <slot />
    </div>
  </Transition>
</template>

<style scoped>
.admin-state-enter-active, .admin-state-leave-active { transition: opacity 160ms ease-out; }
.admin-state-enter-from, .admin-state-leave-to { opacity: 0; }
</style>
