<script setup lang="ts">
import { computed } from 'vue'

import type { IPBlock, OperationReceipt } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
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

function close(): void {
  if (!props.busy) open.value = false
}

useTelegramBackButton(ownsBack, close)
</script>

<template>
  <UModal v-model:open="open" :title="$t('connections.unblockTitle')" :description="block ? $t('connections.unblockDescription', { ip: block.ip }) : ''" :dismissible="!busy" :ui="{ footer: 'justify-end' }">
    <template #body>
      <div v-if="block" class="connection-drop">
        <div class="connection-unblock__target"><UIcon name="i-ph-shield-check" /><div><code>{{ block.ip }}</code><span>{{ $t('connections.expiresAt', { date: formatDateTime(block.expiresAt) }) }}</span></div></div>
        <InlineNotice v-if="receipt" :tone="receipt.status === 'succeeded' ? 'success' : 'warning'" :title="$t(`operations.status.${receipt.status}`)">{{ $t(`connections.unblockOperation.${receipt.status}`) }}</InlineNotice>
        <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
        <UButton v-if="receipt && error" color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :loading="checking" :label="$t('operations.checkStatus')" @click="emit('refresh')" />
      </div>
    </template>
    <template #footer>
      <UButton color="neutral" variant="outline" :disabled="busy" :label="$t('common.close')" @click="close" />
      <UButton v-if="!receipt" color="primary" icon="i-ph-shield-check" :loading="busy" :disabled="busy || !block" :label="busy ? $t('connections.unblocking') : $t('connections.unblock')" data-haptic @click="emit('confirm')" />
    </template>
  </UModal>
</template>
