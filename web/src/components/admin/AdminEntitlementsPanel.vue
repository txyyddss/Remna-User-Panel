<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { PhProhibit, PhStack } from '@phosphor-icons/vue'

import type { AdminEntitlement } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { formatDate, formatMoney } from '@/utils/format'
import AdminReasonDialog from './AdminReasonDialog.vue'
import AdminSectionState from './AdminSectionState.vue'

const entitlements = useAdminSection<AdminEntitlement>('entitlements')
const selected = shallowRef<AdminEntitlement | null>(null)
const reason = shallowRef('')
const statusFilter = shallowRef<'all' | AdminEntitlement['status']>('all')

const filtered = computed(() => statusFilter.value === 'all'
  ? entitlements.items.value
  : entitlements.items.value.filter((item) => item.status === statusFilter.value))

function canCancel(entitlement: AdminEntitlement): boolean {
  return entitlement.status === 'activating' || entitlement.status === 'active' || entitlement.status === 'queued'
}

function tone(status: AdminEntitlement['status']): 'neutral' | 'success' | 'warning' | 'danger' {
  if (status === 'active') return 'success'
  if (status === 'queued' || status === 'activating') return 'warning'
  if (status === 'failed') return 'danger'
  return 'neutral'
}

async function cancelEntitlement(): Promise<void> {
  if (!selected.value) return
  const id = selected.value.id
  const { api } = await import('@/api/client')
  const success = await entitlements.perform(() => api.cancelAdminEntitlement(id, reason.value))
  if (success) {
    selected.value = null
    reason.value = ''
  }
}

function setCancelOpen(open: boolean): void {
  if (open) return
  selected.value = null
  reason.value = ''
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div>
        <h2>Entitlements</h2>
        <p>Review current and historical access, including the owning account.</p>
      </div>
      <select v-model="statusFilter" class="compact-select" aria-label="Filter entitlements">
        <option value="all">All statuses</option>
        <option value="activating">Activating</option>
        <option value="active">Active</option>
        <option value="queued">Queued</option>
        <option value="expired">Expired</option>
        <option value="cancelled">Cancelled</option>
        <option value="failed">Failed</option>
      </select>
    </div>

    <AdminSectionState :loading="entitlements.loading.value" :error="entitlements.error.value" @retry="entitlements.load()">
      <div class="admin-list">
        <article v-for="entitlement in filtered" :key="entitlement.id" class="admin-list-row admin-list-row--entitlement">
          <span class="feature-icon feature-icon--small"><PhStack :size="19" /></span>
          <div>
            <strong>{{ entitlement.comboName }}</strong>
            <small>User {{ entitlement.userId }} / {{ formatDate(entitlement.validFrom) }} to {{ formatDate(entitlement.validUntil) }}</small>
          </div>
          <strong>{{ formatMoney(entitlement.price) }}</strong>
          <StatusBadge :tone="tone(entitlement.status)" :label="entitlement.status" />
          <button
            v-if="canCancel(entitlement)"
            class="button button--ghost-danger button--small"
            type="button"
            @click="selected = entitlement"
          >
            <PhProhibit :size="17" /> Cancel
          </button>
        </article>
        <div v-if="!filtered.length" class="empty-inline">
          <div><h3>No entitlements in this view</h3><p>Purchases appear here once users select a combo.</p></div>
        </div>
      </div>
    </AdminSectionState>

    <AdminReasonDialog
      :open="selected !== null"
      v-model:reason="reason"
      title="Cancel this entitlement?"
      description="The snapshotted TXB price will be credited back. Active access will be revoked through a synchronization job."
      confirm-label="Cancel entitlement"
      :busy="entitlements.busy.value"
      danger
      @update:open="setCancelOpen"
      @confirm="cancelEntitlement"
    />
  </section>
</template>
