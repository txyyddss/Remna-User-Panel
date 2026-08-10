<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'
import { PhArrowClockwise, PhMonitorPlay } from '@phosphor-icons/vue'

import type { EmbyAccount } from '@/api/features'
import { featuresApi } from '@/api/features'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { formatDateTime } from '@/utils/format'

const items = shallowRef<EmbyAccount[]>([])
const loading = shallowRef(true)
const busyId = shallowRef<string | null>(null)
const error = shallowRef<string | null>(null)

async function load(): Promise<void> {
  loading.value = true
  try { items.value = (await featuresApi.getAdminEmbyAccounts()).items } catch (caught) { error.value = caught instanceof Error ? caught.message : 'Emby accounts could not be loaded.' } finally { loading.value = false }
}

async function retry(id: string): Promise<void> {
  busyId.value = id
  try { await featuresApi.retryAdminEmbyAccount(id); await load() } catch (caught) { error.value = caught instanceof Error ? caught.message : 'Provisioning retry failed.' } finally { busyId.value = null }
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading"><div><h2>Emby accounts</h2><p>Provisioning state and safe reconciliation controls.</p></div><button class="button button--secondary" type="button" @click="load"><PhArrowClockwise :size="18" />Refresh</button></div>
    <p v-if="error" class="field-error admin-error">{{ error }}</p>
    <div v-if="loading" class="admin-loading">Loading Emby accounts</div>
    <div v-else class="admin-list"><article v-for="account in items" :key="account.id" class="admin-list-row"><span class="feature-icon feature-icon--small"><PhMonitorPlay :size="18" /></span><div><strong>{{ account.username }}</strong><small>{{ account.maxParentalRating === null ? 'No rating ceiling' : `Rating value ${account.maxParentalRating}` }} · {{ account.disabledLibraryIds.length }} blocked libraries · updated {{ formatDateTime(account.updatedAt) }}</small></div><StatusBadge :tone="account.status === 'active' ? 'success' : account.status === 'failed' ? 'danger' : 'warning'" :label="account.status" /><button v-if="account.retryable" class="button button--secondary button--small" type="button" :disabled="busyId === account.id" @click="retry(account.id)">{{ busyId === account.id ? 'Retrying' : 'Retry safely' }}</button></article><div v-if="!items.length" class="empty-inline"><div><h3>No Emby accounts</h3><p>Member setup records will appear here.</p></div></div></div>
  </section>
</template>

<style scoped>.admin-error { margin: 1rem; }</style>
