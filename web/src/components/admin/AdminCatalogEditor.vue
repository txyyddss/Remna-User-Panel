<script setup lang="ts">
import { reactive, watch } from 'vue'
import { CheckboxIndicator, CheckboxRoot } from 'reka-ui'
import { PhCheck, PhFloppyDisk, PhX } from '@phosphor-icons/vue'

import type { Combo, Money, ResetCadence, SquadProduct } from '@/api/types'
import MarkdownEditorField from '@/components/common/MarkdownEditorField.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useI18n } from '@/i18n'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

type RolloverCombo = Combo & { rolloverMinRemainingBps?: number; rolloverMaxTxbMinor?: string; rolloverMax?: Money }

const props = defineProps<{ combo?: Combo; squads: readonly SquadProduct[]; busy: boolean }>()
const emit = defineEmits<{
  cancel: []
  save: [payload: Record<string, unknown>]
}>()
const { t } = useI18n()

const draft = reactive({
  name: '',
  description: '',
  priceTxb: '',
  validityDays: 30,
  trafficLimitBytes: '',
  resetStrategy: 'MONTH' as ResetCadence,
  squadProductIds: [] as string[],
  rolloverMinRemainingPercent: 0,
  rolloverMaxTxb: '0.00',
  active: true,
})

watch(() => props.combo, (combo) => {
  const rollover = combo as RolloverCombo | undefined
  Object.assign(draft, combo ? {
    name: combo.name,
    description: combo.description,
    priceTxb: txbInputFromMinor(combo.price.minor),
    validityDays: combo.validityDays,
    trafficLimitBytes: combo.trafficLimitBytes,
    resetStrategy: combo.resetStrategy,
    squadProductIds: combo.includedSquads.map((squad) => squad.id),
    rolloverMinRemainingPercent: (rollover?.rolloverMinRemainingBps ?? 0) / 100,
    rolloverMaxTxb: txbInputFromMinor(rollover?.rolloverMax?.minor ?? rollover?.rolloverMaxTxbMinor ?? '0'),
    active: combo.active,
  } : {
    name: '', description: '', priceTxb: '', validityDays: 30,
    trafficLimitBytes: '', resetStrategy: 'MONTH', squadProductIds: [],
    rolloverMinRemainingPercent: 0, rolloverMaxTxb: '0.00', active: true,
  })
}, { immediate: true })

function submit(): void {
  const priceTxbMinor = moneyFromTxbInput(draft.priceTxb)
  const rolloverMaxTxbMinor = moneyFromTxbInput(draft.rolloverMaxTxb)
  if (!priceTxbMinor || rolloverMaxTxbMinor === '') return
  emit('save', {
    name: draft.name,
    description: draft.description,
    priceTxbMinor,
    validityDays: draft.validityDays,
    trafficLimitBytes: draft.trafficLimitBytes,
    resetStrategy: draft.resetStrategy,
    squadProductIds: [...draft.squadProductIds],
    rolloverMinRemainingBps: Math.round(draft.rolloverMinRemainingPercent * 100),
    rolloverMaxTxbMinor,
    active: draft.active,
  })
}

function toggleSquad(id: string): void {
  draft.squadProductIds = draft.squadProductIds.includes(id)
    ? draft.squadProductIds.filter((value) => value !== id)
    : [...draft.squadProductIds, id]
}
</script>

<template>
  <form class="catalog-editor" @submit.prevent="submit">
    <div class="catalog-editor__heading">
      <h3>{{ combo ? t('adminCatalogEditor.edit', { name: combo.name }) : t('adminCatalogEditor.new') }}</h3>
      <button class="icon-button" type="button" :aria-label="t('adminCatalogEditor.close')" @click="$emit('cancel')"><PhX :size="19" /></button>
    </div>
    <label><span>{{ t('adminCatalogEditor.name') }}</span><input v-model.trim="draft.name" required maxlength="80" /></label>
    <MarkdownEditorField v-model="draft.description" class="catalog-editor__wide" :label="t('adminCatalogEditor.description')" :placeholder="t('adminCatalogEditor.descriptionPlaceholder')" required :maxlength="2000" />
    <TxbAmountField id="combo-price" v-model="draft.priceTxb" :label="t('adminCatalogEditor.price')" min-minor="1" required />
    <label><span>{{ t('adminCatalogEditor.validityDays') }}</span><input v-model.number="draft.validityDays" required type="number" min="1" /></label>
    <label><span>{{ t('adminCatalogEditor.trafficLimit') }}</span><input v-model="draft.trafficLimitBytes" required inputmode="numeric" pattern="[0-9]+" /></label>
    <label><span>{{ t('adminCatalogEditor.resetCadence') }}</span><select v-model="draft.resetStrategy"><option>DAY</option><option>WEEK</option><option>MONTH</option></select></label>
    <label><span>{{ t('adminCatalogEditor.rolloverMinimum') }}</span><input v-model.number="draft.rolloverMinRemainingPercent" required type="number" min="0" max="100" step="0.01" /></label>
    <TxbAmountField id="combo-rollover-max" v-model="draft.rolloverMaxTxb" :label="t('adminCatalogEditor.rolloverMaximum')" min-minor="0" required />
    <fieldset class="catalog-editor__wide squad-picker"><legend>{{ t('adminCatalogEditor.includedSquads') }}</legend><label v-for="squad in squads" :key="squad.id" class="squad-picker__option"><span><strong>{{ squad.name }}</strong><small>{{ squad.remnaSquadUuid }}</small></span><CheckboxRoot class="checkbox-control" :model-value="draft.squadProductIds.includes(squad.id)" @update:model-value="toggleSquad(squad.id)"><CheckboxIndicator class="checkbox-indicator"><PhCheck :size="16" weight="bold" /></CheckboxIndicator></CheckboxRoot></label><p v-if="!squads.length">{{ t('adminCatalogEditor.noSquads') }}</p></fieldset>
    <SwitchField id="combo-active" v-model="draft.active" class="catalog-editor__wide" :label="t('adminCatalogEditor.available')" :help="t('adminCatalogEditor.liveTermsHint')" />
    <button class="button button--primary catalog-editor__wide" type="submit" :disabled="busy">
      <PhFloppyDisk :size="18" /> {{ busy ? t('common.saving') : t('adminCatalogEditor.save') }}
    </button>
  </form>
</template>

<style scoped>
.squad-picker { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 0.5rem; margin: 0; padding: 0; border: 0; }
.squad-picker legend { grid-column: 1 / -1; margin-bottom: 0.3rem; color: var(--text-muted); font-size: 0.78rem; font-weight: 700; }
.squad-picker__option { min-height: 54px; display: grid !important; grid-template-columns: minmax(0, 1fr) auto; align-items: center; padding: 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.squad-picker__option strong, .squad-picker__option small { display: block; }
.squad-picker__option small { overflow: hidden; margin-top: 0.2rem; color: var(--text-faint); font-family: var(--font-mono); font-size: 0.6rem; text-overflow: ellipsis; white-space: nowrap; }
</style>
