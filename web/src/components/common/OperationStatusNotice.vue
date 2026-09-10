<script setup lang="ts">
import { computed } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { OperationReceipt } from '@/api/types'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
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
const receiptKey = computed(() => props.receipt
  ? `${props.receipt.status}:${props.message ?? statusLabel.value}`
  : '')
const errorKey = computed(() => props.error ?? '')
const { reducedMotion, offset } = useMotionPreferences()
</script>

<template>
  <motion.div v-if="receipt || error" layout class="operation-status">
    <AnimatePresence mode="wait" :initial="false">
      <motion.div
        v-if="receipt"
        :key="`receipt:${receiptKey}`"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: offset(-5) }"
        :animate="{ opacity: 1, y: 0 }"
        :exit="{ opacity: 0, y: reducedMotion ? 0 : -4 }"
        :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }"
      >
        <InlineNotice :tone="tone" :title="statusLabel">
          {{ message ?? statusLabel }}
        </InlineNotice>
      </motion.div>
      <motion.div
        v-if="error"
        :key="`error:${errorKey}`"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: offset(-5) }"
        :animate="{ opacity: 1, y: 0 }"
        :exit="{ opacity: 0, y: reducedMotion ? 0 : -4 }"
        :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }"
      >
        <InlineNotice tone="warning">{{ error }}</InlineNotice>
      </motion.div>
      <motion.div
        v-if="receipt && error"
        key="refresh"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: offset(4) }"
        :animate="{ opacity: 1, y: 0 }"
        :exit="{ opacity: 0 }"
        :transition="{ duration: reducedMotion ? 0.08 : 0.14, ease: 'easeOut' }"
      >
        <UButton
          color="neutral"
          variant="outline"
          icon="i-ph-arrow-clockwise"
          :loading="checking"
          :label="$t('operations.checkStatus')"
          data-haptic="retry"
          @click="emit('refresh')"
        />
      </motion.div>
    </AnimatePresence>
  </motion.div>
</template>

<style scoped>
.operation-status { display: grid; gap: 0.65rem; }
</style>
