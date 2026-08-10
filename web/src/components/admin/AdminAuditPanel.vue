<script setup lang="ts">
import { shallowRef } from 'vue'

import type { AuditEvent } from '@/api/types'
import { useAdminSection } from '@/composables/useAdminSection'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'
import AdminSectionState from './AdminSectionState.vue'

const { items, loading, error, load } = useAdminSection<AuditEvent>('audit-events')
const action = shallowRef('')
const { t } = useI18n()

function detailText(detail: AuditEvent['detail']): string {
  return typeof detail === 'string' ? detail : JSON.stringify(detail)
}

function applyFilter(): void {
  void load({ action: action.value || undefined })
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminAudit.title') }}</h2><p>{{ t('adminAudit.copy') }}</p></div>
      <form class="admin-search" @submit.prevent="applyFilter">
        <UInput v-model.trim="action" icon="i-ph-funnel" :placeholder="t('adminAudit.filter')" :aria-label="t('adminAudit.filterLabel')" />
      </form>
    </div>
    <AdminSectionState :loading="loading" :error="error" @retry="load()">
      <div v-auto-animate class="timeline-list">
        <article v-for="event in items" :key="event.id" class="timeline-row">
          <span class="timeline-row__icon"><UIcon name="i-ph-shield-check" /></span>
          <div>
            <strong>{{ event.action }}</strong>
            <p>{{ detailText(event.detail) }}</p>
            <small>{{ t('adminAudit.meta', { actor: event.actorUserId ?? t('adminAudit.system'), target: `${event.targetType} ${event.targetId}`, date: formatDateTime(event.createdAt) }) }}</small>
          </div>
        </article>
        <div v-if="!items.length" class="empty-inline"><div><h3>{{ t('adminAudit.none') }}</h3><p>{{ t('adminAudit.noneHint') }}</p></div></div>
      </div>
    </AdminSectionState>
  </section>
</template>
