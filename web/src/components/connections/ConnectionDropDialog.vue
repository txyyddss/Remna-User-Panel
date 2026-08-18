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
const statusTone = computed(() => {
  const status = props.receipt?.status
  if (status === 'succeeded') return 'success'
  return status === 'queued' || status === 'processing' ? 'info' : 'warning'
})
const statusMessage = computed(() => t(`connections.operation.${props.receipt?.status ?? 'queued'}`))

function statusLabel(status: OperationStatus): string {
  return t(`operations.status.${status}`)
}

function close(): void {
  if (!props.busy) open.value = false
}

useTelegramBackButton(ownsBack, close)
</script>

<template>
  <UModal
    v-model:open="open"
    :title="$t('connections.dropTitle')"
    :description="target ? $t('connections.dropDescription', { ip: target.connection.ip }) : ''"
    :dismissible="!busy"
    :ui="{ footer: 'justify-end' }"
  >
    <template #body>
      <div v-if="target" class="connection-drop">
        <div class="connection-drop__target">
          <CountryFlag :code="target.countryCode" />
          <div><strong>{{ target.nodeName }}</strong><code>{{ target.connection.ip }}</code></div>
        </div>
        <InlineNotice tone="warning">{{ $t('connections.sharedIpWarning') }}</InlineNotice>
        <InlineNotice v-if="receipt" :tone="statusTone" :title="statusLabel(receipt.status)">
          {{ statusMessage }}
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
    <template #footer>
      <UButton color="neutral" variant="outline" :disabled="busy" :label="$t('common.close')" @click="close" />
      <UButton
        v-if="!receipt"
        color="error"
        icon="i-ph-plugs-connected"
        :loading="busy"
        :disabled="busy || !target"
        :label="busy ? $t('connections.disconnecting') : $t('connections.disconnect')"
        data-haptic="heavy"
        @click="emit('confirm')"
      />
    </template>
  </UModal>
</template>
