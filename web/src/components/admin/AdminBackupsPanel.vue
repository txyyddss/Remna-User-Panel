<script setup lang="ts">
import { onScopeDispose, shallowRef } from 'vue'

import type { BackupRecord, JobRecord } from '@/api/types'
import type { RestoreOperation } from '@/api/features'
import { featuresApi } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import OperationStatusNotice from '@/components/common/OperationStatusNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { useDurableCommand } from '@/composables/useDurableCommand'
import { localizedError, useI18n } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { formatBytes, formatDateTime } from '@/utils/format'
import AdminSectionState from './AdminSectionState.vue'
import RestoreBackupDialog from './backups/RestoreBackupDialog.vue'
import BackupUploadPanel from './backups/BackupUploadPanel.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const backups = useAdminSection<BackupRecord>('backups')
const jobs = useAdminSection<JobRecord>('jobs')
const jobRetry = useDurableCommand({ errorKey: 'errors.adminAction', onTerminal: (receipt) => receipt.status === 'succeeded' ? jobs.load() : undefined })
const restoring = shallowRef(false)
const restoreTarget = shallowRef<BackupRecord | null>(null)
const restoreOperation = shallowRef<RestoreOperation | null>(null)
const actionError = shallowRef<string | null>(null)
const backupDeleteTarget = shallowRef<BackupRecord | null>(null)
const jobDeleteTarget = shallowRef<JobRecord | null>(null)
const { t } = useI18n()
let restorePollTimer: ReturnType<typeof globalThis.setTimeout> | undefined
let restoreAttempt: { fingerprint: string; key: string } | undefined
function isScheduled(job: JobRecord): boolean { return job.status === 'pending' && new Date(job.availableAt).getTime() > Date.now() }
function restoreKey(fingerprint: string): string {
  if (restoreAttempt?.fingerprint !== fingerprint) restoreAttempt = { fingerprint, key: createUuid() }
  return restoreAttempt.key
}
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
  void backups.perform(() => import('@/api/client').then(({ api }) => api.createAdminBackup()))
}

function retryJob(id: string): void {
  void jobRetry.execute(id, `job-retry:${id}`, (key) => import('@/api/client').then(({ api }) => api.retryAdminJob(id, key)))
}
async function deleteBackup(): Promise<void> {
  if (!backupDeleteTarget.value) return
  actionError.value = null
  try {
    await featuresApi.deleteBackup(backupDeleteTarget.value.id)
    backupDeleteTarget.value = null
    await backups.load()
  } catch (caught) {
    actionError.value = localizedError(caught, 'adminBackups.deleteFailed')
  }
}

async function deleteJob(): Promise<void> {
  if (!jobDeleteTarget.value) return
  actionError.value = null
  try {
    const { api } = await import('@/api/client')
    await api.deleteAdminJob(jobDeleteTarget.value.id)
    jobDeleteTarget.value = null
    await jobs.load()
  } catch (caught) {
    actionError.value = localizedError(caught, 'adminBackups.jobDeleteFailed')
  }
}

function backupName(backup: BackupRecord): string {
  return backup.path.split(/[\\/]/).pop() ?? t('adminBackups.backupName', { id: backup.id })
}

function restoreFollowUp(operation: RestoreOperation): string {
  if (operation.status === 'complete') return t('adminBackups.reauthenticate')
  if (operation.status === 'failed') return t('adminBackups.operationError')
  return t('adminBackups.reconnect')
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
    actionError.value = localizedError(caught, 'adminBackups.downloadFailed')
  }
}

async function restore(payload: { reason: string; confirmation: string }): Promise<void> {
  if (!restoreTarget.value) return
  restoring.value = true
  actionError.value = null
  try {
    const fingerprint = JSON.stringify([restoreTarget.value.id, payload.reason, payload.confirmation])
    restoreOperation.value = await featuresApi.restoreBackup(restoreTarget.value.id, payload.reason, payload.confirmation, restoreKey(fingerprint))
    restoreTarget.value = null
    restoreAttempt = undefined
    scheduleRestorePoll()
  } catch (caught) {
    actionError.value = localizedError(caught, 'adminBackups.restoreFailed')
  } finally {
    restoring.value = false
  }
}

onScopeDispose(stopRestorePolling)
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminBackups.title') }}</h2><p>{{ t('adminBackups.copy') }}</p></div>
      <UButton icon="i-ph-archive" :disabled="backups.busy.value" :loading="backups.busy.value" :label="t('adminBackups.create')" @click="createBackup" />
    </div>
    <BackupUploadPanel @uploaded="backups.load()" />
    <InlineNotice v-if="restoreOperation" :tone="restoreOperation.status === 'failed' ? 'warning' : 'success'" :title="restoreOperation.status === 'complete' ? t('adminBackups.restoreComplete') : restoreOperation.status === 'failed' ? t('adminBackups.restoreFailedTitle') : t('adminBackups.restoreStaged')">{{ t('adminBackups.operationStatus', { id: restoreOperation.id, status: t(`adminBackups.status.${restoreOperation.status}`) }) }} {{ restoreFollowUp(restoreOperation) }}</InlineNotice>
    <InlineNotice v-if="actionError" tone="warning">{{ actionError }}</InlineNotice>
    <AdminSectionState :loading="backups.loading.value" :error="backups.error.value" @retry="backups.load()">
      <div v-auto-animate class="backup-grid">
        <article v-for="backup in backups.items.value" :key="backup.id" class="backup-card">
          <span class="feature-icon"><UIcon name="i-ph-database" /></span>
          <div><strong>{{ backupName(backup) }}</strong><small>{{ formatBytes(backup.sizeBytes) }} / {{ formatDateTime(backup.createdAt) }}</small></div>
          <span class="backup-card__verified"><UIcon name="i-ph-check-circle-fill" /> {{ backup.status === 'complete' ? t('adminBackups.verified') : t(`adminBackups.status.${backup.status}`) }}</span>
          <div v-if="backup.status === 'complete'" class="backup-card__actions"><UButton size="sm" color="neutral" variant="outline" icon="i-ph-download-simple" :label="t('adminBackups.download')" @click="download(backup)" /><UButton size="sm" color="error" variant="ghost" icon="i-ph-upload-simple" :label="t('adminBackups.restore')" data-haptic="destructive" @click="restoreTarget = backup" /><UButton size="sm" color="error" variant="ghost" icon="i-ph-trash" :label="t('adminBackups.delete')" data-haptic="destructive" @click="backupDeleteTarget = backup" /></div>
        </article>
        <div v-if="!backups.items.value.length" class="empty-inline"><div><h3>{{ t('adminBackups.none') }}</h3><p>{{ t('adminBackups.noneHint') }}</p></div></div>
      </div>
    </AdminSectionState>
    <RestoreBackupDialog :open="restoreTarget !== null" :backup-name="restoreTarget ? backupName(restoreTarget) : ''" :busy="restoring" @update:open="!$event && (restoreTarget = null)" @restore="restore" />
    <ConfirmDialog :open="Boolean(backupDeleteTarget)" :title="t('adminBackups.deleteTitle', { name: backupDeleteTarget ? backupName(backupDeleteTarget) : t('adminBackups.backup') })" :description="t('adminBackups.deleteDescription')" :confirm-label="t('adminBackups.deleteBackup')" danger @update:open="!$event && (backupDeleteTarget = null)" @confirm="deleteBackup" />

    <div class="admin-subsection-heading"><div><h3>{{ t('adminBackups.jobs') }}</h3><p>{{ t('adminBackups.jobsHint') }}</p></div><UButton color="neutral" variant="ghost" icon="i-ph-arrow-clockwise" :label="t('common.refresh')" @click="jobs.load()" /></div>
    <OperationStatusNotice :receipt="jobRetry.receipt.value" :error="jobRetry.error.value" :checking="jobRetry.checking.value" @refresh="jobRetry.refresh" />
    <AdminSectionState :loading="jobs.loading.value" :error="jobs.error.value" @retry="jobs.load()">
      <div v-auto-animate class="admin-list admin-list--compact">
        <article v-for="job in jobs.items.value" :key="job.id" class="admin-list-row">
          <div><strong>{{ job.kind }}</strong><small>{{ t('adminBackups.attempts', { count: job.attempts }) }}<template v-if="job.status === 'pending'"> / {{ t('adminBackups.availableAt', { date: formatDateTime(job.availableAt) }) }}</template><template v-else> / {{ formatDateTime(job.updatedAt) }}</template><template v-if="job.lastError"> / {{ t('adminBackups.jobError') }}</template></small></div>
          <StatusBadge :tone="job.status === 'done' ? 'success' : job.status === 'failed' ? 'danger' : 'warning'" :label="t(isScheduled(job) ? 'adminBackups.status.scheduled' : `adminBackups.status.${job.status}`)" />
          <UButton
            v-if="job.status === 'failed'"
            size="sm"
            color="neutral"
            variant="outline"
            icon="i-ph-arrow-clockwise"
            :disabled="jobs.busy.value || jobRetry.blocksMutations.value"
            :loading="jobRetry.busy.value && jobRetry.activeCommandId.value === job.id"
            :label="t('adminBackups.retry')"
            @click="retryJob(job.id)"
          />
          <UButton color="error" variant="ghost" icon="i-ph-trash" :disabled="job.status === 'processing' || jobs.busy.value || jobRetry.blocksMutations.value" :aria-label="t('adminBackups.deleteJobLabel', { kind: job.kind })" data-haptic="destructive" @click="jobDeleteTarget = job" />
        </article>
        <div v-if="!jobs.items.value.length" class="empty-inline"><div><h3>{{ t('adminBackups.noJobs') }}</h3><p>{{ t('adminBackups.noJobsHint') }}</p></div></div>
      </div>
      <UButton v-if="jobs.nextCursor.value" class="database-load-more" color="neutral" variant="outline" icon="i-ph-arrow-down" :loading="jobs.loading.value" :disabled="jobs.loading.value" :label="t('adminBackups.loadMoreJobs')" @click="jobs.loadMore" />
    </AdminSectionState>
    <ConfirmDialog :open="Boolean(jobDeleteTarget)" :title="t('adminBackups.deleteJobTitle')" :description="t('adminBackups.deleteJobDescription')" :confirm-label="t('adminBackups.deleteJob')" danger @update:open="!$event && (jobDeleteTarget = null)" @confirm="deleteJob" />
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
.backup-card > div { min-width: 0; }
.backup-card__actions :deep(button) { min-width: 0; flex: 1 1 8rem; }
@media (max-width: 639px) {
  .backup-card { align-items: start; }
  .backup-card strong { overflow-wrap: anywhere; white-space: normal; }
  .backup-card__verified { min-width: 0; overflow-wrap: anywhere; }
  .backup-card__actions :deep(button) { flex-basis: 100%; width: 100%; }
}
</style>
