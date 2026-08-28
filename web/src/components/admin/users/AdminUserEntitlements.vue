<script setup lang="ts">
import type { AdminEntitlement } from '@/api/adminOperations'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatBytes, formatDateTime, formatMoney } from '@/utils/format'
import { entitlementTone } from './adminUserFormat'

defineProps<{ items: AdminEntitlement[]; busy: boolean }>()
const emit = defineEmits<{ edit: [item: AdminEntitlement]; refund: [item: AdminEntitlement] }>()
const { t } = useI18n()

function refundable(item: AdminEntitlement): boolean {
  return item.status === 'activating' || item.status === 'active' || item.status === 'queued'
}
</script>

<template>
  <section class="admin-profile-section">
    <div class="admin-profile-section__heading">
      <div><h3>{{ t('adminUserProfile.entitlements') }}</h3><p>{{ t('adminUserProfile.entitlementsHint') }}</p></div>
    </div>
    <div v-if="items.length" class="admin-profile-list">
      <article v-for="item in items" :key="item.id" class="admin-profile-row">
        <div class="admin-profile-row__main">
          <strong>{{ item.comboName }}</strong>
          <small>{{ t('adminUserProfile.validity', { from: formatDateTime(item.validFrom), to: formatDateTime(item.validUntil) }) }}</small>
          <small>{{ t('adminUserProfile.entitlementFacts', { traffic: formatBytes(item.trafficLimitBytes), cadence: t(`adminUserProfile.cadence.${item.resetStrategy}`), squads: item.squadUuids.length }) }}</small>
        </div>
        <div class="admin-profile-row__meta">
          <strong>{{ formatMoney(item.price) }}</strong>
          <StatusBadge :tone="entitlementTone(item.status)" :label="t(`adminUserProfile.entitlementStatus.${item.status}`)" />
        </div>
        <div class="row-actions admin-profile-row__actions">
          <UButton color="neutral" variant="ghost" icon="i-ph-pencil-simple" :label="t('adminUserProfile.edit')" :disabled="busy" @click="emit('edit', item)" />
          <UButton v-if="refundable(item)" color="warning" variant="ghost" icon="i-ph-arrow-u-up-left" :label="t('adminUserProfile.refund')" :disabled="busy" @click="emit('refund', item)" />
        </div>
      </article>
    </div>
    <div v-else class="empty-inline"><div><h3>{{ t('adminUserProfile.noEntitlements') }}</h3><p>{{ t('adminUserProfile.noEntitlementsHint') }}</p></div></div>
  </section>
</template>
