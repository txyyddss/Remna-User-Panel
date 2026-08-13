<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from 'vue'

import { api, type AdminPaymentProfile } from '@/api/client'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import { localizedError, useI18n } from '@/i18n'

const { t } = useI18n()
const profiles = shallowRef<AdminPaymentProfile[]>([])
const loading = shallowRef(true)
const busy = shallowRef<string | null>(null)
const error = shallowRef<string | null>(null)
const saved = shallowRef(false)
const drafts = reactive<Record<string, AdminPaymentProfile>>({})
const channelIds: Record<AdminPaymentProfile['provider'], string[]> = {
  ezpay: ['alipay', 'wxpay', 'qqpay', 'bank', 'jdpay'],
  bepusdt: ['usdt.trc20', 'usdt.erc20', 'usdt.polygon', 'usdt.bep20', 'usdt.aptos', 'usdt.solana', 'usdt.xlayer', 'usdt.arbitrum', 'usdt.plasma', 'usdt.ton'],
}
const providerItems = computed(() => (['ezpay', 'bepusdt'] as const).map((provider) => ({
  label: t(`payment.providers.${provider}`),
  value: provider,
})))

function channelItems(provider: AdminPaymentProfile['provider']): { label: string; value: string }[] {
  return channelIds[provider].map((channel) => ({
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
  saved.value = false
  try {
    const response = profile.id.startsWith('draft-')
      ? await api.createAdminPaymentProfile(draft)
      : await api.updateAdminPaymentProfile(profile.id, draft)
    delete drafts[profile.id]
    drafts[response.id] = { ...response, enabledChannels: [...response.enabledChannels] }
    profiles.value = profiles.value.map((item) => item.id === profile.id ? response : item)
    saved.value = true
  } catch (caught) { error.value = localizedError(caught, 'errors.adminAction') } finally { busy.value = null }
}

onMounted(() => void load())
</script>

<template>
  <section class="payment-profiles">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminPaymentProfiles.title') }}</h2><p>{{ t('adminPaymentProfiles.copy') }}</p></div>
      <UButton icon="i-ph-plus" :label="t('adminPaymentProfiles.add')" data-haptic @click="addProfile" />
    </div>
    <InlineNotice v-if="saved" tone="success">{{ t('adminPaymentProfiles.saved') }}</InlineNotice>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <div v-if="loading" class="payment-profiles__loading"><USkeleton v-for="index in 2" :key="index" class="h-32" /></div>
    <div v-else class="payment-profiles__grid">
      <article v-for="profile in profiles" :key="profile.id" class="payment-profile">
        <div class="payment-profile__header">
          <div>
            <span class="eyebrow">{{ t(`payment.providers.${drafts[profile.id].provider}`) }}</span>
            <h3>{{ drafts[profile.id].providerName || t('adminPaymentProfiles.unnamed') }}</h3>
          </div>
          <SwitchField :id="`payment-profile-${profile.id}`" v-model="drafts[profile.id].enabled" :label="t('common.enabled')" />
        </div>
        <UFormField v-if="profile.id.startsWith('draft-')" :label="t('adminPaymentProfiles.provider')">
          <USelect v-model="drafts[profile.id].provider" :items="providerItems" @update:model-value="providerChanged(profile)" />
        </UFormField>
        <UFormField :label="t('adminPaymentProfiles.providerName')"><UInput v-model="drafts[profile.id].providerName" /></UFormField>
        <UFormField :label="t('adminPaymentProfiles.channels')" :description="t('adminPaymentProfiles.channelHint')">
          <UCheckboxGroup
            v-model="drafts[profile.id].enabledChannels"
            :items="channelItems(drafts[profile.id].provider)"
            orientation="vertical"
            variant="card"
          />
        </UFormField>
        <UAlert v-if="!drafts[profile.id].enabledChannels.length" color="warning" variant="soft" :description="t('adminPaymentProfiles.noChannels')" />
        <UFormField :label="t('adminPaymentProfiles.endpoint')"><UInput v-model="drafts[profile.id].endpoint" type="url" /></UFormField>
        <UFormField v-if="drafts[profile.id].provider === 'ezpay'" :label="t('adminPaymentProfiles.merchantId')"><UInput v-model="drafts[profile.id].merchantId" /></UFormField>
        <UFormField :label="t('adminPaymentProfiles.credential')"><UInput v-model="drafts[profile.id].credential" type="password" :placeholder="profile.configured ? t('adminPaymentProfiles.keepCredential') : ''" autocomplete="new-password" /></UFormField>
        <UFormField :label="t('adminPaymentProfiles.acknowledgement')"><UInput v-model="drafts[profile.id].acknowledgement" /></UFormField>
        <UButton block :loading="busy === profile.id" :label="t('common.save')" data-haptic @click="save(profile)" />
      </article>
    </div>
  </section>
</template>

<style scoped>
.payment-profiles, .payment-profiles__grid { display: grid; gap: 0.8rem; }
.payment-profiles__grid { grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr)); }
.payment-profile { display: grid; gap: 0.7rem; padding: 0.9rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.payment-profile__header { display: flex; align-items: start; justify-content: space-between; gap: 0.7rem; }
.payment-profile h3, .payment-profile p { margin: 0; }
.payment-profile h3 { font-size: 1rem; }
</style>
