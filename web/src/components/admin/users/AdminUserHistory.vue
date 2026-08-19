<script setup lang="ts">
import type { AdminUserDetail, OperationReceipt } from '@/api/adminOperations'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatDateTime, formatMoney } from '@/utils/format'

type Payment = AdminUserDetail['payments'][number]

defineProps<{ detail: AdminUserDetail; busy: boolean }>()
const emit = defineEmits<{
  resolve: [operation: OperationReceipt]
  refundPayment: [payment: Payment]
  creditPayment: [payment: Payment]
}>()
const { t } = useI18n()

function operationTone(status: OperationReceipt['status']): 'neutral' | 'success' | 'warning' | 'danger' {
  if (status === 'succeeded' || status === 'compensated') return 'success'
  if (status === 'failed') return 'danger'
  return status === 'queued' || status === 'processing' ? 'neutral' : 'warning'
}

function canIssueCourtesyCredit(payment: Payment): boolean {
  return payment.status === 'failed' || payment.status === 'expired'
}
</script>

<template>
  <section class="admin-profile-section">
    <div class="admin-profile-section__heading"><div><h3>{{ t('adminUserProfile.records') }}</h3><p>{{ t('adminUserProfile.recordsHint') }}</p></div></div>
    <div class="admin-profile-history">
      <section class="admin-profile-history__group">
        <h4><UIcon name="i-ph-monitor-play" />{{ t('adminUserProfile.emby') }}</h4>
        <div v-if="detail.embyAccounts.length" class="admin-profile-list admin-profile-list--compact">
          <article v-for="account in detail.embyAccounts" :key="account.id" class="admin-profile-row">
            <div class="admin-profile-row__main"><strong>{{ account.username }}</strong><small>{{ t('adminUserProfile.embyFacts', { libraries: account.disabledLibraryIds.length, date: formatDateTime(account.updatedAt) }) }}</small></div>
            <StatusBadge :tone="account.status === 'active' ? 'success' : account.status === 'failed' ? 'danger' : 'warning'" :label="t(`adminUserProfile.embyStatus.${account.status}`)" />
          </article>
        </div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noEmby') }}</p>
      </section>

      <section class="admin-profile-history__group">
        <h4><UIcon name="i-ph-credit-card" />{{ t('adminUserProfile.payments') }}</h4>
        <div v-if="detail.payments.length" class="admin-profile-list admin-profile-list--compact">
          <article v-for="payment in detail.payments" :key="payment.id" class="admin-profile-row">
            <div class="admin-profile-row__main"><strong>{{ payment.provider }} / {{ payment.providerRail }}</strong><small>{{ formatDateTime(payment.createdAt) }} / {{ payment.id }}</small></div>
            <div class="admin-profile-row__meta"><strong>{{ formatMoney(payment.txb) }}</strong><StatusBadge :tone="payment.status === 'paid' ? 'success' : payment.status === 'failed' ? 'danger' : 'neutral'" :label="t(`adminUserProfile.paymentStatus.${payment.status}`)" /></div>
            <div class="row-actions">
              <UButton v-if="payment.status === 'paid'" size="sm" color="warning" variant="outline" icon="i-ph-arrow-counter-clockwise" :disabled="busy" :label="t('adminUserProfile.refundPayment')" @click="emit('refundPayment', payment)" />
              <UButton v-if="canIssueCourtesyCredit(payment)" size="sm" color="primary" variant="outline" icon="i-ph-heart" :disabled="busy" :label="t('adminUserProfile.courtesyCredit')" @click="emit('creditPayment', payment)" />
            </div>
          </article>
        </div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noPayments') }}</p>
      </section>

      <section class="admin-profile-history__group">
        <h4><UIcon name="i-ph-receipt" />{{ t('adminUserProfile.refunds') }}</h4>
        <div v-if="detail.refunds.length" class="admin-profile-list admin-profile-list--compact">
          <article v-for="refund in detail.refunds" :key="refund.id" class="admin-profile-row">
            <div class="admin-profile-row__main"><strong>{{ refund.reason }}</strong><small>{{ formatDateTime(refund.createdAt) }} / {{ refund.paymentOrderId }}</small></div>
            <strong>{{ formatMoney(refund.txb) }}</strong>
          </article>
        </div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noRefunds') }}</p>
      </section>

      <section class="admin-profile-history__group admin-profile-history__group--wide">
        <h4><UIcon name="i-ph-activity" />{{ t('adminUserProfile.openOperations') }}</h4>
        <div v-if="detail.operations.length" class="admin-profile-list admin-profile-list--compact">
          <article v-for="operation in detail.operations" :key="operation.id" class="admin-profile-row">
            <div class="admin-profile-row__main"><strong>{{ operation.kind }}</strong><small>{{ formatDateTime(operation.updatedAt) }} / {{ operation.id }}</small></div>
            <StatusBadge :tone="operationTone(operation.status)" :label="t(`adminUserProfile.operationStatus.${operation.status}`)" />
            <UButton v-if="operation.status === 'pending_review' || operation.status === 'partial'" color="warning" variant="outline" icon="i-ph-gavel" :label="t('adminUserProfile.resolve')" :disabled="busy" @click="emit('resolve', operation)" />
          </article>
        </div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noOpenOperations') }}</p>
      </section>
    </div>
  </section>
</template>
