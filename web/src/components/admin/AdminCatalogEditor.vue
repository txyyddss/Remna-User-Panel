<script setup lang="ts">
import { reactive, watch } from 'vue'
import { CheckboxIndicator, CheckboxRoot } from 'reka-ui'
import { PhCheck, PhFloppyDisk, PhX } from '@phosphor-icons/vue'

import type { Combo, Money, ResetCadence, SquadProduct } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

type RolloverCombo = Combo & { rolloverMinRemainingBps?: number; rolloverMaxTxbMinor?: string; rolloverMax?: Money }

const props = defineProps<{ combo?: Combo; squads: readonly SquadProduct[]; busy: boolean }>()
const emit = defineEmits<{
  cancel: []
  save: [payload: Record<string, unknown>]
}>()

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
      <h3>{{ combo ? `Edit ${combo.name}` : 'New combo' }}</h3>
      <button class="icon-button" type="button" aria-label="Close editor" @click="$emit('cancel')"><PhX :size="19" /></button>
    </div>
    <label><span>Name</span><input v-model.trim="draft.name" required maxlength="80" /></label>
    <label class="catalog-editor__wide"><span>Description</span><textarea v-model.trim="draft.description" required rows="3" /></label>
    <div class="catalog-editor__wide markdown-preview"><span>Preview</span><MarkdownContent :source="draft.description || 'Description preview'" compact /></div>
    <TxbAmountField id="combo-price" v-model="draft.priceTxb" label="Price" min-minor="1" required />
    <label><span>Validity days</span><input v-model.number="draft.validityDays" required type="number" min="1" /></label>
    <label><span>Traffic limit, bytes</span><input v-model="draft.trafficLimitBytes" required inputmode="numeric" pattern="[0-9]+" /></label>
    <label><span>Reset cadence</span><select v-model="draft.resetStrategy"><option>DAY</option><option>WEEK</option><option>MONTH</option></select></label>
    <label><span>Minimum remaining traffic, percent</span><input v-model.number="draft.rolloverMinRemainingPercent" required type="number" min="0" max="100" step="0.01" /></label>
    <TxbAmountField id="combo-rollover-max" v-model="draft.rolloverMaxTxb" label="Maximum rollover" min-minor="0" required />
    <fieldset class="catalog-editor__wide squad-picker"><legend>Included imported squads</legend><label v-for="squad in squads" :key="squad.id" class="squad-picker__option"><span><strong>{{ squad.name }}</strong><small>{{ squad.remnaSquadUuid }}</small></span><CheckboxRoot class="checkbox-control" :model-value="draft.squadProductIds.includes(squad.id)" @update:model-value="toggleSquad(squad.id)"><CheckboxIndicator class="checkbox-indicator"><PhCheck :size="16" weight="bold" /></CheckboxIndicator></CheckboxRoot></label><p v-if="!squads.length">Import Remnawave squads before assigning them.</p></fieldset>
    <SwitchField id="combo-active" v-model="draft.active" class="catalog-editor__wide" label="Available to users" help="Existing purchases keep their saved terms when disabled." />
    <button class="button button--primary catalog-editor__wide" type="submit" :disabled="busy">
      <PhFloppyDisk :size="18" /> {{ busy ? 'Saving' : 'Save combo' }}
    </button>
  </form>
</template>

<style scoped>
.markdown-preview { padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.markdown-preview > span { display: block; margin-bottom: 0.45rem; color: var(--text-faint); font-size: 0.68rem; font-weight: 700; }
.squad-picker { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 0.5rem; margin: 0; padding: 0; border: 0; }
.squad-picker legend { grid-column: 1 / -1; margin-bottom: 0.3rem; color: var(--text-muted); font-size: 0.78rem; font-weight: 700; }
.squad-picker__option { min-height: 54px; display: grid !important; grid-template-columns: minmax(0, 1fr) auto; align-items: center; padding: 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.squad-picker__option strong, .squad-picker__option small { display: block; }
.squad-picker__option small { overflow: hidden; margin-top: 0.2rem; color: var(--text-faint); font-family: var(--font-mono); font-size: 0.6rem; text-overflow: ellipsis; white-space: nowrap; }
</style>
