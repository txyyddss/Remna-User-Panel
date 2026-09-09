<script setup lang="ts">
import { restoreItems } from '@/api/cache/restore'
import { onMounted, shallowRef } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { EmbyAccount } from '@/api/features'
import { featuresApi } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import OperationStatusNotice from '@/components/common/OperationStatusNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useDurableCommand } from '@/composables/useDurableCommand'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { localizedError, useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const items = shallowRef<EmbyAccount[]>([])
const loading = shallowRef(true)
const error = shallowRef<string | null>(null)
const retryCommand = useDurableCommand({ errorKey: 'adminEmby.retryFailed', onTerminal: () => load() })
const { reducedMotion, offset } = useMotionPreferences()

async function load(): Promise<void> {
  loading.value = !restoreItems('/api/v1/admin/emby-accounts', items)
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
    <motion.div v-else layout class="admin-list">
      <AnimatePresence :initial="false" mode="popLayout">
        <motion.article v-for="account in items" :key="account.id" layout class="admin-list-row admin-list-row--emby" :initial="reducedMotion ? false : { opacity: 0, y: offset(6) }" :animate="{ opacity: 1, y: 0 }" :exit="{ opacity: 0 }" :transition="{ duration: reducedMotion ? 0.08 : 0.18, ease: 'easeOut' }">
          <span class="feature-icon feature-icon--small"><UIcon name="i-ph-monitor-play" /></span>
          <div>
            <strong>{{ account.username }}</strong>
            <small>{{ t('adminEmby.summary', { rating: account.maxParentalRating === null ? t('emby.noRating') : account.maxParentalRating, count: account.disabledLibraryIds.length, date: formatDateTime(account.updatedAt) }) }}</small>
          </div>
          <StatusBadge :tone="account.status === 'active' ? 'success' : account.status === 'failed' ? 'danger' : 'warning'" :label="t(`emby.status.${account.status}`)" />
          <UButton v-if="account.retryable" size="sm" color="neutral" variant="outline" :disabled="retryCommand.blocksMutations.value" :loading="retryCommand.busy.value && retryCommand.activeCommandId.value === account.id" :label="retryCommand.busy.value && retryCommand.activeCommandId.value === account.id ? t('adminEmby.retrying') : t('adminEmby.retry')" @click="retry(account.id)" />
        </motion.article>
        <motion.div v-if="!items.length" key="empty" class="empty-inline" :initial="{ opacity: 0 }" :animate="{ opacity: 1 }"><div><h3>{{ t('adminEmby.none') }}</h3><p>{{ t('adminEmby.noneHint') }}</p></div></motion.div>
      </AnimatePresence>
    </motion.div>
  </section>
</template>
