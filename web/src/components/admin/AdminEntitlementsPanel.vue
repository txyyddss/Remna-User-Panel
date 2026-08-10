<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { PhProhibit, PhStack } from '@phosphor-icons/vue'

import type { AdminEntitlement } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { useI18n } from '@/i18n'
import { formatDate, formatMoney } from '@/utils/format'
import AdminReasonDialog from './AdminReasonDialog.vue'
import AdminSectionState from './AdminSectionState.vue'

const entitlements = useAdminSection<AdminEntitlement>('entitlements')
const selected = shallowRef<AdminEntitlement | null>(null)
const reason = shallowRef('')
const statusFilter = shallowRef<'all' | AdminEntitlement['status']>('all')
const { t } = useI18n()

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
        <h2>{{ t('adminEntitlements.title') }}</h2>
        <p>{{ t('adminEntitlements.copy') }}</p>
      </div>
      <select v-model="statusFilter" class="compact-select" :aria-label="t('adminEntitlements.filter')">
        <option value="all">{{ t('adminEntitlements.all') }}</option>
        <option value="activating">{{ t('adminEntitlements.activating') }}</option>
        <option value="active">{{ t('common.active') }}</option>
        <option value="queued">{{ t('common.queued') }}</option>
        <option value="expired">{{ t('adminEntitlements.expired') }}</option>
        <option value="cancelled">{{ t('adminEntitlements.cancelled') }}</option>
        <option value="failed">{{ t('adminPayments.failed') }}</option>
      </select>
    </div>

    <AdminSectionState :loading="entitlements.loading.value" :error="entitlements.error.value" @retry="entitlements.load()">
      <div class="admin-list">
        <article v-for="entitlement in filtered" :key="entitlement.id" class="admin-list-row admin-list-row--entitlement">
          <span class="feature-icon feature-icon--small"><PhStack :size="19" /></span>
          <div>
            <strong>{{ entitlement.comboName }}</strong>
            <small>{{ t('adminEntitlements.summary', { user: entitlement.userId, from: formatDate(entitlement.validFrom), to: formatDate(entitlement.validUntil) }) }}</small>
          </div>
          <strong>{{ formatMoney(entitlement.price) }}</strong>
          <StatusBadge :tone="tone(entitlement.status)" :label="entitlement.status" />
          <button
            v-if="canCancel(entitlement)"
            class="button button--ghost-danger button--small"
            type="button"
            @click="selected = entitlement"
          >
            <PhProhibit :size="17" /> {{ t('common.cancel') }}
          </button>
        </article>
        <div v-if="!filtered.length" class="empty-inline">
          <div><h3>{{ t('adminEntitlements.none') }}</h3><p>{{ t('adminEntitlements.noneHint') }}</p></div>
        </div>
      </div>
    </AdminSectionState>

    <AdminReasonDialog
      :open="selected !== null"
      v-model:reason="reason"
      :title="t('adminEntitlements.cancelTitle')"
      :description="t('adminEntitlements.cancelDescription')"
      :confirm-label="t('adminEntitlements.cancelEntitlement')"
      :busy="entitlements.busy.value"
      danger
      @update:open="setCancelOpen"
      @confirm="cancelEntitlement"
    />
  </section>
</template>
