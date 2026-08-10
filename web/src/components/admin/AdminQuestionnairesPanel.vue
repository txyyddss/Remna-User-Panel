<script setup lang="ts">
import { onMounted, reactive, shallowRef } from 'vue'
import { PhArrowSquareOut, PhFileCsv, PhPencilSimple, PhPlus, PhStopCircle, PhTrash } from '@phosphor-icons/vue'

import type { QuestionnaireAdminRecord } from '@/api/features'
import { featuresApi } from '@/api/features'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { localizedError, useI18n } from '@/i18n'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'
import QuestionnaireImportWorkflow from './questionnaires/QuestionnaireImportWorkflow.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const items = shallowRef<QuestionnaireAdminRecord[]>([])
const loading = shallowRef(true)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const editingId = shallowRef<string | null | undefined>(undefined)
const importing = shallowRef<QuestionnaireAdminRecord | null>(null)
const deleting = shallowRef<QuestionnaireAdminRecord | null>(null)
const draft = reactive({ title: '', description: '', formUrl: '', rewardTxb: '5.00' })
const { t } = useI18n()

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    items.value = (await featuresApi.getAdminQuestionnaires()).items
  } catch (caught) {
    error.value = localizedError(caught, 'adminQuestionnaires.loadFailed')
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
    error.value = localizedError(caught, 'adminQuestionnaires.saveFailed')
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
    error.value = localizedError(caught, 'adminQuestionnaires.activateFailed')
  } finally {
    busy.value = false
  }
}

async function close(questionnaire: QuestionnaireAdminRecord): Promise<void> {
  if (!globalThis.confirm(t('adminQuestionnaires.closeConfirm', { name: questionnaire.title }))) return
  busy.value = true
  error.value = null
  try {
    await featuresApi.closeAdminQuestionnaire(questionnaire.id)
    await load()
  } catch (caught) {
    error.value = localizedError(caught, 'adminQuestionnaires.closeFailed')
  } finally {
    busy.value = false
  }
}

async function remove(): Promise<void> {
  if (!deleting.value || busy.value) return
  busy.value = true
  error.value = null
  try {
    await featuresApi.deleteAdminQuestionnaire(deleting.value.id)
    deleting.value = null
    await load()
  } catch (caught) {
    error.value = localizedError(caught, 'adminQuestionnaires.deleteFailed')
  } finally {
    busy.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading"><div><h2>{{ t('adminQuestionnaires.title') }}</h2><p>{{ t('adminQuestionnaires.copy') }}</p></div><button class="button button--primary" type="button" @click="edit()"><PhPlus :size="18" />{{ t('adminQuestionnaires.new') }}</button></div>
    <form v-if="editingId !== undefined" class="catalog-editor" @submit.prevent="save">
      <div class="catalog-editor__heading"><h3>{{ editingId ? t('adminQuestionnaires.edit') : t('adminQuestionnaires.draft') }}</h3></div>
      <label><span>{{ t('adminQuestionnaires.formTitle') }}</span><input v-model.trim="draft.title" required maxlength="120" /></label>
      <label><span>{{ t('adminQuestionnaires.formUrl') }}</span><input v-model.trim="draft.formUrl" type="url" required /></label>
      <label class="catalog-editor__wide"><span>{{ t('adminQuestionnaires.description') }}</span><textarea v-model.trim="draft.description" rows="3" required /></label>
      <TxbAmountField id="questionnaire-reward" v-model="draft.rewardTxb" :label="t('adminQuestionnaires.reward')" min-minor="1" required />
      <div class="button-row"><button class="button button--secondary" type="button" @click="editingId = undefined">{{ t('common.cancel') }}</button><button class="button button--primary" type="submit" :disabled="busy">{{ busy ? t('common.saving') : t('adminQuestionnaires.saveDraft') }}</button></div>
    </form>
    <p v-if="error" class="field-error admin-error" role="alert">{{ error }}</p>
    <div v-if="loading" class="admin-loading">{{ t('adminQuestionnaires.loading') }}</div>
    <div v-else class="admin-list">
      <article v-for="questionnaire in items" :key="questionnaire.id" class="admin-list-row questionnaire-row">
        <div><strong>{{ questionnaire.title }}</strong><small>{{ t('adminQuestionnaires.summary', { status: questionnaire.status, count: questionnaire.participantCount, reward: txbInputFromMinor(questionnaire.rewardTxbMinor) }) }}</small></div>
        <div class="row-actions">
          <button v-if="questionnaire.status === 'draft'" class="button button--secondary button--small" type="button" :disabled="busy" @click="activate(questionnaire.id)"><PhArrowSquareOut :size="17" />{{ t('adminQuestionnaires.activate') }}</button>
          <button v-if="questionnaire.status === 'active'" class="button button--secondary button--small" type="button" :disabled="busy" @click="close(questionnaire)"><PhStopCircle :size="17" />{{ t('common.close') }}</button>
          <button v-if="['active', 'closed', 'failed'].includes(questionnaire.status)" class="button button--secondary button--small" type="button" @click="importing = questionnaire"><PhFileCsv :size="17" />{{ t('adminQuestionnaires.importCsv') }}</button>
          <button class="icon-button" type="button" :aria-label="t('adminQuestionnaires.editNamed', { name: questionnaire.title })" @click="edit(questionnaire)"><PhPencilSimple :size="18" /></button>
          <button class="icon-button icon-button--danger" type="button" :aria-label="t('adminQuestionnaires.deleteNamed', { name: questionnaire.title })" @click="deleting = questionnaire"><PhTrash :size="18" /></button>
        </div>
      </article>
      <div v-if="!items.length" class="empty-inline"><div><h3>{{ t('adminQuestionnaires.none') }}</h3><p>{{ t('adminQuestionnaires.noneHint') }}</p></div></div>
    </div>
    <QuestionnaireImportWorkflow v-if="importing" :questionnaire="importing" @close="importing = null" />
    <ConfirmDialog :open="Boolean(deleting)" :title="t('adminQuestionnaires.deleteTitle', { name: deleting?.title ?? t('adminQuestionnaires.questionnaire') })" :description="t('adminQuestionnaires.deleteDescription')" :confirm-label="t('adminQuestionnaires.deletePermanently')" :busy="busy" danger @update:open="!$event && (deleting = null)" @confirm="remove" />
  </section>
</template>

<style scoped>
.admin-error { margin: 1rem; }
.questionnaire-row { align-items: center; }
.row-actions { display: flex; align-items: center; flex-wrap: wrap; justify-content: flex-end; gap: 0.4rem; }
</style>
