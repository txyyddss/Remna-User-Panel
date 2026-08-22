<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import type { EmbyAccount, EmbyLibrary, EmbyRating } from '@/api/features'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import EmbyLibraryPicker from './EmbyLibraryPicker.vue'

const props = defineProps<{
  account: Omit<EmbyAccount, 'disabledLibraryIds'> & { readonly disabledLibraryIds: readonly string[] }
  ratings: readonly EmbyRating[]
  libraries: readonly EmbyLibrary[]
  busy: 'preferences' | 'password' | 'setup' | null
  blocked: boolean
}>()
const emit = defineEmits<{
  save: [payload: { maxParentalRating: number | null; disabledLibraryIds: string[] }]
  changePassword: [password: string]
}>()
const { t } = useI18n()

const draft = reactive({ maxParentalRating: null as number | null, disabledLibraryIds: [] as string[], password: '' })
const ratingItems = computed(() => [
  { label: t('emby.noRating'), value: null },
  ...props.ratings.map((rating) => ({ label: rating.name, value: rating.value })),
])

watch(() => props.account, (account) => {
  draft.maxParentalRating = account.maxParentalRating
  draft.disabledLibraryIds = [...account.disabledLibraryIds]
  draft.password = ''
}, { immediate: true })

function toggleLibrary(id: string): void {
  draft.disabledLibraryIds = draft.disabledLibraryIds.includes(id)
    ? draft.disabledLibraryIds.filter((value) => value !== id)
    : [...draft.disabledLibraryIds, id]
}

function savePassword(): void {
  if (draft.password.length < 8) return
  emit('changePassword', draft.password)
  draft.password = ''
}
</script>

<template>
  <div class="emby-account-grid">
    <section class="section-block emby-account-summary">
      <span class="feature-icon"><UIcon name="i-ph-user-circle" /></span>
      <div><h2>{{ account.username }}</h2><p>{{ $t('emby.usernameHint') }}</p></div>
      <StatusBadge :tone="account.status === 'active' ? 'success' : account.status === 'failed' ? 'danger' : 'warning'" :label="$t(`emby.status.${account.status}`)" />
      <p v-if="account.errorMessage" class="field-error">{{ $t('emby.provisioningFailed') }}</p>
    </section>
    <form class="section-block emby-form" @submit.prevent="emit('save', { maxParentalRating: draft.maxParentalRating, disabledLibraryIds: [...draft.disabledLibraryIds] })">
      <div class="section-heading section-heading--stacked"><h2>{{ $t('emby.preferences') }}</h2><p>{{ $t('emby.preferencesHint') }}</p></div>
      <UFormField name="rating" :label="$t('emby.rating')">
        <USelect v-model="draft.maxParentalRating" :items="ratingItems" value-key="value" :disabled="blocked || account.status !== 'active'" />
      </UFormField>
      <EmbyLibraryPicker :libraries="libraries" :selected-ids="draft.disabledLibraryIds" :disabled="blocked || Boolean(busy) || account.status !== 'active'" @toggle="toggleLibrary" />
      <UButton type="submit" icon="i-ph-floppy-disk" :disabled="blocked || Boolean(busy) || account.status !== 'active'" :loading="busy === 'preferences'" :label="busy === 'preferences' ? $t('common.saving') : $t('emby.savePreferences')" />
    </form>
    <form class="section-block emby-form" autocomplete="off" @submit.prevent="savePassword">
      <div class="section-heading section-heading--stacked"><h2>{{ $t('emby.changePassword') }}</h2><p>{{ $t('emby.changePasswordHint') }}</p></div>
      <UFormField name="password" :label="$t('emby.newPassword')" required>
        <UInput v-model="draft.password" icon="i-ph-key" type="password" :minlength="8" :disabled="blocked" required autocomplete="new-password" />
      </UFormField>
      <UButton type="submit" color="neutral" variant="outline" :disabled="blocked || Boolean(busy) || account.status !== 'active' || draft.password.length < 8" :loading="busy === 'password'" :label="busy === 'password' ? $t('emby.changingPassword') : $t('emby.changePassword')" />
    </form>
  </div>
</template>

<style scoped>
.emby-account-grid, .emby-form { display: grid; gap: 0.9rem; }
.emby-account-summary { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.75rem; }
.emby-account-summary h2, .emby-account-summary p { margin: 0; }
.emby-account-summary h2 { font-size: 1.15rem; }
.emby-account-summary p { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.72rem; }
.emby-account-summary > .field-error { grid-column: 1 / -1; }
@media (min-width: 1180px) { .emby-account-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; } .emby-account-summary { grid-column: 1 / -1; } }
</style>
