<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { PhArrowCounterClockwise, PhReceipt } from '@phosphor-icons/vue'

import type { AdminPaymentOrder, Refund } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { useI18n } from '@/i18n'
import { formatDateTime, formatMoney } from '@/utils/format'
import AdminReasonDialog from './AdminReasonDialog.vue'
import AdminSectionState from './AdminSectionState.vue'

const payments = useAdminSection<AdminPaymentOrder>('payments')
const refunds = useAdminSection<Refund>('refunds')
const selected = shallowRef<AdminPaymentOrder | null>(null)
const reason = shallowRef('')
const statusFilter = shallowRef<'all' | AdminPaymentOrder['status']>('all')
const { t } = useI18n()

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
      <div><h2>{{ t('adminPayments.title') }}</h2><p>{{ t('adminPayments.copy') }}</p></div>
      <select v-model="statusFilter" class="compact-select" :aria-label="t('adminPayments.filter')">
        <option value="all">{{ t('adminPayments.all') }}</option>
        <option value="creating">{{ t('adminPayments.creating') }}</option>
        <option value="pending">{{ t('adminPayments.pending') }}</option>
        <option value="paid">{{ t('adminPayments.paid') }}</option>
        <option value="refunded">{{ t('adminPayments.refunded') }}</option>
        <option value="failed">{{ t('adminPayments.failed') }}</option>
        <option value="expired">{{ t('adminPayments.expired') }}</option>
      </select>
    </div>
    <AdminSectionState :loading="payments.loading.value" :error="payments.error.value" @retry="payments.load()">
      <div class="admin-list">
        <article v-for="payment in filtered" :key="payment.id" class="admin-list-row admin-list-row--payment">
          <span class="feature-icon feature-icon--small"><PhReceipt :size="19" /></span>
          <div><strong>{{ payment.provider }} {{ payment.id }}</strong><small>{{ t('adminPayments.paymentSummary', { user: payment.userId, date: formatDateTime(payment.createdAt), amount: formatMoney(payment.txb) }) }}</small></div>
          <strong>{{ payment.payableAmount }} {{ payment.payableCurrency }}</strong>
          <StatusBadge :tone="tone(payment.status)" :label="payment.status" />
          <button v-if="payment.status === 'paid'" class="button button--secondary button--small" type="button" @click="selected = payment"><PhArrowCounterClockwise :size="17" /> {{ t('adminPayments.refund') }}</button>
        </article>
        <div v-if="!filtered.length" class="empty-inline"><div><h3>{{ t('adminPayments.none') }}</h3><p>{{ t('adminPayments.noneHint') }}</p></div></div>
      </div>
    </AdminSectionState>

    <div class="admin-subsection-heading">
      <div><h3>{{ t('adminPayments.refundRecords') }}</h3><p>{{ t('adminPayments.refundRecordsHint') }}</p></div>
    </div>
    <AdminSectionState :loading="refunds.loading.value" :error="refunds.error.value" @retry="refunds.load()">
      <div class="admin-list admin-list--compact">
        <article v-for="refundRecord in refunds.items.value" :key="refundRecord.id" class="admin-list-row admin-list-row--refund">
          <span class="feature-icon feature-icon--small"><PhArrowCounterClockwise :size="19" /></span>
          <div>
            <strong>{{ refundRecord.reason }}</strong>
            <small>{{ t('adminPayments.refundSummary', { payment: refundRecord.paymentOrderId, date: formatDateTime(refundRecord.createdAt), actor: refundRecord.actorUserId ?? t('adminAudit.system') }) }}</small>
          </div>
          <strong>{{ formatMoney(refundRecord.txb) }}</strong>
          <StatusBadge tone="success" :label="t('adminPayments.completed')" />
        </article>
        <div v-if="!refunds.items.value.length" class="empty-inline">
          <div><h3>{{ t('adminPayments.noRefunds') }}</h3><p>{{ t('adminPayments.noRefundsHint') }}</p></div>
        </div>
      </div>
    </AdminSectionState>
    <AdminReasonDialog
      :open="selected !== null"
      v-model:reason="reason"
      :title="t('adminPayments.refundTitle')"
      :description="t('adminPayments.refundDescription')"
      :confirm-label="t('adminPayments.issueRefund')"
      :busy="payments.busy.value"
      danger
      @update:open="setRefundOpen"
      @confirm="refund"
    />
  </section>
</template>
