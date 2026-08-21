<script setup lang="ts">
import { computed } from 'vue'

import type { OperationReceipt, OperationStatus } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import CountryFlag from '@/components/common/CountryFlag.vue'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { t } from '@/i18n'
import type { ConnectionTarget } from './types'

const open = defineModel<boolean>('open', { required: true })
const props = withDefaults(defineProps<{
  target: ConnectionTarget | null
  receipt?: OperationReceipt | null
  busy?: boolean
  checking?: boolean
  error?: string | null
}>(), { receipt: null, busy: false, checking: false, error: null })
const emit = defineEmits<{ confirm: []; refresh: [] }>()

const ownsBack = computed(() => open.value)
const statusTone = computed(() => props.receipt?.status === 'succeeded' ? 'success'
  : props.receipt?.status === 'queued' || props.receipt?.status === 'processing' ? 'info' : 'warning')

function close(): void {
  if (!props.busy) open.value = false
}

function statusLabel(status: OperationStatus): string {
  return t(`operations.status.${status}`)
}

useTelegramBackButton(ownsBack, close)
</script>

<template>
  <UModal v-model:open="open" :title="$t('connections.blockTitle')" :description="target ? $t('connections.blockDescription', { ip: target.connection.ip }) : ''" :dismissible="!busy" :close="{ 'data-haptic': 'dismiss' }" :ui="{ footer: 'justify-end' }">
    <template #body>
      <div v-if="target" class="connection-drop">
        <div class="connection-drop__target">
          <CountryFlag :code="target.countryCode" />
          <div><strong>{{ target.nodeName }}</strong><code>{{ target.connection.ip }}</code></div>
        </div>
        <InlineNotice tone="warning">{{ $t('connections.sharedIpWarning') }}</InlineNotice>
        <InlineNotice tone="info">{{ $t('connections.expiryWarning') }}</InlineNotice>
        <InlineNotice v-if="receipt" :tone="statusTone" :title="statusLabel(receipt.status)">{{ $t(`connections.blockOperation.${receipt.status}`) }}</InlineNotice>
        <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
        <UButton v-if="receipt && error" color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :loading="checking" :label="$t('operations.checkStatus')" data-haptic="retry" @click="emit('refresh')" />
      </div>
    </template>
    <template #footer>
      <UButton color="neutral" variant="outline" :disabled="busy" :label="$t('common.close')" data-haptic="dismiss" @click="close" />
      <UButton v-if="!receipt" color="error" icon="i-ph-shield-warning" :loading="busy" :disabled="busy || !target" :label="busy ? $t('connections.blocking') : $t('connections.block')" data-haptic="destructive" @click="emit('confirm')" />
    </template>
  </UModal>
</template>
