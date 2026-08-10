<script setup lang="ts">
import { computed, reactive } from 'vue'

import type { EmbyLibrary, EmbyRating } from '@/api/features'
import type { Money } from '@/api/types'
import { useI18n } from '@/i18n'
import { formatMoney } from '@/utils/format'
import EmbyLibraryPicker from './EmbyLibraryPicker.vue'

const props = defineProps<{
  price: Money
  ratings: readonly EmbyRating[]
  libraries: readonly EmbyLibrary[]
  busy: boolean
}>()
const emit = defineEmits<{ setup: [payload: { password: string; maxParentalRating: number | null; disabledLibraryIds: string[] }] }>()
const { t } = useI18n()

const draft = reactive({ password: '', maxParentalRating: props.ratings[0]?.value ?? null as number | null, disabledLibraryIds: [] as string[] })
const ratingItems = computed(() => [
  { label: t('emby.noRating'), value: null },
  ...props.ratings.map((rating) => ({ label: rating.name, value: rating.value })),
])

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
    <UFormField name="password" :label="$t('emby.initialPassword')" :description="$t('emby.passwordHint')" required>
      <UInput v-model="draft.password" icon="i-ph-lock-key" type="password" :minlength="8" required autocomplete="new-password" />
    </UFormField>
    <UFormField name="rating" :label="$t('emby.rating')">
      <USelect v-model="draft.maxParentalRating" :items="ratingItems" value-key="value" />
    </UFormField>
    <EmbyLibraryPicker :libraries="libraries" :selected-ids="draft.disabledLibraryIds" :disabled="busy" @toggle="toggleLibrary" />
    <UAlert color="success" variant="soft" icon="i-ph-shield-check" :description="$t('emby.safetyControls')" />
    <UButton type="submit" :disabled="busy || draft.password.length < 8" :loading="busy" :label="busy ? $t('emby.startingSetup') : $t('emby.payAndCreate', { amount: formatMoney(price) })" />
  </form>
</template>

<style scoped>
.emby-form { display: grid; gap: 1rem; }
</style>
