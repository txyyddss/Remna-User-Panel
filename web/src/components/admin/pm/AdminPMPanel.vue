<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { pmApi } from '@/api/pm'
import type { PMConversation } from '@/api/types'
import SwitchField from '@/components/common/SwitchField.vue'
import OperationStatusNotice from '@/components/common/OperationStatusNotice.vue'
import AdminSectionState from '@/components/admin/AdminSectionState.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { useDurableCommand } from '@/composables/useDurableCommand'
import { useI18n } from '@/i18n'
import { openExternalLink } from '@/utils/telegram'
import PMDeliveryDialog from './PMDeliveryDialog.vue'
import PMTopicRepairDialog from './PMTopicRepairDialog.vue'
import { pmTopicLink } from './links'

const { t } = useI18n()
const inventory = useAdminSection<PMConversation>('pm')
const search = shallowRef('')
const selected = shallowRef<PMConversation | null>(null)
const reviewOpen = shallowRef(false)
const repairOpen = shallowRef(false)
const command = useDurableCommand({ errorKey: 'pm.actionFailed', onTerminal: () => inventory.load() })
const busy = computed(() => command.busy.value)
const message = computed(() => t(command.receipt.value?.status === 'succeeded' ? 'pm.updated' : command.receipt.value?.status === 'pending_review' ? 'pm.reviewHint' : command.receipt.value?.status === 'failed' ? 'pm.actionFailed' : 'pm.pending'))
function begin(): boolean { if (busy.value) return false; if (command.blocksMutations.value) command.reset(); return true }
async function moderate(item: PMConversation, key: 'blocked' | 'muted', value: boolean): Promise<void> {
  if (!begin()) return
  await command.execute(item.id, `${item.id}:${key}:${value}`, idempotencyKey => pmApi.moderate(item.id, { [key]: value }, idempotencyKey))
  await inventory.load()
}
function review(item: PMConversation): void { selected.value = item; reviewOpen.value = true }
function repair(item: PMConversation): void { selected.value = item; repairOpen.value = true }
async function repairTopic(body: { topicId: string; profileMessageId?: string }): Promise<void> {
  if (!selected.value || !begin()) return
  const id = selected.value.id
  if (await command.execute(id, `${id}:${JSON.stringify(body)}`, key => pmApi.repair(id, body, key))) repairOpen.value = false
}
function openTopic(item: PMConversation): void { const link = pmTopicLink(item.chatId, item.topicId); if (link) openExternalLink(link) }
</script>

<template>
  <section class="admin-panel pm-panel">
    <div class="admin-panel__heading"><div><h2>{{ t('pm.title') }}</h2><p>{{ t('pm.description') }}</p></div><UButton color="neutral" variant="ghost" icon="i-ph-arrow-clockwise" :label="t('pm.refresh')" :disabled="inventory.loading.value" @click="inventory.load()" /></div>
    <form class="pm-search" @submit.prevent="inventory.load({ search })"><UFormField :label="t('pm.search')"><UInput v-model="search" class="w-full" icon="i-ph-magnifying-glass" :maxlength="100" /></UFormField><UButton type="submit" color="neutral" :label="t('pm.searchButton')" :disabled="inventory.loading.value" /></form>
    <OperationStatusNotice :receipt="command.receipt.value" :error="command.error.value" :checking="command.checking.value" :message="message" @refresh="command.refresh" />
    <AdminSectionState :loading="inventory.loading.value" :error="inventory.error.value" @retry="inventory.load()">
      <div v-if="!inventory.items.value.length" class="empty-state"><h3>{{ t('pm.empty') }}</h3><p>{{ t('pm.emptyHint') }}</p></div>
      <div v-else class="pm-conversations">
        <article v-for="item in inventory.items.value" :key="item.id" class="pm-conversation" :aria-labelledby="`pm-name-${item.id}`">
          <div class="pm-conversation__identity"><h3 :id="`pm-name-${item.id}`">{{ item.name }}</h3><p>{{ item.username ? `@${item.username}` : item.telegramId }}</p><small>{{ t(`pm.topicState.${item.topicState}`) }}</small><small v-if="item.profileState === 'pending_review'">{{ t('pm.errors.PM_PROFILE_UNCERTAIN') }}</small></div>
          <SwitchField :id="`pm-block-${item.id}`" :model-value="item.blocked" :label="t('pm.blocked')" :disabled="busy" @update:model-value="moderate(item, 'blocked', $event)" />
          <SwitchField :id="`pm-mute-${item.id}`" :model-value="item.muted" :label="t('pm.muted')" :help="t('pm.muteHint')" :disabled="busy" @update:model-value="moderate(item, 'muted', $event)" />
          <div class="pm-conversation__actions"><UButton color="neutral" variant="ghost" :label="t('pm.deliveries')" @click="review(item)" /><UButton v-if="item.topicId" color="neutral" variant="ghost" icon="i-ph-arrow-square-out" :label="t('pm.openTopic')" @click="openTopic(item)" /><UButton v-if="item.topicState === 'pending_review' || item.profileState === 'pending_review'" color="neutral" variant="outline" :label="t('pm.repair')" :disabled="busy" @click="repair(item)" /></div>
        </article>
      </div>
      <UButton v-if="inventory.nextCursor.value" color="neutral" variant="outline" :label="t('pm.loadMore')" @click="inventory.loadMore()" />
    </AdminSectionState>
    <PMDeliveryDialog v-model:open="reviewOpen" :conversation="selected" />
    <PMTopicRepairDialog v-model:open="repairOpen" :conversation="selected" :busy="busy" @confirm="repairTopic" />
  </section>
</template>

<style scoped>
.pm-panel { display: grid; gap: 1rem; }
.pm-search { display: flex; align-items: end; gap: 0.75rem; padding-inline: 1rem; }
.pm-search > :first-child { flex: 1; min-width: 0; }
.pm-conversations { padding-inline: 1rem; container-type: inline-size; }
.pm-conversation { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 0.5rem 1rem; padding-block: 1rem; border-bottom: 1px solid var(--line); min-width: 0; }
.pm-conversation__identity,.pm-conversation__actions { grid-column: 1/-1; min-width: 0; }
.pm-conversation__identity h3 { margin: 0; font-size: 0.9rem; overflow-wrap: anywhere; }
.pm-conversation__identity p,.pm-conversation__identity small { margin: 0.3rem 0 0; color: var(--text-muted); font-size: 0.78rem; overflow-wrap: anywhere; }
.pm-conversation__actions { display: flex; flex-wrap: wrap; gap: 0.5rem; }
@container(min-width:850px) { .pm-conversation { grid-template-columns: minmax(10rem,1fr) 8rem 10rem auto; align-items: center; gap: 1.25rem; }.pm-conversation__identity,.pm-conversation__actions { grid-column: auto; } }
</style>
