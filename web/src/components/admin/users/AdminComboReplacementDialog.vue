<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import type { AdminEntitlement, Combo, ComboReplacementRequest, SquadProduct } from '@/api/adminOperations'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useI18n } from '@/i18n'

const props = defineProps<{
  open: boolean
  current: AdminEntitlement | null
  combos: Combo[]
  squads: SquadProduct[]
  busy: boolean
  optionsLoading: boolean
  error?: string | null
}>()
const emit = defineEmits<{ 'update:open': [value: boolean]; replace: [body: ComboReplacementRequest] }>()
const { t } = useI18n()
const draft = reactive({ comboId: '', addonSquadUuids: [] as string[], reason: '' })

const comboItems = computed(() => props.combos.map((combo) => ({ value: combo.id, label: combo.name })))
const squadItems = computed(() => props.squads.map((squad) => ({ value: squad.remnaSquadUuid, label: squad.name })))
const canSubmit = computed(() => draft.comboId !== '' && draft.reason.trim().length >= 3)

watch(() => [props.open, props.current, props.combos] as const, ([open, current]) => {
  if (!open) return
  const combo = props.combos.find((candidate) => candidate.id === current?.comboId)
  const included = new Set(combo?.includedSquads.map((squad) => squad.remnaSquadUuid) ?? [])
  draft.comboId = current?.comboId ?? props.combos[0]?.id ?? ''
  draft.addonSquadUuids = current?.squadUuids.filter((uuid) => !included.has(uuid)) ?? []
  draft.reason = ''
}, { deep: true })

function submit(): void {
  if (!canSubmit.value) return
  const combo = props.combos.find((candidate) => candidate.id === draft.comboId)
  const included = new Set(combo?.includedSquads.map((squad) => squad.remnaSquadUuid) ?? [])
  emit('replace', {
    comboId: draft.comboId,
    addonSquadUuids: draft.addonSquadUuids.filter((uuid) => !included.has(uuid)),
    reason: draft.reason.trim(),
  })
}
</script>

<template>
  <UModal :open="open" :title="t('adminUserProfile.replaceTitle')" :description="t('adminUserProfile.replaceHint')" :dismissible="!busy" :ui="{ footer: 'justify-end' }" @update:open="emit('update:open', $event)">
    <template #body>
      <UAlert color="info" variant="soft" icon="i-ph-coins" :title="t('adminUserProfile.noCharge')" :description="t('adminUserProfile.noChargeHint')" />
      <UForm id="combo-replacement" :state="draft" class="form-stack" @submit="submit">
        <UFormField name="comboId" :label="t('adminUserProfile.combo')" required><USelect v-model="draft.comboId" class="w-full" :items="comboItems" value-key="value" :loading="optionsLoading" /></UFormField>
        <UFormField name="addonSquadUuids" :label="t('adminUserProfile.optionalSquads')"><USelectMenu v-model="draft.addonSquadUuids" class="w-full" :items="squadItems" value-key="value" label-key="label" multiple :loading="optionsLoading" /></UFormField>
        <UFormField name="reason" :label="t('adminReason.reason')" required><UTextarea v-model.trim="draft.reason" :rows="3" :minlength="3" :maxlength="500" /></UFormField>
        <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      </UForm>
    </template>
    <template #footer>
      <UButton color="neutral" variant="outline" :label="t('common.cancel')" :disabled="busy" @click="emit('update:open', false)" />
      <UButton type="submit" form="combo-replacement" icon="i-ph-swap" :label="busy ? t('common.working') : t('adminUserProfile.applyReplacement')" :loading="busy" :disabled="!canSubmit || busy" />
    </template>
  </UModal>
</template>
