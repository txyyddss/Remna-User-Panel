<script setup lang="ts">
import { onScopeDispose, shallowRef } from 'vue'
import { PhArchive, PhArrowClockwise, PhCheckCircle, PhDatabase, PhDownloadSimple, PhUploadSimple } from '@phosphor-icons/vue'

import type { BackupRecord, JobRecord } from '@/api/types'
import type { RestoreOperation } from '@/api/features'
import { featuresApi } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { formatBytes, formatDateTime } from '@/utils/format'
import AdminSectionState from './AdminSectionState.vue'
import RestoreBackupDialog from './backups/RestoreBackupDialog.vue'

const backups = useAdminSection<BackupRecord>('backups')
const jobs = useAdminSection<JobRecord>('jobs')
const restoring = shallowRef(false)
const restoreTarget = shallowRef<BackupRecord | null>(null)
const restoreOperation = shallowRef<RestoreOperation | null>(null)
const actionError = shallowRef<string | null>(null)
let restorePollTimer: ReturnType<typeof globalThis.setTimeout> | undefined

function stopRestorePolling(): void {
  if (restorePollTimer !== undefined) globalThis.clearTimeout(restorePollTimer)
  restorePollTimer = undefined
}

function scheduleRestorePoll(): void {
  stopRestorePolling()
  restorePollTimer = globalThis.setTimeout(() => void pollRestore(), 2000)
}

async function pollRestore(): Promise<void> {
  if (!restoreOperation.value || ['complete', 'failed'].includes(restoreOperation.value.status)) return
  try {
    restoreOperation.value = await featuresApi.getRestoreStatus(restoreOperation.value.id)
    if (['complete', 'failed'].includes(restoreOperation.value.status)) return
  } catch {
    // A short connection loss is expected while the service performs the swap.
  }
  scheduleRestorePoll()
}

function createBackup(): void {
  void backups.create({ action: 'create' })
}

function retryJob(id: string): void {
  void jobs.perform(() => import('@/api/client').then(({ api }) => api.retryAdminJob(id)))
}

function backupName(backup: BackupRecord): string {
  return backup.path.split(/[\\/]/).pop() ?? `Backup ${backup.id}`
}

async function download(backup: BackupRecord): Promise<void> {
  actionError.value = null
  try {
    const blob = await featuresApi.downloadBackup(backup.id)
    const url = globalThis.URL.createObjectURL(blob)
    const anchor = globalThis.document.createElement('a')
    anchor.href = url
    anchor.download = backupName(backup)
    anchor.click()
    globalThis.URL.revokeObjectURL(url)
  } catch (caught) {
    actionError.value = caught instanceof Error ? caught.message : 'Backup download failed.'
  }
}

async function restore(payload: { reason: string; confirmation: string }): Promise<void> {
  if (!restoreTarget.value) return
  restoring.value = true
  actionError.value = null
  try {
    restoreOperation.value = await featuresApi.restoreBackup(restoreTarget.value.id, payload.reason, payload.confirmation)
    restoreTarget.value = null
    scheduleRestorePoll()
  } catch (caught) {
    actionError.value = caught instanceof Error ? caught.message : 'Restore staging failed.'
  } finally {
    restoring.value = false
  }
}

onScopeDispose(stopRestorePolling)
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>Backups</h2><p>Verified SQLite snapshots are retained for seven days.</p></div>
      <button class="button button--primary" type="button" :disabled="backups.busy.value" @click="createBackup"><PhArchive :size="18" /> Create backup</button>
    </div>
    <InlineNotice v-if="restoreOperation" :tone="restoreOperation.status === 'failed' ? 'warning' : 'success'" :title="restoreOperation.status === 'complete' ? 'Restore complete' : restoreOperation.status === 'failed' ? 'Restore failed' : 'Restore staged'">Operation {{ restoreOperation.id }} is {{ restoreOperation.status }}. {{ restoreOperation.error || (restoreOperation.status === 'complete' ? 'Reauthenticate if your previous session no longer exists.' : 'The app will reconnect after restart.') }}</InlineNotice>
    <InlineNotice v-if="actionError" tone="warning">{{ actionError }}</InlineNotice>
    <AdminSectionState :loading="backups.loading.value" :error="backups.error.value" @retry="backups.load()">
      <div class="backup-grid">
        <article v-for="backup in backups.items.value" :key="backup.id" class="backup-card">
          <span class="feature-icon"><PhDatabase :size="22" /></span>
          <div><strong>{{ backupName(backup) }}</strong><small>{{ formatBytes(backup.sizeBytes) }} / {{ formatDateTime(backup.createdAt) }}</small></div>
          <span class="backup-card__verified"><PhCheckCircle :size="17" weight="fill" /> {{ backup.status === 'complete' ? 'Verified' : backup.status }}</span>
          <div v-if="backup.status === 'complete'" class="backup-card__actions"><button class="button button--secondary button--small" type="button" @click="download(backup)"><PhDownloadSimple :size="17" />Download</button><button class="button button--ghost-danger button--small" type="button" @click="restoreTarget = backup"><PhUploadSimple :size="17" />Restore</button></div>
        </article>
        <div v-if="!backups.items.value.length" class="empty-inline"><div><h3>No backups yet</h3><p>The daily scheduler will create the first verified copy.</p></div></div>
      </div>
    </AdminSectionState>
    <RestoreBackupDialog :open="restoreTarget !== null" :backup-name="restoreTarget ? backupName(restoreTarget) : ''" :busy="restoring" @update:open="!$event && (restoreTarget = null)" @restore="restore" />

    <div class="admin-subsection-heading"><div><h3>Synchronization jobs</h3><p>Outbox retries and scheduled entitlement changes.</p></div><button class="text-button" type="button" @click="jobs.load()"><PhArrowClockwise :size="17" /> Refresh</button></div>
    <AdminSectionState :loading="jobs.loading.value" :error="jobs.error.value" @retry="jobs.load()">
      <div class="admin-list admin-list--compact">
        <article v-for="job in jobs.items.value.slice(0, 12)" :key="job.id" class="admin-list-row">
          <div><strong>{{ job.kind }}</strong><small>{{ job.attempts }} attempt{{ job.attempts === 1 ? '' : 's' }} / {{ formatDateTime(job.createdAt) }}{{ job.lastError ? ` / ${job.lastError}` : '' }}</small></div>
          <StatusBadge :tone="job.status === 'done' ? 'success' : job.status === 'failed' ? 'danger' : 'warning'" :label="job.status" />
          <button
            v-if="job.status === 'failed'"
            class="button button--secondary button--small"
            type="button"
            :disabled="jobs.busy.value"
            @click="retryJob(job.id)"
          >
            <PhArrowClockwise :size="17" /> Retry
          </button>
        </article>
        <div v-if="!jobs.items.value.length" class="empty-inline"><div><h3>No synchronization jobs</h3><p>Scheduled work and retries will appear here.</p></div></div>
      </div>
    </AdminSectionState>
  </section>
</template>

<style scoped>
.backup-card__actions {
  grid-column: 1 / -1;
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  padding-top: 0.6rem;
  border-top: 1px solid var(--line);
}
</style>
