<script setup lang="ts">
import type { AdminUserDetail } from '@/api/adminOperations'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatDateTime, formatMoney } from '@/utils/format'

defineProps<{ detail: AdminUserDetail }>()
const { t } = useI18n()

function abuseTone(action: string): 'warning' | 'danger' | 'neutral' {
  return action === 'warning' ? 'warning' : action === 'ip_ban' ? 'danger' : 'neutral'
}
</script>

<template>
  <section class="admin-profile-section">
    <div class="admin-profile-section__heading"><div><h3>{{ t('adminUserProfile.accountContext') }}</h3><p>{{ t('adminUserProfile.accountContextHint') }}</p></div></div>
    <div class="admin-profile-history">
      <section class="admin-profile-history__group">
        <h4><UIcon name="i-ph-ticket" />{{ t('adminUserProfile.couponWallet') }}</h4>
        <div v-if="detail.couponWallet.length" class="admin-profile-list admin-profile-list--compact">
          <article v-for="grant in detail.couponWallet" :key="grant.id" class="admin-profile-row">
            <div class="admin-profile-row__main"><strong>{{ grant.coupon.name }}</strong><small>{{ grant.coupon.code }} / {{ formatDateTime(grant.createdAt) }}</small></div>
            <StatusBadge tone="success" :label="t('adminUserProfile.activeCoupon')" />
          </article>
        </div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noCoupons') }}</p>
      </section>
      <section class="admin-profile-history__group">
        <h4><UIcon name="i-ph-shield-warning" />{{ t('adminUserProfile.abuseHistory') }}</h4>
        <div v-if="detail.abuseHistory.length" class="admin-profile-list admin-profile-list--compact">
          <article v-for="record in detail.abuseHistory" :key="record.id" class="admin-profile-row">
            <div class="admin-profile-row__main"><strong>{{ record.reason }}</strong><small>{{ formatDateTime(record.occurredAt) }} / {{ record.measuredQPS }} QPS</small></div>
            <StatusBadge :tone="abuseTone(record.action)" :label="t(`adminUserProfile.abuseAction.${record.action}`)" />
          </article>
        </div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noAbuseHistory') }}</p>
      </section>
      <section class="admin-profile-history__group">
        <h4><UIcon name="i-ph-users-three" />{{ t('adminUserProfile.affiliateHistory') }}</h4>
        <div v-if="detail.affiliateHistory.items.length" class="admin-profile-list admin-profile-list--compact">
          <article v-for="(referral, index) in detail.affiliateHistory.items" :key="`${referral.registeredAt}-${index}`" class="admin-profile-row">
            <div class="admin-profile-row__main"><strong>{{ [referral.firstName, referral.lastName].filter(Boolean).join(' ') }}</strong><small>{{ formatDateTime(referral.registeredAt) }}</small></div>
            <strong v-if="referral.commissionAmount">{{ formatMoney(referral.commissionAmount) }}</strong>
            <StatusBadge :tone="referral.status === 'successful' ? 'success' : 'neutral'" :label="t(`adminUserProfile.affiliateStatus.${referral.status}`)" />
          </article>
        </div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noAffiliateHistory') }}</p>
      </section>
    </div>
  </section>
</template>
