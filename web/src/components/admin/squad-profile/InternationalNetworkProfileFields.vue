<script setup lang="ts">
import { computed } from 'vue'

import { countryOptions } from '@/components/squad-profile/profile'
import { useI18n } from '@/i18n'
import SquadPortField from './SquadPortField.vue'
import { updateDraft, type ProfileDraft } from './profileForm'

const props = defineProps<{ draft: ProfileDraft }>()
const emit = defineEmits<{ 'update:draft': [draft: ProfileDraft] }>()
const { t, locale } = useI18n()
const countryItems = computed(() => countryOptions(locale.value))

function update(patch: Partial<ProfileDraft>): void {
  emit('update:draft', updateDraft(props.draft, patch))
}
</script>

<template>
  <div class="profile-fields">
    <SquadPortField id="squad-international-port" :model-value="draft.portMbps" :label="t('squadProfile.port')" :unlimited="draft.unlimited" allow-unlimited @update:model-value="update({ portMbps: $event })" @update:unlimited="update({ unlimited: $event })" />
    <UFormField name="squad-country" :label="t('squadProfile.country')" required>
      <USelect :model-value="draft.countryCode" :items="countryItems" value-key="value" :placeholder="t('squadProfile.countryPlaceholder')" @update:model-value="update({ countryCode: $event })" />
    </UFormField>
    <UFormField name="squad-upstream" :label="t('squadProfile.upstreamCarriers')" :hint="t('squadProfile.upstreamCarriersHint')" required>
      <UTextarea :model-value="draft.upstreamCarriers" :rows="2" :placeholder="t('squadProfile.upstreamCarriersPlaceholder')" autoresize @update:model-value="update({ upstreamCarriers: $event })" />
    </UFormField>
  </div>
</template>

<style scoped>
.profile-fields { display: grid; gap: 0.75rem; }
</style>
