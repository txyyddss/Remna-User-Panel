<script setup lang="ts">
import { computed } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { IPBlock, OperationReceipt } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { formatDateTime } from '@/utils/format'

const open = defineModel<boolean>('open', { required: true })
const props = withDefaults(defineProps<{
  block: IPBlock | null
  receipt?: OperationReceipt | null
  busy?: boolean
  checking?: boolean
  error?: string | null
}>(), { receipt: null, busy: false, checking: false, error: null })
const emit = defineEmits<{ confirm: []; refresh: [] }>()
const ownsBack = computed(() => open.value)
const workflowState = computed(() => props.busy ? 'processing' : props.error ? 'error' : props.receipt?.status ?? 'confirm')
const { reducedMotion } = useMotionPreferences()

function close(): void {
  if (!props.busy) open.value = false
}

useTelegramBackButton(ownsBack, close)
</script>

<template>
  <UModal v-model:open="open" :title="$t('connections.unblockTitle')" :description="block ? $t('connections.unblockDescription', { ip: block.ip }) : ''" :dismissible="!busy" :close="false" :ui="{ header: 'tg-overlay-header--centered', wrapper: 'tg-overlay-copy--centered', footer: 'justify-end' }">
    <template #body>
      <AnimatePresence mode="wait" :initial="false">
        <motion.div
          v-if="block"
          :key="workflowState"
          class="connection-drop"
          :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 4 }"
          :animate="{ opacity: 1, y: 0 }"
          :exit="{ opacity: 0 }"
          :transition="{ duration: reducedMotion ? 0.08 : 0.18, ease: 'easeOut' }"
        >
          <div class="connection-unblock__target"><UIcon name="i-ph-shield-check" /><div><code>{{ block.ip }}</code><span>{{ $t('connections.expiresAt', { date: formatDateTime(block.expiresAt) }) }}</span></div></div>
          <InlineNotice v-if="receipt" :tone="receipt.status === 'succeeded' ? 'success' : 'warning'" :title="$t(`operations.status.${receipt.status}`)">{{ $t(`connections.unblockOperation.${receipt.status}`) }}</InlineNotice>
          <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
          <UButton v-if="receipt && error" color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :loading="checking" :label="$t('operations.checkStatus')" data-haptic="retry" @click="emit('refresh')" />
        </motion.div>
      </AnimatePresence>
    </template>
    <template #footer>
      <UButton color="neutral" variant="outline" :disabled="busy" :label="$t('common.close')" data-haptic="dismiss" @click="close" />
      <UButton v-if="!receipt" color="primary" icon="i-ph-shield-check" :loading="busy" :disabled="busy || !block" :label="busy ? $t('connections.unblocking') : $t('connections.unblock')" data-haptic="confirm" @click="emit('confirm')" />
    </template>
  </UModal>
</template>
