<script setup lang="ts">
import { computed } from 'vue'

import type { AdminUserDetail } from '@/api/adminOperations'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatDateTime, formatMoney } from '@/utils/format'
import { adminUserName } from './adminUserFormat'

const props = defineProps<{ detail: AdminUserDetail }>()
const { t } = useI18n()

const identity = computed(() => props.detail.user.telegramUsername
  ? `@${props.detail.user.telegramUsername}`
  : props.detail.user.username || props.detail.user.telegramId)
const synchronizationTone = computed(() => props.detail.synchronization.status === 'synchronized'
  ? 'success' : props.detail.synchronization.status === 'failed' ? 'danger' : 'warning')
</script>

<template>
  <section class="admin-profile-section admin-profile-overview">
    <div class="admin-profile-identity">
      <UAvatar :text="adminUserName(detail).slice(0, 2).toUpperCase()" size="lg" />
      <div>
        <p class="eyebrow">{{ t('adminUserProfile.overview') }}</p>
        <h2>{{ adminUserName(detail) }}</h2>
        <p>{{ identity }}</p>
      </div>
      <StatusBadge :tone="synchronizationTone" :label="t(`adminUserProfile.sync.${detail.synchronization.status}`)" />
    </div>
    <dl class="admin-profile-facts">
      <div><dt>{{ t('adminUserProfile.balance') }}</dt><dd>{{ formatMoney(detail.balance) }}</dd></div>
      <div><dt>{{ t('adminUserProfile.telegramId') }}</dt><dd>{{ detail.user.telegramId }}</dd></div>
      <div><dt>{{ t('adminUserProfile.joined') }}</dt><dd>{{ formatDateTime(detail.user.createdAt) }}</dd></div>
      <div><dt>{{ t('adminUserProfile.remoteId') }}</dt><dd>{{ detail.synchronization.remoteUserId || t('adminUserProfile.notAvailable') }}</dd></div>
    </dl>
    <div class="admin-profile-signals">
      <div class="admin-profile-signal"><UIcon name="i-ph-arrows-clockwise" /><span><small>{{ t('adminUserProfile.trafficResetAutomation') }}</small><strong>{{ detail.autoTrafficResetEnabled ? t('adminUserProfile.enabled') : t('adminUserProfile.disabled') }}</strong></span></div>
      <div class="admin-profile-signal"><UIcon name="i-ph-clock-countdown" /><span><small>{{ t('adminUserProfile.accountUpdated') }}</small><strong>{{ formatDateTime(detail.user.updatedAt) }}</strong></span></div>
    </div>
    <div class="admin-profile-active">
      <span class="feature-icon feature-icon--small"><UIcon name="i-ph-stack" /></span>
      <div v-if="detail.activeCombo">
        <strong>{{ detail.activeCombo.comboName }}</strong>
        <small>{{ t('adminUserProfile.validity', { from: formatDateTime(detail.activeCombo.validFrom), to: formatDateTime(detail.activeCombo.validUntil) }) }}</small>
      </div>
      <div v-else><strong>{{ t('adminUserProfile.noActiveCombo') }}</strong><small>{{ t('adminUserProfile.noActiveComboHint') }}</small></div>
    </div>
    <UAlert v-if="detail.synchronization.lastError" color="warning" variant="soft" icon="i-ph-warning-circle" :description="detail.synchronization.lastError" />
  </section>
</template>
