<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'

import type { Combo, ResetCadence, SquadProduct } from '@/api/types'
import MarkdownEditorField from '@/components/common/MarkdownEditorField.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useI18n } from '@/i18n'
import { moneyFromTxbInput, trafficBytesFromInput, txbInputFromMinor } from '@/utils/format'

const props = defineProps<{ combo?: Combo; squads: readonly SquadProduct[]; busy: boolean }>()
const emit = defineEmits<{ cancel: []; save: [payload: Record<string, unknown>] }>()
const { t } = useI18n()
const resetCadences: ResetCadence[] = ['DAY', 'WEEK', 'MONTH_ROLLING']
const resetItems = computed(() => resetCadences.map((value) => ({ value, label: t(`adminCatalogEditor.reset.${value}`) })))
const draft = reactive({
  name: '', description: '', priceTxb: '', validityDays: 30, trafficLimitBytes: '',
  resetStrategy: 'MONTH_ROLLING' as ResetCadence, squadProductIds: [] as string[],
  rolloverMinRemainingPercent: 0, active: true,
})
const trafficInvalid = ref(false)

watch(() => props.combo, (combo) => {
  Object.assign(draft, combo ? {
    name: combo.name,
    description: combo.description,
    priceTxb: txbInputFromMinor(combo.price.minor),
    validityDays: combo.validityDays,
    trafficLimitBytes: combo.trafficLimitBytes,
    resetStrategy: combo.resetStrategy,
    squadProductIds: combo.includedSquads.map((squad) => squad.id),
    rolloverMinRemainingPercent: combo.rolloverMinRemainingBps / 100,
    active: combo.active,
  } : {
    name: '', description: '', priceTxb: '', validityDays: 30,
    trafficLimitBytes: '', resetStrategy: 'MONTH_ROLLING', squadProductIds: [], rolloverMinRemainingPercent: 0, active: true,
  })
}, { immediate: true })

function submit(): void {
  const priceTxbMinor = moneyFromTxbInput(draft.priceTxb)
  const trafficLimitBytes = trafficBytesFromInput(draft.trafficLimitBytes)
  trafficInvalid.value = trafficLimitBytes === ''
  if (!priceTxbMinor || trafficInvalid.value) return
  draft.trafficLimitBytes = trafficLimitBytes
  emit('save', {
    name: draft.name, description: draft.description, priceTxbMinor,
    validityDays: draft.validityDays, trafficLimitBytes,
    resetStrategy: draft.resetStrategy, squadProductIds: [...draft.squadProductIds],
    rolloverMinRemainingBps: Math.round(draft.rolloverMinRemainingPercent * 100), active: draft.active,
  })
}

function normalizeTrafficLimit(): void {
  const trafficLimitBytes = trafficBytesFromInput(draft.trafficLimitBytes)
  trafficInvalid.value = trafficLimitBytes === ''
  if (trafficLimitBytes) draft.trafficLimitBytes = trafficLimitBytes
}

function setSquad(id: string, selected: boolean): void {
  draft.squadProductIds = selected
    ? [...new Set([...draft.squadProductIds, id])]
    : draft.squadProductIds.filter((value) => value !== id)
}
</script>

<template>
  <form class="catalog-editor" @submit.prevent="submit">
    <div class="catalog-editor__heading">
      <h3>{{ combo ? t('adminCatalogEditor.edit', { name: combo.name }) : t('adminCatalogEditor.new') }}</h3>
      <UButton color="neutral" variant="ghost" square icon="i-ph-x" :aria-label="t('adminCatalogEditor.close')" @click="emit('cancel')" />
    </div>
    <UFormField name="combo-name" :label="t('adminCatalogEditor.name')" required><UInput v-model.trim="draft.name" class="w-full" :maxlength="80" /></UFormField>
    <MarkdownEditorField v-model="draft.description" class="catalog-editor__wide" :label="t('adminCatalogEditor.description')" :placeholder="t('adminCatalogEditor.descriptionPlaceholder')" required :maxlength="2000" />
    <TxbAmountField id="combo-price" v-model="draft.priceTxb" :label="t('adminCatalogEditor.price')" min-minor="1" required />
    <UFormField name="validity-days" :label="t('adminCatalogEditor.validityDays')" required><UInput v-model.number="draft.validityDays" class="w-full" type="number" :min="1" /></UFormField>
    <UFormField name="traffic-limit" :label="t('adminCatalogEditor.trafficLimit')" :hint="t('adminCatalogEditor.trafficLimitHint')" :error="trafficInvalid ? t('adminCatalogEditor.trafficLimitInvalid') : undefined" required><UInput v-model="draft.trafficLimitBytes" class="w-full" inputmode="text" @blur="normalizeTrafficLimit" /></UFormField>
    <UFormField name="reset-cadence" :label="t('adminCatalogEditor.resetCadence')"><USelect v-model="draft.resetStrategy" class="w-full" :items="resetItems" /></UFormField>
    <UFormField name="rollover-minimum" :label="t('adminCatalogEditor.rolloverMinimum')" required><UInput v-model.number="draft.rolloverMinRemainingPercent" class="w-full" type="number" :min="0" :max="100" :step="0.01" /></UFormField>
    <fieldset class="catalog-editor__wide squad-picker">
      <legend>{{ t('adminCatalogEditor.includedSquads') }}</legend>
      <div v-auto-animate class="squad-picker__items">
        <div v-for="squad in squads" :key="squad.id" class="squad-picker__option">
          <span><strong>{{ squad.name }}</strong><small>{{ squad.remnaSquadUuid }}</small></span>
          <UCheckbox :model-value="draft.squadProductIds.includes(squad.id)" :aria-label="squad.name" @update:model-value="setSquad(squad.id, Boolean($event))" />
        </div>
        <p v-if="!squads.length">{{ t('adminCatalogEditor.noSquads') }}</p>
      </div>
    </fieldset>
    <SwitchField id="combo-active" v-model="draft.active" class="catalog-editor__wide" :label="t('adminCatalogEditor.available')" :help="t('adminCatalogEditor.liveTermsHint')" />
    <UButton class="catalog-editor__wide" type="submit" icon="i-ph-floppy-disk" :loading="busy" :disabled="busy" :label="busy ? t('common.saving') : t('adminCatalogEditor.save')" />
  </form>
</template>

<style scoped>
.squad-picker { display: grid; gap: 0.5rem; margin: 0; padding: 0; border: 0; }
.squad-picker legend { margin-bottom: 0.3rem; color: var(--text-muted); font-size: 0.78rem; font-weight: 700; }
.squad-picker__items { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 0.5rem; }
.squad-picker__option { min-height: 54px; display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; padding: 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.squad-picker__option strong, .squad-picker__option small { display: block; }
.squad-picker__option small { overflow: hidden; margin-top: 0.2rem; color: var(--text-faint); font-family: var(--font-mono); font-size: 0.6rem; text-overflow: ellipsis; white-space: nowrap; }
</style>
