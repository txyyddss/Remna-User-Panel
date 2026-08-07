<script setup lang="ts">
import { PhArchive, PhArrowClockwise, PhCheckCircle, PhDatabase } from '@phosphor-icons/vue'

import type { BackupRecord, JobRecord } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { formatBytes, formatDateTime } from '@/utils/format'
import AdminSectionState from './AdminSectionState.vue'

const backups = useAdminSection<BackupRecord>('backups')
const jobs = useAdminSection<JobRecord>('jobs')

function createBackup(): void {
  void backups.create({ action: 'create' })
}

function retryJob(id: string): void {
  void jobs.perform(() => import('@/api/client').then(({ api }) => api.retryAdminJob(id)))
}

function backupName(backup: BackupRecord): string {
  return backup.path.split(/[\\/]/).pop() ?? `Backup ${backup.id}`
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>Backups</h2><p>Verified SQLite snapshots are retained for seven days.</p></div>
      <button class="button button--primary" type="button" :disabled="backups.busy.value" @click="createBackup"><PhArchive :size="18" /> Create backup</button>
    </div>
    <AdminSectionState :loading="backups.loading.value" :error="backups.error.value" @retry="backups.load()">
      <div class="backup-grid">
        <article v-for="backup in backups.items.value" :key="backup.id" class="backup-card">
          <span class="feature-icon"><PhDatabase :size="22" /></span>
          <div><strong>{{ backupName(backup) }}</strong><small>{{ formatBytes(backup.sizeBytes) }} / {{ formatDateTime(backup.createdAt) }}</small></div>
          <span class="backup-card__verified"><PhCheckCircle :size="17" weight="fill" /> {{ backup.status === 'complete' ? 'Verified' : backup.status }}</span>
        </article>
        <div v-if="!backups.items.value.length" class="empty-inline"><div><h3>No backups yet</h3><p>The daily scheduler will create the first verified copy.</p></div></div>
      </div>
    </AdminSectionState>

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
