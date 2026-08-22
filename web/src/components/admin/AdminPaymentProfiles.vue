<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from 'vue'

import { api, type AdminPaymentProfile } from '@/api/client'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import { localizedError, useI18n } from '@/i18n'

const { t } = useI18n()
const profiles = shallowRef<AdminPaymentProfile[]>([])
const loading = shallowRef(true)
const busy = shallowRef<string | null>(null)
const error = shallowRef<string | null>(null)
const success = shallowRef<string | null>(null)
const deleting = shallowRef<AdminPaymentProfile | null>(null)
const drafts = reactive<Record<string, AdminPaymentProfile>>({})
const channelIds = ['alipay', 'wxpay', 'qqpay', 'bank', 'jdpay']
const providerItems = computed(() => (['ezpay', 'bepusdt'] as const).map((provider) => ({
  label: t(`payment.providers.${provider}`),
  value: provider,
})))

function channelItems(): { label: string; value: string }[] {
  return channelIds.map((channel) => ({
    label: t(`adminPaymentProfiles.channelNames.${channel.replace('.', '_')}`),
    value: channel,
  }))
}

function emptyProfile(): AdminPaymentProfile {
  return {
    id: `draft-${Date.now()}`,
    provider: 'ezpay',
    providerName: '',
    enabledChannels: [],
    endpoint: '',
    merchantId: '',
    credential: '',
    acknowledgement: 'ok',
    enabled: false,
    configured: false,
  }
}

function addProfile(): void {
  const draft = emptyProfile()
  drafts[draft.id] = draft
  profiles.value = [...profiles.value, draft]
}

function providerChanged(profile: AdminPaymentProfile): void {
  drafts[profile.id].enabledChannels = []
  drafts[profile.id].merchantId = ''
}

async function load(): Promise<void> {
  loading.value = true
  try {
    profiles.value = (await api.getAdminPaymentProfiles()).items
    for (const profile of profiles.value) drafts[profile.id] = { ...profile, enabledChannels: [...profile.enabledChannels] }
  } catch (caught) { error.value = localizedError(caught, 'errors.adminLoad') } finally { loading.value = false }
}

async function save(profile: AdminPaymentProfile): Promise<void> {
  const draft = drafts[profile.id]
  if (!draft) return
  busy.value = profile.id
  error.value = null
  success.value = null
  try {
    const response = profile.id.startsWith('draft-')
      ? await api.createAdminPaymentProfile(draft)
      : await api.updateAdminPaymentProfile(profile.id, draft)
    delete drafts[profile.id]
    drafts[response.id] = { ...response, enabledChannels: [...response.enabledChannels] }
    profiles.value = profiles.value.map((item) => item.id === profile.id ? response : item)
    success.value = t('adminPaymentProfiles.saved')
  } catch (caught) { error.value = localizedError(caught, 'errors.adminAction') } finally { busy.value = null }
}

async function removeProfile(): Promise<void> {
  const profile = deleting.value
  if (!profile) return
  busy.value = profile.id
  error.value = null
  success.value = null
  try {
    if (!profile.id.startsWith('draft-')) await api.deleteAdminPaymentProfile(profile.id)
    profiles.value = profiles.value.filter((item) => item.id !== profile.id)
    delete drafts[profile.id]
    deleting.value = null
    success.value = t('adminPaymentProfiles.deleted')
  } catch (caught) { error.value = localizedError(caught, 'errors.adminAction') } finally { busy.value = null }
}

async function saveAll(): Promise<void> {
  for (const profile of profiles.value) await save(profile)
}

defineExpose({ saveAll, loading })

onMounted(() => void load())
</script>

<template>
  <section class="payment-profiles">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminPaymentProfiles.title') }}</h2><p>{{ t('adminPaymentProfiles.copy') }}</p></div>
      <UButton icon="i-ph-plus" :label="t('adminPaymentProfiles.add')" data-haptic="action" @click="addProfile" />
    </div>
    <InlineNotice v-if="success" tone="success">{{ success }}</InlineNotice>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <div v-if="loading" class="payment-profiles__loading"><USkeleton v-for="index in 2" :key="index" class="h-32" /></div>
    <div v-else class="payment-profiles__grid">
      <article v-for="profile in profiles" :key="profile.id" class="payment-profile">
        <div class="payment-profile__header">
          <div>
            <span class="eyebrow">{{ t(`payment.providers.${drafts[profile.id].provider}`) }}</span>
            <h3>{{ drafts[profile.id].providerName || t('adminPaymentProfiles.unnamed') }}</h3>
          </div>
          <div class="payment-profile__actions">
            <SwitchField :id="`payment-profile-${profile.id}`" v-model="drafts[profile.id].enabled" :label="t('common.enabled')" />
            <UTooltip :text="t('adminPaymentProfiles.delete')">
              <UButton color="error" variant="ghost" icon="i-ph-trash" :aria-label="t('adminPaymentProfiles.delete')" :disabled="busy === profile.id" data-haptic="destructive" @click="deleting = profile" />
            </UTooltip>
          </div>
        </div>
        <UFormField v-if="profile.id.startsWith('draft-')" :label="t('adminPaymentProfiles.provider')">
          <USelect v-model="drafts[profile.id].provider" :items="providerItems" @update:model-value="providerChanged(profile)" />
        </UFormField>
        <UFormField :label="t('adminPaymentProfiles.providerName')"><UInput v-model="drafts[profile.id].providerName" /></UFormField>
        <UFormField v-if="drafts[profile.id].provider === 'ezpay'" :label="t('adminPaymentProfiles.channels')" :description="t('adminPaymentProfiles.channelHint')">
          <UCheckboxGroup
            v-model="drafts[profile.id].enabledChannels"
            :items="channelItems()"
            orientation="vertical"
            variant="card"
          />
        </UFormField>
        <UAlert v-if="drafts[profile.id].provider === 'ezpay' && !drafts[profile.id].enabledChannels.length" color="warning" variant="soft" :description="t('adminPaymentProfiles.noChannels')" />
        <UAlert v-else-if="drafts[profile.id].provider === 'bepusdt'" color="neutral" variant="soft" icon="i-ph-arrows-clockwise" :description="t('adminPaymentProfiles.discoveryHint')" />
        <UFormField :label="t('adminPaymentProfiles.endpoint')"><UInput v-model="drafts[profile.id].endpoint" type="url" /></UFormField>
        <UFormField v-if="drafts[profile.id].provider === 'ezpay'" :label="t('adminPaymentProfiles.merchantId')"><UInput v-model="drafts[profile.id].merchantId" /></UFormField>
        <UFormField :label="t('adminPaymentProfiles.credential')"><UInput v-model="drafts[profile.id].credential" type="password" :placeholder="profile.configured ? t('adminPaymentProfiles.keepCredential') : ''" autocomplete="new-password" /></UFormField>
        <UFormField :label="t('adminPaymentProfiles.acknowledgement')"><UInput v-model="drafts[profile.id].acknowledgement" /></UFormField>
      </article>
    </div>
    <ConfirmDialog
      :open="Boolean(deleting)"
      :title="t('adminPaymentProfiles.deleteTitle', { name: deleting?.providerName || t('adminPaymentProfiles.unnamed') })"
      :description="t('adminPaymentProfiles.deleteDescription')"
      :confirm-label="t('adminPaymentProfiles.deleteConfirm')"
      :busy="Boolean(deleting && busy === deleting.id)"
      danger
      @update:open="!$event && (deleting = null)"
      @confirm="removeProfile"
    />
  </section>
</template>

<style scoped>
.payment-profiles, .payment-profiles__grid { display: grid; gap: 0.8rem; }
.payment-profiles__grid { grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr)); }
.payment-profile { display: grid; gap: 0.7rem; padding: 0.9rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.payment-profile__header { display: flex; align-items: start; justify-content: space-between; gap: 0.7rem; }
.payment-profile__actions { display: flex; align-items: center; gap: 0.35rem; }
.payment-profile h3, .payment-profile p { margin: 0; }
.payment-profile h3 { font-size: 1rem; }
</style>
