<script setup lang="ts">
import { computed } from 'vue'

import type { OperationReceipt } from '@/api/types'
import { t } from '@/i18n'
import InlineNotice from './InlineNotice.vue'

const props = withDefaults(defineProps<{
  receipt?: OperationReceipt | null
  error?: string | null
  checking?: boolean
  message?: string | null
}>(), { receipt: null, error: null, checking: false, message: null })
const emit = defineEmits<{ refresh: [] }>()

const statusLabel = computed(() => props.receipt ? t(`operations.status.${props.receipt.status}`) : '')
const tone = computed(() => {
  if (props.receipt?.status === 'succeeded') return 'success'
  if (props.receipt?.status === 'queued' || props.receipt?.status === 'processing') return 'info'
  return 'warning'
})
</script>

<template>
  <div v-if="receipt || error" class="operation-status">
    <InlineNotice v-if="receipt" :tone="tone" :title="statusLabel">
      {{ message ?? statusLabel }}
    </InlineNotice>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <UButton
      v-if="receipt && error"
      color="neutral"
      variant="outline"
      icon="i-ph-arrow-clockwise"
      :loading="checking"
      :label="$t('operations.checkStatus')"
      @click="emit('refresh')"
    />
  </div>
</template>

<style scoped>
.operation-status { display: grid; gap: 0.65rem; }
</style>
