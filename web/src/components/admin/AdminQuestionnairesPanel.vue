<script setup lang="ts">
import { onMounted, reactive, shallowRef } from 'vue'
import { PhArrowSquareOut, PhFileCsv, PhPencilSimple, PhPlus, PhStopCircle } from '@phosphor-icons/vue'

import type { QuestionnaireAdminRecord } from '@/api/features'
import { featuresApi } from '@/api/features'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'
import QuestionnaireImportWorkflow from './questionnaires/QuestionnaireImportWorkflow.vue'

const items = shallowRef<QuestionnaireAdminRecord[]>([])
const loading = shallowRef(true)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const editingId = shallowRef<string | null | undefined>(undefined)
const importing = shallowRef<QuestionnaireAdminRecord | null>(null)
const draft = reactive({ title: '', description: '', formUrl: '', rewardTxb: '5.00' })

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    items.value = (await featuresApi.getAdminQuestionnaires()).items
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'Questionnaires could not be loaded.'
  } finally {
    loading.value = false
  }
}

function edit(questionnaire?: QuestionnaireAdminRecord): void {
  editingId.value = questionnaire?.id ?? null
  Object.assign(draft, questionnaire ? {
    title: questionnaire.title,
    description: questionnaire.description,
    formUrl: questionnaire.formUrl,
    rewardTxb: txbInputFromMinor(questionnaire.rewardTxbMinor),
  } : { title: '', description: '', formUrl: '', rewardTxb: '5.00' })
}

async function save(): Promise<void> {
  const rewardTxbMinor = moneyFromTxbInput(draft.rewardTxb)
  if (!rewardTxbMinor) return
  busy.value = true
  error.value = null
  try {
    await featuresApi.saveAdminQuestionnaire(editingId.value ?? null, {
      title: draft.title,
      description: draft.description,
      formUrl: draft.formUrl,
      rewardTxbMinor,
    })
    editingId.value = undefined
    await load()
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'The questionnaire could not be saved.'
  } finally {
    busy.value = false
  }
}

async function activate(id: string): Promise<void> {
  busy.value = true
  error.value = null
  try {
    await featuresApi.activateAdminQuestionnaire(id)
    await load()
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'Activation failed.'
  } finally {
    busy.value = false
  }
}

async function close(questionnaire: QuestionnaireAdminRecord): Promise<void> {
  if (!globalThis.confirm(`Close ${questionnaire.title}? New participation will stop.`)) return
  busy.value = true
  error.value = null
  try {
    await featuresApi.closeAdminQuestionnaire(questionnaire.id)
    await load()
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'The questionnaire could not be closed.'
  } finally {
    busy.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading"><div><h2>Questionnaires</h2><p>One active form, one validation code, one reward per participant.</p></div><button class="button button--primary" type="button" @click="edit()"><PhPlus :size="18" />New questionnaire</button></div>
    <form v-if="editingId !== undefined" class="catalog-editor" @submit.prevent="save">
      <div class="catalog-editor__heading"><h3>{{ editingId ? 'Edit questionnaire' : 'Draft questionnaire' }}</h3></div>
      <label><span>Title</span><input v-model.trim="draft.title" required maxlength="120" /></label>
      <label><span>HTTPS Google or Microsoft Forms URL</span><input v-model.trim="draft.formUrl" type="url" required /></label>
      <label class="catalog-editor__wide"><span>Description, Markdown</span><textarea v-model.trim="draft.description" rows="3" required /></label>
      <TxbAmountField id="questionnaire-reward" v-model="draft.rewardTxb" label="Reward" min-minor="1" required />
      <div class="button-row"><button class="button button--secondary" type="button" @click="editingId = undefined">Cancel</button><button class="button button--primary" type="submit" :disabled="busy">{{ busy ? 'Saving' : 'Save draft' }}</button></div>
    </form>
    <p v-if="error" class="field-error admin-error" role="alert">{{ error }}</p>
    <div v-if="loading" class="admin-loading">Loading questionnaires</div>
    <div v-else class="admin-list">
      <article v-for="questionnaire in items" :key="questionnaire.id" class="admin-list-row questionnaire-row">
        <div><strong>{{ questionnaire.title }}</strong><small>{{ questionnaire.status }} · {{ questionnaire.participantCount }} participants · {{ txbInputFromMinor(questionnaire.rewardTxbMinor) }} TXB</small></div>
        <div class="row-actions">
          <button v-if="questionnaire.status === 'draft'" class="button button--secondary button--small" type="button" :disabled="busy" @click="activate(questionnaire.id)"><PhArrowSquareOut :size="17" />Activate</button>
          <button v-if="questionnaire.status === 'active'" class="button button--secondary button--small" type="button" :disabled="busy" @click="close(questionnaire)"><PhStopCircle :size="17" />Close</button>
          <button v-if="['active', 'closed', 'failed'].includes(questionnaire.status)" class="button button--secondary button--small" type="button" @click="importing = questionnaire"><PhFileCsv :size="17" />Import CSV</button>
          <button class="icon-button" type="button" :aria-label="`Edit ${questionnaire.title}`" @click="edit(questionnaire)"><PhPencilSimple :size="18" /></button>
        </div>
      </article>
      <div v-if="!items.length" class="empty-inline"><div><h3>No questionnaires</h3><p>Create a draft to begin.</p></div></div>
    </div>
    <QuestionnaireImportWorkflow v-if="importing" :questionnaire="importing" @close="importing = null" />
  </section>
</template>

<style scoped>
.admin-error { margin: 1rem; }
.questionnaire-row { align-items: center; }
.row-actions { display: flex; align-items: center; flex-wrap: wrap; justify-content: flex-end; gap: 0.4rem; }
</style>
