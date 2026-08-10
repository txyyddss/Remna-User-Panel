<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'

import type { EmbyAccount } from '@/api/features'
import { featuresApi } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { localizedError, useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const items = shallowRef<EmbyAccount[]>([])
const loading = shallowRef(true)
const busyId = shallowRef<string | null>(null)
const error = shallowRef<string | null>(null)

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try { items.value = (await featuresApi.getAdminEmbyAccounts()).items }
  catch (caught) { error.value = localizedError(caught, 'adminEmby.loadFailed') }
  finally { loading.value = false }
}

async function retry(id: string): Promise<void> {
  busyId.value = id
  try { await featuresApi.retryAdminEmbyAccount(id); await load() }
  catch (caught) { error.value = localizedError(caught, 'adminEmby.retryFailed') }
  finally { busyId.value = null }
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
    <USkeleton v-if="loading" class="m-4 h-24" />
    <div v-else v-auto-animate class="admin-list">
      <article v-for="account in items" :key="account.id" class="admin-list-row">
        <span class="feature-icon feature-icon--small"><UIcon name="i-ph-monitor-play" /></span>
        <div>
          <strong>{{ account.username }}</strong>
          <small>{{ t('adminEmby.summary', { rating: account.maxParentalRating === null ? t('emby.noRating') : account.maxParentalRating, count: account.disabledLibraryIds.length, date: formatDateTime(account.updatedAt) }) }}</small>
        </div>
        <StatusBadge :tone="account.status === 'active' ? 'success' : account.status === 'failed' ? 'danger' : 'warning'" :label="t(`emby.status.${account.status}`)" />
        <UButton v-if="account.retryable" size="sm" color="neutral" variant="outline" :loading="busyId === account.id" :label="busyId === account.id ? t('adminEmby.retrying') : t('adminEmby.retry')" @click="retry(account.id)" />
      </article>
      <div v-if="!items.length" class="empty-inline"><div><h3>{{ t('adminEmby.none') }}</h3><p>{{ t('adminEmby.noneHint') }}</p></div></div>
    </div>
  </section>
</template>
