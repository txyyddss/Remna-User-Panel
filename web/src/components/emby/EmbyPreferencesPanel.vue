<script setup lang="ts">
import { reactive, watch } from 'vue'
import { PhFloppyDisk, PhKey, PhUserCircle } from '@phosphor-icons/vue'

import type { EmbyAccount, EmbyLibrary, EmbyRating } from '@/api/features'
import StatusBadge from '@/components/common/StatusBadge.vue'
import EmbyLibraryPicker from './EmbyLibraryPicker.vue'

const props = defineProps<{
  account: Omit<EmbyAccount, 'libraryIds'> & { readonly libraryIds: readonly string[] }
  ratings: readonly EmbyRating[]
  libraries: readonly EmbyLibrary[]
  busy: 'preferences' | 'password' | 'setup' | null
}>()
const emit = defineEmits<{
  save: [payload: { maxParentalRating: number | null; libraryIds: string[] }]
  changePassword: [password: string]
}>()

const draft = reactive({ maxParentalRating: null as number | null, libraryIds: [] as string[], password: '' })
watch(() => props.account, (account) => {
  draft.maxParentalRating = account.maxParentalRating
  draft.libraryIds = [...account.libraryIds]
  draft.password = ''
}, { immediate: true })

function toggleLibrary(id: string): void {
  draft.libraryIds = draft.libraryIds.includes(id)
    ? draft.libraryIds.filter((value) => value !== id)
    : [...draft.libraryIds, id]
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
      <span class="feature-icon"><PhUserCircle :size="24" /></span>
      <div><h2>{{ account.username }}</h2><p>Your Emby sign-in name.</p></div>
      <StatusBadge :tone="account.status === 'active' ? 'success' : account.status === 'failed' ? 'danger' : 'warning'" :label="account.status.replace('_', ' ')" />
      <p v-if="account.errorMessage" class="field-error">{{ account.errorMessage }}</p>
    </section>
    <form class="section-block emby-form" @submit.prevent="emit('save', { maxParentalRating: draft.maxParentalRating, libraryIds: [...draft.libraryIds] })">
      <div class="section-heading section-heading--stacked"><h2>Viewing preferences</h2><p>Restrictions are reapplied by the server on every update.</p></div>
      <label><span class="field-label">Maximum parental rating</span><select v-model="draft.maxParentalRating" class="compact-select" :disabled="account.status !== 'active'"><option :value="null">No rating ceiling</option><option v-for="rating in ratings" :key="rating.value" :value="rating.value">{{ rating.name }}</option></select></label>
      <EmbyLibraryPicker :libraries="libraries" :selected-ids="draft.libraryIds" :disabled="Boolean(busy) || account.status !== 'active'" @toggle="toggleLibrary" />
      <button class="button button--primary" type="submit" :disabled="Boolean(busy) || account.status !== 'active'"><PhFloppyDisk :size="18" />{{ busy === 'preferences' ? 'Saving' : 'Save preferences' }}</button>
    </form>
    <form class="section-block emby-form" autocomplete="off" @submit.prevent="savePassword">
      <div class="section-heading section-heading--stacked"><h2>Change password</h2><p>The new password is sent directly to Emby and not retained.</p></div>
      <label><span class="field-label">New password</span><span class="input-shell"><PhKey :size="18" /><input v-model="draft.password" type="password" minlength="8" required autocomplete="new-password" /></span></label>
      <button class="button button--secondary" type="submit" :disabled="Boolean(busy) || account.status !== 'active' || draft.password.length < 8">{{ busy === 'password' ? 'Changing password' : 'Change password' }}</button>
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
.emby-form > label { display: grid; gap: 0.4rem; }
</style>
