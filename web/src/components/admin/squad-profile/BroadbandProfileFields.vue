<script setup lang="ts">
import SwitchField from '@/components/common/SwitchField.vue'
import { useI18n } from '@/i18n'
import SquadPortField from './SquadPortField.vue'
import type { ProfileDraft } from './profileForm'

const props = defineProps<{ draft: ProfileDraft }>()
const emit = defineEmits<{ 'update:draft': [draft: ProfileDraft] }>()
const { t } = useI18n()

function update(patch: Partial<ProfileDraft>): void {
  emit('update:draft', { ...props.draft, ...patch })
}
</script>

<template>
  <div class="profile-fields">
    <UFormField name="squad-isp" :label="t('squadProfile.isp')" required>
      <UInput :model-value="draft.isp" :placeholder="t('squadProfile.ispPlaceholder')" @update:model-value="update({ isp: $event })" />
    </UFormField>
    <SquadPortField id="squad-broadband-port" :model-value="draft.portMbps" :label="t('squadProfile.port')" :unlimited="draft.unlimited" @update:model-value="update({ portMbps: $event })" @update:unlimited="update({ unlimited: $event })" />
    <SwitchField id="squad-dynamic" :model-value="draft.dynamic" :label="t('squadProfile.dynamic')" :help="t('squadProfile.dynamicHint')" @update:model-value="update({ dynamic: $event })" />
    <UFormField name="squad-location" :label="t('squadProfile.location')" required>
      <UInput :model-value="draft.location" :placeholder="t('squadProfile.locationBroadbandPlaceholder')" @update:model-value="update({ location: $event })" />
    </UFormField>
  </div>
</template>

<style scoped>
.profile-fields { display: grid; gap: 0.75rem; }
</style>
