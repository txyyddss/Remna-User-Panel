<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'

import type { EmbyAccount } from '@/api/features'
import { featuresApi } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import OperationStatusNotice from '@/components/common/OperationStatusNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useDurableCommand } from '@/composables/useDurableCommand'
import { localizedError, useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const items = shallowRef<EmbyAccount[]>([])
const loading = shallowRef(true)
const error = shallowRef<string | null>(null)
const retryCommand = useDurableCommand({ errorKey: 'adminEmby.retryFailed', onTerminal: () => load() })

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try { items.value = (await featuresApi.getAdminEmbyAccounts()).items }
  catch (caught) { error.value = localizedError(caught, 'adminEmby.loadFailed') }
  finally { loading.value = false }
}

async function retry(id: string): Promise<void> {
  await retryCommand.execute(id, `emby-retry:${id}`, (key) => featuresApi.retryAdminEmbyAccount(id, key))
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminEmby.title') }}</h2><p>{{ t('adminEmby.copy') }}</p></div>
      <UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="t('common.refresh')" @click="load" />
    </div>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <OperationStatusNotice :receipt="retryCommand.receipt.value" :error="retryCommand.error.value" :checking="retryCommand.checking.value" @refresh="retryCommand.refresh" />
    <USkeleton v-if="loading" class="m-4 h-24" />
    <div v-else v-auto-animate class="admin-list">
      <article v-for="account in items" :key="account.id" class="admin-list-row">
        <span class="feature-icon feature-icon--small"><UIcon name="i-ph-monitor-play" /></span>
        <div>
          <strong>{{ account.username }}</strong>
          <small>{{ t('adminEmby.summary', { rating: account.maxParentalRating === null ? t('emby.noRating') : account.maxParentalRating, count: account.disabledLibraryIds.length, date: formatDateTime(account.updatedAt) }) }}</small>
        </div>
        <StatusBadge :tone="account.status === 'active' ? 'success' : account.status === 'failed' ? 'danger' : 'warning'" :label="t(`emby.status.${account.status}`)" />
        <UButton v-if="account.retryable" size="sm" color="neutral" variant="outline" :disabled="retryCommand.blocksMutations.value" :loading="retryCommand.busy.value && retryCommand.activeCommandId.value === account.id" :label="retryCommand.busy.value && retryCommand.activeCommandId.value === account.id ? t('adminEmby.retrying') : t('adminEmby.retry')" @click="retry(account.id)" />
      </article>
      <div v-if="!items.length" class="empty-inline"><div><h3>{{ t('adminEmby.none') }}</h3><p>{{ t('adminEmby.noneHint') }}</p></div></div>
    </div>
  </section>
</template>
