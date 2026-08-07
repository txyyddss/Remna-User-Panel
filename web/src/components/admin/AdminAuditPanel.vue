<script setup lang="ts">
import { shallowRef } from 'vue'
import { PhFunnel, PhShieldCheck } from '@phosphor-icons/vue'

import type { AuditEvent } from '@/api/types'
import { useAdminSection } from '@/composables/useAdminSection'
import { formatDateTime } from '@/utils/format'
import AdminSectionState from './AdminSectionState.vue'

const { items, loading, error, load } = useAdminSection<AuditEvent>('audit-events')
const action = shallowRef('')

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
      <div><h2>Audit trail</h2><p>Append-only records for sensitive administrative actions.</p></div>
      <form class="admin-search" @submit.prevent="applyFilter"><PhFunnel :size="18" /><input v-model.trim="action" placeholder="Filter action" aria-label="Filter by action" /></form>
    </div>
    <AdminSectionState :loading="loading" :error="error" @retry="load()">
      <div class="timeline-list">
        <article v-for="event in items" :key="event.id" class="timeline-row">
          <span class="timeline-row__icon"><PhShieldCheck :size="18" /></span>
          <div><strong>{{ event.action }}</strong><p>{{ detailText(event.detail) }}</p><small>{{ event.actorUserId ?? 'System' }} / {{ event.targetType }} {{ event.targetId }} / {{ formatDateTime(event.createdAt) }}</small></div>
        </article>
        <div v-if="!items.length" class="empty-inline"><div><h3>No audit events</h3><p>Changes appear here as administrators use the console.</p></div></div>
      </div>
    </AdminSectionState>
  </section>
</template>
