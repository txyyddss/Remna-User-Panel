<script setup lang="ts">
import { shallowRef } from 'vue'

import { adminOperationsApi, type AdminUserDetail } from '@/api/adminOperations'
import InlineNotice from '@/components/common/InlineNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { formatDateTime, formatMoney } from '@/utils/format'
import { notifyHaptic } from '@/utils/telegram'

defineProps<{ userId: string; detail: AdminUserDetail }>()
const emit = defineEmits<{ changed: [] }>()
const { t } = useI18n()
const deleting = shallowRef<AdminUserDetail['abuseHistory'][number] | null>(null)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)

function abuseTone(action: string): 'warning' | 'danger' | 'neutral' {
  return action === 'warning' ? 'warning' : action === 'ip_ban' ? 'danger' : 'neutral'
}

async function deleteAbuse(): Promise<void> {
  if (!deleting.value || busy.value) return
  busy.value = true
  try {
    await adminOperationsApi.deleteAbuseRecord(deleting.value.id, createUuid())
    deleting.value = null; emit('changed'); notifyHaptic('success')
  } catch { error.value = t('adminUserProfile.actionFailed'); notifyHaptic('error') } finally { busy.value = false }
}
</script>

<template>
  <section class="admin-profile-section">
    <div class="admin-profile-section__heading"><div><h3>{{ t('adminUserProfile.accountContext') }}</h3><p>{{ t('adminUserProfile.accountContextHint') }}</p></div></div>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <div class="admin-profile-history">
      <section class="admin-profile-history__group">
        <h4><UIcon name="i-ph-shield-warning" />{{ t('adminUserProfile.abuseHistory') }}</h4>
        <div v-if="detail.abuseHistory.length" class="admin-profile-list admin-profile-list--compact"><article v-for="record in detail.abuseHistory" :key="record.id" class="admin-profile-row"><div class="admin-profile-row__main"><strong>{{ record.reason }}</strong><small>{{ formatDateTime(record.occurredAt) }} / {{ record.measuredQPS }} QPS</small></div><StatusBadge :tone="abuseTone(record.action)" :label="t(`adminUserProfile.abuseAction.${record.action}`)" /><div class="admin-profile-row__actions"><UButton color="error" variant="ghost" square icon="i-ph-trash" :aria-label="t('adminUserProfile.deleteAbuse')" data-haptic="destructive" @click="deleting = record" /></div></article></div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noAbuseHistory') }}</p>
      </section>
      <section class="admin-profile-history__group">
        <h4><UIcon name="i-ph-users-three" />{{ t('adminUserProfile.affiliateHistory') }}</h4>
        <div v-if="detail.affiliateHistory.items.length" class="admin-profile-list admin-profile-list--compact"><article v-for="(referral, index) in detail.affiliateHistory.items" :key="`${referral.registeredAt}-${index}`" class="admin-profile-row"><div class="admin-profile-row__main"><strong>{{ [referral.firstName, referral.lastName].filter(Boolean).join(' ') }}</strong><small>{{ formatDateTime(referral.registeredAt) }}</small></div><div class="admin-profile-row__meta"><strong v-if="referral.commissionAmount">{{ formatMoney(referral.commissionAmount) }}</strong><StatusBadge :tone="referral.status === 'successful' ? 'success' : 'warning'" :label="t(`adminUserProfile.affiliateStatus.${referral.status}`)" /></div></article></div>
        <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noAffiliateHistory') }}</p>
      </section>
    </div>
    <UModal :open="deleting !== null" :title="t('adminUserProfile.deleteAbuse')" :description="t('adminUserProfile.deleteAbuseHint')" :dismissible="!busy" :close="false" :ui="{ header: 'tg-overlay-header--centered', wrapper: 'tg-overlay-copy--centered', footer: 'justify-end' }" @update:open="!$event && (deleting = null)"><template #footer="{ close }"><UButton color="neutral" variant="outline" :label="t('common.cancel')" @click="close" /><UButton color="error" :label="t('common.delete')" :loading="busy" data-haptic="destructive" @click="deleteAbuse" /></template></UModal>
  </section>
</template>
