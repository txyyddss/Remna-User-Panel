<script setup lang="ts">
import { computed, onScopeDispose, reactive, shallowRef } from 'vue'

import { adminOperationsApi, type AdminUserDetail, type OperationReceipt } from '@/api/adminOperations'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useI18n } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { formatDateTime } from '@/utils/format'
import { notifyHaptic } from '@/utils/telegram'

const props = defineProps<{ userId: string; detail: AdminUserDetail }>()
const emit = defineEmits<{ changed: []; queued: [receipt: OperationReceipt] }>()
const { t } = useI18n()
const panel = shallowRef<'ban' | 'unban' | 'relink' | null>(null)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const scan = shallowRef<Awaited<ReturnType<typeof adminOperationsApi.requestUserConnections>> | null>(null)
const form = reactive({ durationMinutes: 1440, reason: '', remnaUserId: '' })
let polling: ReturnType<typeof globalThis.setTimeout> | undefined
const hasBan = computed(() => props.detail.temporaryBan !== null)

function open(next: 'ban' | 'unban' | 'relink'): void {
  error.value = null
  form.reason = ''
  form.remnaUserId = props.detail.synchronization.remoteUserId ?? ''
  form.durationMinutes = 1440
  panel.value = next
}

async function submit(): Promise<void> {
  if (busy.value || form.reason.length < 4 || !panel.value) return
  busy.value = true
  error.value = null
  try {
    const key = createUuid()
    let receipt: OperationReceipt
    if (panel.value === 'ban') receipt = await adminOperationsApi.temporaryBan(props.userId, form.durationMinutes, form.reason, key)
    else if (panel.value === 'unban') receipt = await adminOperationsApi.temporaryUnban(props.userId, form.reason, key)
    else receipt = await adminOperationsApi.relinkRemnaUser(props.userId, form.remnaUserId, form.reason, key)
    panel.value = null
    emit('queued', receipt)
    emit('changed')
    notifyHaptic('success')
  } catch {
    error.value = t('adminUserProfile.actionFailed')
    notifyHaptic('error')
  } finally { busy.value = false }
}

async function startScan(): Promise<void> {
  if (busy.value) return
  busy.value = true
  error.value = null
  try {
    scan.value = await adminOperationsApi.requestUserConnections(props.userId, createUuid())
    pollScan()
    notifyHaptic('success')
  } catch { error.value = t('adminUserProfile.actionFailed'); notifyHaptic('error') } finally { busy.value = false }
}

function pollScan(): void {
  const current = scan.value
  if (!current || current.isCompleted || current.isFailed) return
  if (polling !== undefined) globalThis.clearTimeout(polling)
  polling = globalThis.setTimeout(async () => {
    try { scan.value = await adminOperationsApi.pollUserConnections(props.userId, current.id) } catch { error.value = t('adminUserProfile.actionFailed') }
    pollScan()
  }, 1200)
}

onScopeDispose(() => { if (polling !== undefined) globalThis.clearTimeout(polling) })
</script>

<template>
  <section class="admin-profile-section">
    <div class="admin-profile-section__heading"><div><h3>{{ t('adminUserProfile.connectionControls') }}</h3><p>{{ t('adminUserProfile.connectionControlsHint') }}</p></div></div>
    <div class="row-actions">
      <UButton v-if="!hasBan" color="error" variant="outline" icon="i-ph-prohibit" :disabled="busy" :label="t('adminUserProfile.temporaryBan')" data-haptic="destructive" @click="open('ban')" />
      <UButton v-else color="primary" variant="outline" icon="i-ph-check-circle" :disabled="busy" :label="t('adminUserProfile.temporaryUnban')" data-haptic="confirm" @click="open('unban')" />
      <UButton color="neutral" variant="outline" icon="i-ph-identification-card" :disabled="busy" :label="t('adminUserProfile.changeRemoteId')" data-haptic="open" @click="open('relink')" />
      <UButton color="neutral" variant="outline" icon="i-ph-radar" :disabled="busy" :loading="busy && scan !== null" :label="t('adminUserProfile.scanActiveIp')" data-haptic="refresh" @click="startScan" />
    </div>
    <p v-if="detail.temporaryBan" class="admin-profile-empty">{{ t('adminUserProfile.temporaryBanUntil', { date: formatDateTime(detail.temporaryBan.expiresAt) }) }}</p>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <div v-if="scan" class="admin-profile-list admin-profile-list--compact">
      <UProgress v-if="!scan.isCompleted && !scan.isFailed" :model-value="scan.progressPercent" />
      <article v-for="node in scan.nodes" :key="node.uuid" class="admin-profile-row"><div class="admin-profile-row__main"><strong>{{ node.name }}</strong><small v-for="connection in node.ips" :key="connection.handle">{{ connection.ip }} / {{ formatDateTime(connection.lastSeen) }}</small></div></article>
      <p v-if="scan.isCompleted && !scan.nodes.length" class="admin-profile-empty">{{ t('adminUserProfile.noActiveIp') }}</p>
    </div>
    <UDrawer :open="panel !== null" :title="panel === 'ban' ? t('adminUserProfile.temporaryBan') : panel === 'unban' ? t('adminUserProfile.temporaryUnban') : t('adminUserProfile.changeRemoteId')" :dismissible="!busy" @update:open="!$event && (panel = null)">
      <template #body><div class="form-stack"><UFormField v-if="panel === 'ban'" name="duration" :label="t('adminUserProfile.banDuration')"><UInputNumber v-model="form.durationMinutes" :min="1" :max="525600" :step="1" class="w-full" /></UFormField><UFormField v-if="panel === 'relink'" name="remnaUserId" :label="t('adminUserProfile.remoteId')"><UInput v-model.trim="form.remnaUserId" inputmode="numeric" /></UFormField><UFormField name="reason" :label="t('adminReason.reason')" required><UTextarea v-model.trim="form.reason" :rows="3" :minlength="4" :maxlength="300" /></UFormField></div></template>
      <template #footer><UButton color="neutral" variant="outline" :label="t('common.cancel')" :disabled="busy" @click="panel = null" /><UButton :color="panel === 'ban' ? 'error' : 'primary'" :label="t('common.confirm')" :loading="busy" :disabled="busy || form.reason.length < 4 || (panel === 'relink' && !form.remnaUserId)" data-haptic="confirm" @click="submit" /></template>
    </UDrawer>
  </section>
</template>
