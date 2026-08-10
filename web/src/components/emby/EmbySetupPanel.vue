<script setup lang="ts">
import { reactive } from 'vue'
import { PhLockKey, PhShieldCheck } from '@phosphor-icons/vue'

import type { EmbyLibrary, EmbyRating } from '@/api/features'
import type { Money } from '@/api/types'
import { formatMoney } from '@/utils/format'
import EmbyLibraryPicker from './EmbyLibraryPicker.vue'

const props = defineProps<{
  price: Money
  ratings: readonly EmbyRating[]
  libraries: readonly EmbyLibrary[]
  busy: boolean
}>()
const emit = defineEmits<{ setup: [payload: { password: string; maxParentalRating: number | null; disabledLibraryIds: string[] }] }>()

const draft = reactive({ password: '', maxParentalRating: props.ratings[0]?.value ?? null as number | null, disabledLibraryIds: [] as string[] })

function toggleLibrary(id: string): void {
  draft.disabledLibraryIds = draft.disabledLibraryIds.includes(id)
    ? draft.disabledLibraryIds.filter((value) => value !== id)
    : [...draft.disabledLibraryIds, id]
}

function submit(): void {
  if (draft.password.length < 8) return
  emit('setup', { password: draft.password, maxParentalRating: draft.maxParentalRating, disabledLibraryIds: [...draft.disabledLibraryIds] })
  draft.password = ''
}
</script>

<template>
  <form class="section-block emby-form" autocomplete="off" @submit.prevent="submit">
    <div class="section-heading section-heading--stacked"><h2>{{ $t('emby.createAccount') }}</h2><p>{{ $t('emby.setupCost', { amount: formatMoney(price) }) }}</p></div>
    <label><span class="field-label">{{ $t('emby.initialPassword') }}</span><span class="input-shell"><PhLockKey :size="19" /><input v-model="draft.password" type="password" minlength="8" required autocomplete="new-password" /></span><small class="field-hint">{{ $t('emby.passwordHint') }}</small></label>
    <label><span class="field-label">{{ $t('emby.rating') }}</span><select v-model="draft.maxParentalRating" class="compact-select"><option :value="null">{{ $t('emby.noRating') }}</option><option v-for="rating in ratings" :key="rating.value" :value="rating.value">{{ rating.name }}</option></select></label>
    <EmbyLibraryPicker :libraries="libraries" :selected-ids="draft.disabledLibraryIds" :disabled="busy" @toggle="toggleLibrary" />
    <div class="emby-restrictions"><PhShieldCheck :size="20" /><p>{{ $t('emby.safetyControls') }}</p></div>
    <button class="button button--primary" type="submit" :disabled="busy || draft.password.length < 8">{{ busy ? $t('emby.startingSetup') : $t('emby.payAndCreate', { amount: formatMoney(price) }) }}</button>
  </form>
</template>

<style scoped>
.emby-form { display: grid; gap: 1rem; }
.emby-form > label { display: grid; gap: 0.4rem; }
.emby-restrictions { display: flex; align-items: flex-start; gap: 0.6rem; padding: 0.8rem; border-radius: var(--radius-control); color: var(--accent); background: var(--accent-soft); }
.emby-restrictions p { margin: 0; color: var(--text-muted); font-size: 0.72rem; line-height: 1.45; }
</style>
