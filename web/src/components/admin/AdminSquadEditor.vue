<script setup lang="ts">
import { reactive, shallowRef, watch } from 'vue'
import { CheckboxIndicator, CheckboxRoot } from 'reka-ui'
import { PhCheck, PhFloppyDisk, PhNetwork, PhX } from '@phosphor-icons/vue'

import { api } from '@/api/client'
import type { RemnaNode, SquadProduct, SquadProductWrite } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'
import InlineNotice from '@/components/common/InlineNotice.vue'
import MarkdownEditorField from '@/components/common/MarkdownEditorField.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useI18n } from '@/i18n'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

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
})
const nodes = shallowRef<RemnaNode[]>([])
const selectedNodeUuids = shallowRef<string[]>([])
const nodesBusy = shallowRef(false)
const nodesError = shallowRef<string | null>(null)
const { t } = useI18n()

async function loadNodes(): Promise<void> {
  nodesBusy.value = true
  nodesError.value = null
  try {
    nodes.value = (await api.getAdminSquadNodes(props.squad.id)).items
    selectedNodeUuids.value = nodes.value.filter((node) => node.accessible).map((node) => node.uuid)
  } catch (caught) {
    nodesError.value = t('adminSquad.loadError')
  } finally {
    nodesBusy.value = false
  }
}

watch(() => props.squad, (squad) => {
  Object.assign(draft, {
    description: squad.description,
    priceTxb: txbInputFromMinor(squad.price.minor),
    visible: squad.visible,
  })
  void loadNodes()
}, { immediate: true })

function toggleNode(uuid: string): void {
  selectedNodeUuids.value = selectedNodeUuids.value.includes(uuid)
    ? selectedNodeUuids.value.filter((value) => value !== uuid)
    : [...selectedNodeUuids.value, uuid]
}

async function saveNodes(): Promise<void> {
  if (nodesBusy.value) return
  nodesBusy.value = true
  nodesError.value = null
  try {
    nodes.value = (await api.updateAdminSquadNodes(props.squad.id, selectedNodeUuids.value)).items
    selectedNodeUuids.value = nodes.value.filter((node) => node.accessible).map((node) => node.uuid)
  } catch (caught) {
    nodesError.value = t('adminSquad.saveError')
  } finally {
    nodesBusy.value = false
  }
}

function submit(): void {
  const priceTxbMinor = moneyFromTxbInput(draft.priceTxb)
  if (!priceTxbMinor) return
  emit('save', {
    remnaSquadUuid: props.squad.remnaSquadUuid,
    name: props.squad.name,
    description: draft.description.trim(),
    priceTxbMinor,
    visible: draft.visible,
  })
}
</script>

<template>
  <form class="catalog-editor squad-editor" @submit.prevent="submit">
    <div class="catalog-editor__heading">
      <div>
        <h3>{{ t('adminSquad.edit', { name: squad.name }) }}</h3>
        <p>{{ t('adminSquad.identityHint') }}</p>
      </div>
      <button class="icon-button" type="button" :aria-label="t('adminSquad.close')" @click="$emit('cancel')">
        <PhX :size="19" />
      </button>
    </div>
    <MarkdownEditorField v-model="draft.description" class="catalog-editor__wide" :label="t('adminSquad.description')" :placeholder="t('adminSquad.descriptionPlaceholder')" :maxlength="1000" />
    <TxbAmountField id="squad-price" v-model="draft.priceTxb" :label="t('adminSquad.price')" min-minor="0" required />
    <SwitchField id="squad-visible" v-model="draft.visible" :label="t('adminSquad.visible')" :help="t('adminSquad.visibleHint')" />
    <section class="node-assignment catalog-editor__wide">
      <div class="node-assignment__heading"><div><h4><PhNetwork :size="18" /> {{ t('adminSquad.nodes') }}</h4><p>{{ t('adminSquad.nodeHint') }}</p></div><button class="button button--secondary" type="button" :disabled="nodesBusy" @click="saveNodes">{{ nodesBusy ? t('adminSquad.verifying') : t('adminSquad.saveNodes') }}</button></div>
      <InlineNotice v-if="nodesError" tone="warning">{{ nodesError }}</InlineNotice>
      <div class="node-list">
        <label v-for="node in nodes" :key="node.uuid" class="node-row">
          <CountryFlag :code="node.countryCode" />
          <span><strong>{{ node.name }}</strong><small>{{ t('adminSquad.nodeTraffic', { country: node.countryCode || t('adminSquad.isoUnavailable'), multiplier: node.consumptionMultiplier.toFixed(2) }) }}</small></span>
          <CheckboxRoot class="checkbox-control" :model-value="selectedNodeUuids.includes(node.uuid)" :disabled="nodesBusy" @update:model-value="toggleNode(node.uuid)"><CheckboxIndicator class="checkbox-indicator"><PhCheck :size="16" weight="bold" /></CheckboxIndicator></CheckboxRoot>
        </label>
        <p v-if="!nodesBusy && !nodes.length" class="field-hint">{{ t('adminSquad.noNodes') }}</p>
      </div>
    </section>
    <button class="button button--primary catalog-editor__wide" type="submit" :disabled="busy">
      <PhFloppyDisk :size="18" /> {{ busy ? t('common.saving') : t('adminSquad.saveSquad') }}
    </button>
  </form>
</template>

<style scoped>
.node-assignment { display: grid; gap: 0.7rem; padding: 0.8rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.node-assignment__heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 0.8rem; }
.node-assignment__heading h4, .node-assignment__heading p { margin: 0; }
.node-assignment__heading h4 { display: flex; align-items: center; gap: 0.4rem; font-size: 0.82rem; }
.node-assignment__heading p { margin-top: 0.25rem; color: var(--text-faint); font-size: 0.66rem; line-height: 1.45; }
.node-list { display: grid; gap: 0.45rem; }
.node-row { min-height: 52px; display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.65rem; padding: 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.node-row strong, .node-row small { display: block; }
.node-row strong { font-size: 0.76rem; }
.node-row small { margin-top: 0.2rem; color: var(--text-faint); font-size: 0.63rem; }
</style>
