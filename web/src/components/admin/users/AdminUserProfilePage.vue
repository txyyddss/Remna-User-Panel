<script setup lang="ts">
import { computed, shallowRef, toRef } from 'vue'

import { adminOperationsApi, type AdminEntitlement, type EntitlementEditRequest, type ComboReplacementRequest, type OperationReceipt, type OperationResolution } from '@/api/adminOperations'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useI18n } from '@/i18n'
import AdminReasonDialog from '../AdminReasonDialog.vue'
import AdminComboReplacementDialog from './AdminComboReplacementDialog.vue'
import AdminEntitlementEditor from './AdminEntitlementEditor.vue'
import AdminOperationResolutionDialog from './AdminOperationResolutionDialog.vue'
import AdminUserEntitlements from './AdminUserEntitlements.vue'
import AdminUserHistory from './AdminUserHistory.vue'
import AdminUserOverview from './AdminUserOverview.vue'
import { useAdminUserProfile } from './useAdminUserProfile'

const props = defineProps<{ userId: string }>()
const emit = defineEmits<{ back: [] }>()
const { t } = useI18n()
const profile = useAdminUserProfile(toRef(props, 'userId'))
const editing = shallowRef<AdminEntitlement | null>(null)
const refunding = shallowRef<AdminEntitlement | null>(null)
const refundReason = shallowRef('')
const replacementOpen = shallowRef(false)
const resolving = shallowRef<OperationReceipt | null>(null)
const busy = computed(() => profile.busyAction.value !== null)
const dialogError = computed(() => profile.optionsError.value ?? profile.error.value)

async function openEditor(item: AdminEntitlement): Promise<void> {
  profile.clearActionError()
  editing.value = item
  await profile.loadOptions()
}

function openRefund(item: AdminEntitlement): void {
  profile.clearActionError()
  refundReason.value = ''
  refunding.value = item
}

async function openReplacement(): Promise<void> {
  profile.clearActionError()
  replacementOpen.value = true
  await profile.loadOptions()
}

async function saveEntitlement(body: EntitlementEditRequest): Promise<void> {
  const item = editing.value
  if (!item) return
  const ok = await profile.perform(`edit:${item.id}`, (key) =>
    adminOperationsApi.editEntitlement(props.userId, item.id, body, key))
  if (ok) editing.value = null
}

async function refundEntitlement(): Promise<void> {
  const item = refunding.value
  if (!item) return
  const ok = await profile.perform(`refund:${item.id}`, (key) =>
    adminOperationsApi.refundEntitlement(props.userId, item.id, refundReason.value, key))
  if (ok) refunding.value = null
}

async function replaceCombo(body: ComboReplacementRequest): Promise<void> {
  const ok = await profile.perform(`replace:${props.userId}`, (key) =>
    adminOperationsApi.replaceCombo(props.userId, body, key))
  if (ok) replacementOpen.value = false
}

async function resolveOperation(payload: { resolution: OperationResolution; reason: string }): Promise<void> {
  const operation = resolving.value
  if (!operation) return
  const ok = await profile.perform(`resolve:${operation.id}`, (key) =>
    adminOperationsApi.resolveOperation(operation.id, payload.resolution, payload.reason, key))
  if (ok) resolving.value = null
}

function refresh(): void {
  profile.clearActionError()
  void profile.load()
}
</script>

<template>
  <section class="admin-panel admin-profile">
    <div class="admin-profile-toolbar">
      <UButton color="neutral" variant="ghost" icon="i-ph-arrow-left" :label="t('adminUserProfile.backToUsers')" @click="emit('back')" />
      <UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="t('common.refresh')" :loading="profile.loading.value" @click="refresh" />
    </div>
    <InlineNotice v-if="profile.conflict.value" tone="warning" :title="t('adminUserProfile.conflictTitle')">{{ t('adminUserProfile.conflict') }}</InlineNotice>
    <InlineNotice v-else-if="profile.error.value" tone="warning">{{ profile.error.value }}</InlineNotice>
    <div v-if="profile.loading.value" class="admin-loading" :aria-label="t('adminUserProfile.loading')">
      <USkeleton class="h-24" /><USkeleton class="h-48" /><USkeleton class="h-36" />
    </div>
    <template v-else-if="profile.detail.value">
      <AdminUserOverview :detail="profile.detail.value" />
      <AdminUserEntitlements :items="profile.detail.value.entitlements" :busy="busy" :can-replace="Boolean(profile.detail.value.activeCombo)" @edit="openEditor" @refund="openRefund" @replace="openReplacement" />
      <AdminUserHistory :detail="profile.detail.value" :busy="busy" @resolve="resolving = $event" />
    </template>
    <div v-else class="empty-inline"><div><h3>{{ t('adminUserProfile.unavailable') }}</h3><p>{{ t('adminUserProfile.unavailableHint') }}</p></div></div>

    <AdminEntitlementEditor :open="editing !== null" :item="editing" :combos="profile.options.value.combos" :squads="profile.options.value.squads" :busy="busy" :options-loading="profile.optionsLoading.value" :error="dialogError" @update:open="!$event && (editing = null)" @save="saveEntitlement" />
    <AdminReasonDialog :open="refunding !== null" v-model:reason="refundReason" :title="t('adminUserProfile.refundTitle')" :description="t('adminUserProfile.refundHint')" :confirm-label="t('adminUserProfile.issueRefund')" :busy="busy" :error="profile.error.value" danger @update:open="!$event && (refunding = null)" @confirm="refundEntitlement" />
    <AdminComboReplacementDialog :open="replacementOpen" :current="profile.detail.value?.activeCombo ?? null" :combos="profile.options.value.combos" :squads="profile.options.value.squads" :busy="busy" :options-loading="profile.optionsLoading.value" :error="dialogError" @update:open="replacementOpen = $event" @replace="replaceCombo" />
    <AdminOperationResolutionDialog :open="resolving !== null" :operation="resolving" :busy="busy" :error="profile.error.value" @update:open="!$event && (resolving = null)" @resolve="resolveOperation" />
  </section>
</template>
