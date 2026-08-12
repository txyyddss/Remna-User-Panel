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

async function load(): Promise<void> {
  loading.value = true
  try {
    profiles.value = (await api.getAdminPaymentProfiles()).items
    for (const profile of profiles.value) drafts[profile.id] = { ...profile }
  } catch (caught) { error.value = localizedError(caught, 'errors.adminLoad') } finally { loading.value = false }
}

async function save(profile: AdminPaymentProfile): Promise<void> {
  busy.value = profile.id
  error.value = null
  saved.value = false
  try {
    drafts[profile.id] = await api.updateAdminPaymentProfile(profile.provider, profile.rail, drafts[profile.id])
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
        <div class="payment-profile__header"><div><span class="eyebrow">{{ profile.provider }}</span><h3>{{ profile.rail }}</h3></div><SwitchField :id="`payment-profile-${profile.id}`" v-model="drafts[profile.id].enabled" :label="t('common.enabled')" /></div>
        <UFormField :label="t('adminPaymentProfiles.channelName')"><UInput v-model="drafts[profile.id].channelName" /></UFormField>
        <UFormField :label="t('adminPaymentProfiles.endpoint')"><UInput v-model="drafts[profile.id].endpoint" type="url" /></UFormField>
        <UFormField v-if="profile.provider === 'ezpay'" :label="t('adminPaymentProfiles.merchantId')"><UInput v-model="drafts[profile.id].merchantId" /></UFormField>
        <UFormField :label="t('adminPaymentProfiles.credential')"><UInput v-model="drafts[profile.id].credential" type="password" :placeholder="profile.configured ? t('adminPaymentProfiles.keepCredential') : ''" autocomplete="new-password" /></UFormField>
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
</style>
