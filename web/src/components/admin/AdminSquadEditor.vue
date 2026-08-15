<script setup lang="ts">
import { reactive, shallowRef, watch } from 'vue'

import type { SquadProduct, SquadProductWrite, SquadProfileWrite } from '@/api/types'
import MarkdownEditorField from '@/components/common/MarkdownEditorField.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useI18n } from '@/i18n'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'
import AdminSquadProfileEditor from './squad-profile/AdminSquadProfileEditor.vue'
import { editableProfile } from './squad-profile/profileForm'

const props = defineProps<{
  squad: SquadProduct
  busy: boolean
}>()

const emit = defineEmits<{
  cancel: []
  save: [payload: SquadProductWrite]
}>()

const draft = reactive({
  description: '',
  priceTxb: '',
  visible: false,
  stockLimit: '',
  activationRequired: false,
  activationCode: '',
  profile: null as SquadProfileWrite | null,
})
const profileAttempted = shallowRef(false)
const { t } = useI18n()

watch(() => props.squad, (squad) => {
  Object.assign(draft, {
    description: squad.description,
    priceTxb: txbInputFromMinor(squad.price.minor),
    visible: squad.visible,
    stockLimit: squad.stockLimit === null ? '' : String(squad.stockLimit ?? ''),
    activationRequired: squad.activationRequired,
    activationCode: '',
    profile: editableProfile(squad.profile),
  })
  profileAttempted.value = false
}, { immediate: true })

function submit(): void {
  const priceTxbMinor = moneyFromTxbInput(draft.priceTxb)
  const stockLimitStr = String(draft.stockLimit ?? '').trim()
  const stockLimit = stockLimitStr === '' ? null : Number(stockLimitStr)
  if (!priceTxbMinor || (stockLimit !== null && (!Number.isInteger(stockLimit) || stockLimit < 0)) || !draft.profile) {
    profileAttempted.value = true
    return
  }
  emit('save', {
    remnaSquadUuid: props.squad.remnaSquadUuid,
    name: props.squad.name,
    description: draft.description.trim(),
    priceTxbMinor,
    visible: draft.visible,
    stockLimit,
    activationRequired: draft.activationRequired,
    ...(draft.activationCode.trim() ? { activationCode: draft.activationCode.trim() } : {}),
    profile: draft.profile,
  })
}
</script>

<template>
  <UForm :state="draft" class="catalog-editor squad-editor" @submit="submit">
    <div class="catalog-editor__heading">
      <div>
        <h3>{{ t('adminSquad.edit', { name: squad.name }) }}</h3>
        <p>{{ t('adminSquad.identityHint') }}</p>
      </div>
      <UButton color="neutral" variant="ghost" square icon="i-ph-x" :aria-label="t('adminSquad.close')" @click="emit('cancel')" />
    </div>
    <AdminSquadProfileEditor v-model:profile="draft.profile" :description="draft.description" :show-validation="profileAttempted" />
    <MarkdownEditorField v-model="draft.description" class="catalog-editor__wide" :label="t('adminSquad.description')" :placeholder="t('adminSquad.descriptionPlaceholder')" :maxlength="1000" />
    <TxbAmountField id="squad-price" v-model="draft.priceTxb" :label="t('adminSquad.price')" min-minor="0" required />
    <UFormField name="squad-stock-limit" :label="t('adminSquad.stockLimit')" :hint="t('adminSquad.stockLimitHint')"><UInput v-model="draft.stockLimit" type="number" min="0" step="1" inputmode="numeric" :placeholder="t('adminSquad.stockUnlimited')" /></UFormField>
    <SwitchField id="squad-visible" v-model="draft.visible" :label="t('adminSquad.visible')" :help="t('adminSquad.visibleHint')" />
    <SwitchField id="squad-activation-required" v-model="draft.activationRequired" :label="t('adminSquad.activationRequired')" :help="t('adminSquad.activationRequiredHint')" />
    <UFormField v-if="draft.activationRequired" name="squad-activation-code" :label="t('adminSquad.activationCode')" :hint="t('adminSquad.activationCodeHint')" required><UInput v-model="draft.activationCode" type="password" autocomplete="new-password" inputmode="text" /></UFormField>
    <UButton class="catalog-editor__wide" type="submit" icon="i-ph-floppy-disk" :loading="busy" :disabled="busy" :label="busy ? t('common.saving') : t('adminSquad.saveSquad')" />
  </UForm>
</template>

<style scoped>
</style>
