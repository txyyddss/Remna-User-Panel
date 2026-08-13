<script setup lang="ts">
import { useI18n } from '@/i18n'
import CountryCodeField from './CountryCodeField.vue'
import SquadPortField from './SquadPortField.vue'
import { updateDraft, type ProfileDraft } from './profileForm'

const props = defineProps<{ draft: ProfileDraft }>()
const emit = defineEmits<{ 'update:draft': [draft: ProfileDraft] }>()
const { t } = useI18n()

function update(patch: Partial<ProfileDraft>): void {
  emit('update:draft', updateDraft(props.draft, patch))
}
</script>

<template>
  <div class="profile-fields">
    <div class="carrier-grid">
      <UFormField name="squad-ct" :label="t('squadProfile.ct')" required><UInput :model-value="draft.ct" :placeholder="t('squadProfile.carrierPlaceholder')" @update:model-value="update({ ct: $event })" /></UFormField>
      <UFormField name="squad-cu" :label="t('squadProfile.cu')" required><UInput :model-value="draft.cu" :placeholder="t('squadProfile.carrierPlaceholder')" @update:model-value="update({ cu: $event })" /></UFormField>
      <UFormField name="squad-cm" :label="t('squadProfile.cm')" required><UInput :model-value="draft.cm" :placeholder="t('squadProfile.carrierPlaceholder')" @update:model-value="update({ cm: $event })" /></UFormField>
    </div>
    <SquadPortField id="squad-china-port" :model-value="draft.portMbps" :label="t('squadProfile.port')" :unlimited="draft.unlimited" allow-unlimited @update:model-value="update({ portMbps: $event })" @update:unlimited="update({ unlimited: $event })" />
    <CountryCodeField id="squad-country" :model-value="draft.countryCode" @update:model-value="update({ countryCode: $event })" />
  </div>
</template>

<style scoped>
.profile-fields, .carrier-grid { display: grid; gap: 0.75rem; }
@media (min-width: 640px) { .carrier-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
</style>
