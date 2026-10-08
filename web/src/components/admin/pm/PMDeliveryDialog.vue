<script setup lang="ts">
import './modal.css'
import { onScopeDispose, shallowRef, watch } from 'vue'
import type { PMConversation, PMDelivery } from '@/api/types'
import { pmApi } from '@/api/pm'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { localizedError, useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'
import { createLatestRequest } from '@/utils/latestRequest'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { openExternalLink } from '@/utils/telegram'
import { pmTopicLink } from './links'

const open = defineModel<boolean>('open', { required: true })
const props = defineProps<{ conversation: PMConversation | null }>()
const { t } = useI18n()
const { reducedMotion } = useMotionPreferences()
const items = shallowRef<PMDelivery[]>([])
const loading = shallowRef(false)
const error = shallowRef<string | null>(null)
const latest = createLatestRequest()
async function load(): Promise<void> {
  if (!props.conversation) return
  const token = latest.begin()
  loading.value = true; error.value = null
  try { const result = await pmApi.deliveries(props.conversation.id); if (latest.isCurrent(token)) items.value = result.items }
  catch (caught) { if (latest.isCurrent(token)) error.value = localizedError(caught, 'pm.loadFailed') }
  finally { if (latest.isCurrent(token)) loading.value = false }
}
watch(open, next => { if (next) { items.value = []; void load() } else latest.invalidate() })
onScopeDispose(latest.dispose)
function view(item: PMDelivery): void {
  if (!props.conversation) return
  const link = pmTopicLink(props.conversation.chatId, item.topicId, item.direction === 'outbound' ? item.sourceMessageId : item.resultMessageId)
  if (link) openExternalLink(link)
}
</script>

<template>
  <UModal v-model:open="open" :title="t('pm.deliveries')" :description="t('pm.reviewHint')" :ui="{ content: reducedMotion ? 'pm-modal--reduced' : '', overlay: reducedMotion ? 'pm-modal--reduced' : '' }">
    <template #body>
      <USkeleton v-if="loading" class="h-24" />
      <div v-else-if="error"><InlineNotice tone="warning">{{ error }}</InlineNotice><UButton color="neutral" variant="ghost" :label="t('pm.retry')" @click="load" /></div>
      <p v-else-if="!items.length" class="pm-deliveries__empty">{{ t('pm.noDeliveries') }}</p>
      <div v-else class="pm-deliveries">
        <article v-for="item in items" :key="item.operationId" class="pm-delivery">
          <div><strong>{{ t(`pm.direction.${item.direction}`) }}</strong><small>{{ formatDateTime(item.createdAt) }} · {{ t('pm.sourceMessage', { id: item.sourceMessageId }) }}</small></div>
          <p>{{ t(`operations.status.${item.status}`) }}</p>
          <InlineNotice v-if="item.errorCode" tone="warning">{{ t(`pm.errors.${item.errorCode}`) === `pm.errors.${item.errorCode}` ? t('pm.deliveryFailed') : t(`pm.errors.${item.errorCode}`) }}</InlineNotice>
          <UButton v-if="item.topicId && (item.direction === 'outbound' || item.resultMessageId)" color="neutral" variant="ghost" icon="i-ph-arrow-square-out" :label="t('pm.viewInTelegram')" @click="view(item)" />
        </article>
      </div>
    </template>
  </UModal>
</template>

<style scoped>
.pm-deliveries { display: grid; gap: 1rem; }
.pm-delivery { display: grid; gap: 0.6rem; padding-block: 0.8rem; border-bottom: 1px solid var(--line); min-width: 0; }
.pm-delivery small { display: block; color: var(--text-muted); margin-top: 0.35rem; overflow-wrap: anywhere; }
.pm-delivery p { margin: 0; font-size: 0.8rem; color: var(--text-muted); }
.pm-deliveries__empty { color: var(--text-muted); }
</style>
