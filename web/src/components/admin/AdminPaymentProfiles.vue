<script setup lang="ts">
import { onMounted, reactive, shallowRef } from 'vue'

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

function channelItems(provider: AdminPaymentProfile['provider']): string[] {
  return channelIds[provider]
}

function channelLabel(channel: string): string {
  return t(`adminPaymentProfiles.channelNames.${channel.replace('.', '_')}`)
}

function toggleChannel(profile: AdminPaymentProfile, channel: string, selected: boolean): void {
  const current = drafts[profile.id].enabledChannels
  drafts[profile.id].enabledChannels = selected
    ? [...new Set([...current, channel])]
    : current.filter((value) => value !== channel)
}

async function load(): Promise<void> {
  loading.value = true
  try {
    profiles.value = (await api.getAdminPaymentProfiles()).items
    for (const profile of profiles.value) drafts[profile.id] = { ...profile, enabledChannels: [...profile.enabledChannels] }
  } catch (caught) { error.value = localizedError(caught, 'errors.adminLoad') } finally { loading.value = false }
}

async function save(profile: AdminPaymentProfile): Promise<void> {
  busy.value = profile.id
  error.value = null
  saved.value = false
  try {
    drafts[profile.id] = await api.updateAdminPaymentProfile(profile.provider, drafts[profile.id])
    saved.value = true
  } catch (caught) { error.value = localizedError(caught, 'errors.adminAction') } finally { busy.value = null }
}

onMounted(() => void load())
</script>

<template>
  <section class="payment-profiles">
    <div class="admin-panel__heading"><div><h2>{{ t('adminPaymentProfiles.title') }}</h2><p>{{ t('adminPaymentProfiles.copy') }}</p></div></div>
    <InlineNotice v-if="saved" tone="success">{{ t('adminPaymentProfiles.saved') }}</InlineNotice>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <div v-if="loading" class="payment-profiles__loading"><USkeleton v-for="index in 2" :key="index" class="h-32" /></div>
    <div v-else class="payment-profiles__grid">
      <article v-for="profile in profiles" :key="profile.id" class="payment-profile">
        <div class="payment-profile__header"><div><span class="eyebrow">{{ t(`payment.providers.${profile.provider}`) }}</span><h3>{{ drafts[profile.id].providerName }}</h3></div><SwitchField :id="`payment-profile-${profile.id}`" v-model="drafts[profile.id].enabled" :label="t('common.enabled')" /></div>
        <UFormField :label="t('adminPaymentProfiles.providerName')"><UInput v-model="drafts[profile.id].providerName" /></UFormField>
        <fieldset class="payment-profile__channels">
          <legend>{{ t('adminPaymentProfiles.channels') }}</legend>
          <p>{{ t('adminPaymentProfiles.channelHint') }}</p>
          <UCheckbox v-for="channel in channelItems(profile.provider)" :key="channel" :model-value="drafts[profile.id].enabledChannels.includes(channel)" :label="channelLabel(channel)" @update:model-value="toggleChannel(profile, channel, Boolean($event))" />
        </fieldset>
        <UAlert v-if="!drafts[profile.id].enabledChannels.length" color="warning" variant="soft" :description="t('adminPaymentProfiles.noChannels')" />
        <UFormField :label="t('adminPaymentProfiles.endpoint')"><UInput v-model="drafts[profile.id].endpoint" type="url" /></UFormField>
        <UFormField v-if="profile.provider === 'ezpay'" :label="t('adminPaymentProfiles.merchantId')"><UInput v-model="drafts[profile.id].merchantId" /></UFormField>
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
.payment-profile { display: grid; gap: 0.7rem; padding: 0.9rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.payment-profile__header { display: flex; align-items: start; justify-content: space-between; gap: 0.7rem; }
.payment-profile h3, .payment-profile p { margin: 0; }
.payment-profile h3 { font-size: 1rem; }
.payment-profile__channels { display: grid; gap: 0.45rem; margin: 0; padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); }
.payment-profile__channels legend { color: var(--text-muted); font-size: 0.76rem; font-weight: 700; }
.payment-profile__channels p { color: var(--text-faint); font-size: 0.68rem; line-height: 1.4; }
</style>
