<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { PhArrowCounterClockwise, PhReceipt } from '@phosphor-icons/vue'

import type { AdminPaymentOrder, Refund } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { formatDateTime, formatMoney } from '@/utils/format'
import AdminReasonDialog from './AdminReasonDialog.vue'
import AdminSectionState from './AdminSectionState.vue'

const payments = useAdminSection<AdminPaymentOrder>('payments')
const refunds = useAdminSection<Refund>('refunds')
const selected = shallowRef<AdminPaymentOrder | null>(null)
const reason = shallowRef('')
const statusFilter = shallowRef<'all' | AdminPaymentOrder['status']>('all')

const filtered = computed(() => statusFilter.value === 'all'
  ? payments.items.value
  : payments.items.value.filter((item) => item.status === statusFilter.value))

function tone(status: AdminPaymentOrder['status']): 'neutral' | 'success' | 'warning' | 'danger' {
  if (status === 'paid') return 'success'
  if (status === 'pending') return 'warning'
  if (status === 'failed') return 'danger'
  return 'neutral'
}

async function refund(): Promise<void> {
  if (!selected.value) return
  const { api } = await import('@/api/client')
  const ok = await payments.perform(() => api.refundPayment(selected.value!.id, reason.value))
  if (ok) {
    selected.value = null
    reason.value = ''
    await refunds.load()
  }
}

function setRefundOpen(open: boolean): void {
  if (open) return
  selected.value = null
  reason.value = ''
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>Payments</h2><p>Provider events, credits, and immutable reversals.</p></div>
      <select v-model="statusFilter" class="compact-select" aria-label="Filter payments">
        <option value="all">All statuses</option>
        <option value="creating">Creating</option>
        <option value="pending">Pending</option>
        <option value="paid">Paid</option>
        <option value="refunded">Refunded</option>
        <option value="failed">Failed</option>
        <option value="expired">Expired</option>
      </select>
    </div>
    <AdminSectionState :loading="payments.loading.value" :error="payments.error.value" @retry="payments.load()">
      <div class="admin-list">
        <article v-for="payment in filtered" :key="payment.id" class="admin-list-row admin-list-row--payment">
          <span class="feature-icon feature-icon--small"><PhReceipt :size="19" /></span>
          <div><strong>{{ payment.provider }} {{ payment.id }}</strong><small>User {{ payment.userId }} / {{ formatDateTime(payment.createdAt) }} / {{ formatMoney(payment.txb) }} credit</small></div>
          <strong>{{ payment.payableAmount }} {{ payment.payableCurrency }}</strong>
          <StatusBadge :tone="tone(payment.status)" :label="payment.status" />
          <button v-if="payment.status === 'paid'" class="button button--secondary button--small" type="button" @click="selected = payment"><PhArrowCounterClockwise :size="17" /> Refund</button>
        </article>
        <div v-if="!filtered.length" class="empty-inline"><div><h3>No payments in this view</h3><p>Provider orders will appear after users start a top-up.</p></div></div>
      </div>
    </AdminSectionState>

    <div class="admin-subsection-heading">
      <div><h3>Refund records</h3><p>Completed reversals are immutable and retain the operator reason.</p></div>
    </div>
    <AdminSectionState :loading="refunds.loading.value" :error="refunds.error.value" @retry="refunds.load()">
      <div class="admin-list admin-list--compact">
        <article v-for="refundRecord in refunds.items.value" :key="refundRecord.id" class="admin-list-row admin-list-row--refund">
          <span class="feature-icon feature-icon--small"><PhArrowCounterClockwise :size="19" /></span>
          <div>
            <strong>{{ refundRecord.reason }}</strong>
            <small>Payment {{ refundRecord.paymentOrderId }} / {{ formatDateTime(refundRecord.createdAt) }} / Actor {{ refundRecord.actorUserId ?? 'System' }}</small>
          </div>
          <strong>{{ formatMoney(refundRecord.txb) }}</strong>
          <StatusBadge tone="success" label="completed" />
        </article>
        <div v-if="!refunds.items.value.length" class="empty-inline">
          <div><h3>No refunds</h3><p>Immutable reversal records will remain visible here.</p></div>
        </div>
      </div>
    </AdminSectionState>
    <AdminReasonDialog
      :open="selected !== null"
      v-model:reason="reason"
      title="Refund this payment?"
      description="The TXB credit will be reversed. Entitlements may be cancelled if the account becomes negative."
      confirm-label="Issue refund"
      :busy="payments.busy.value"
      danger
      @update:open="setRefundOpen"
      @confirm="refund"
    />
  </section>
</template>
