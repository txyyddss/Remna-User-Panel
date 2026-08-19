<script setup lang="ts">
import type { IPBlock } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { formatDateTime } from '@/utils/format'

withDefaults(defineProps<{ items: readonly IPBlock[]; loading?: boolean; error?: string | null; disabled?: boolean }>(), {
  loading: false, error: null, disabled: false,
})
const emit = defineEmits<{ unblock: [block: IPBlock]; refresh: [] }>()
</script>

<template>
  <section class="connection-blocks" aria-labelledby="connection-blocks-title">
    <header class="connection-blocks__heading">
      <div><p class="eyebrow">{{ $t('connections.blocksEyebrow') }}</p><h2 id="connection-blocks-title">{{ $t('connections.blocksTitle') }}</h2></div>
      <UButton v-if="error" color="neutral" variant="ghost" square icon="i-ph-arrow-clockwise" :aria-label="$t('common.tryAgain')" @click="emit('refresh')" />
    </header>
    <div v-if="loading" class="connection-blocks__loading"><USkeleton class="h-16" /><USkeleton class="h-16" /></div>
    <p v-else-if="error" class="connection-blocks__empty">{{ error }}</p>
    <p v-else-if="!items.length" class="connection-blocks__empty">{{ $t('connections.blocksEmpty') }}</p>
    <div v-else class="connection-blocks__list">
      <article v-for="block in items" :key="block.id" class="connection-block-row">
        <UIcon name="i-ph-shield-warning" aria-hidden="true" />
        <div><code>{{ block.ip }}</code><span>{{ $t('connections.expiresAt', { date: formatDateTime(block.expiresAt) }) }}</span></div>
        <StatusBadge :tone="block.status === 'active' ? 'success' : 'warning'" :label="$t(`connections.blockStatus.${block.status}`)" />
        <UButton color="neutral" variant="ghost" square icon="i-ph-shield-check" :disabled="disabled || block.status === 'unblocking'" :aria-label="$t('connections.unblockIp', { ip: block.ip })" data-haptic @click="emit('unblock', block)" />
      </article>
    </div>
  </section>
</template>
