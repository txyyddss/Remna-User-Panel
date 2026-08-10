<script setup lang="ts">
import { computed, useTemplateRef, watch } from 'vue'
import { PhCheckCircle, PhFileCsv, PhUploadSimple, PhX } from '@phosphor-icons/vue'

import type { QuestionnaireAdminRecord } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useQuestionnaireImport } from '@/composables/useQuestionnaireImport'
import { useI18n } from '@/i18n'
import { txbInputFromMinor } from '@/utils/format'

const props = defineProps<{ questionnaire: QuestionnaireAdminRecord }>()
defineEmits<{ close: [] }>()

const fileInput = useTemplateRef<globalThis.HTMLInputElement>('fileInput')
const { preview, summary, codeColumn, busy, settlementImportId, report, error, upload, analyze, settle, reset } = useQuestionnaireImport(() => props.questionnaire.id)
const canSettle = computed(() => Boolean(summary.value && summary.value.matchedCount > 0 && !settlementImportId.value))
const { t } = useI18n()

watch(() => props.questionnaire.id, reset)

function chooseFile(): void {
  fileInput.value?.click()
}

function handleFile(event: globalThis.Event): void {
  const file = (event.target as globalThis.HTMLInputElement).files?.[0]
  if (file) void upload(file)
}
</script>

<template>
  <aside class="questionnaire-import" aria-labelledby="import-title">
    <header><div><h3 id="import-title">{{ t('questionnaireImport.settle', { name: questionnaire.title }) }}</h3><p>{{ t('questionnaireImport.copy') }}</p></div><button class="icon-button" type="button" :aria-label="t('questionnaireImport.close')" @click="$emit('close')"><PhX :size="19" /></button></header>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <InlineNotice v-if="settlementImportId && preview?.status !== 'settled'" :tone="preview?.status === 'failed' ? 'warning' : 'success'" :title="preview?.status === 'processing' ? t('questionnaireImport.running') : preview?.status === 'failed' ? t('questionnaireImport.failed') : t('questionnaireImport.queued')">{{ t('questionnaireImport.durable', { id: settlementImportId }) }}</InlineNotice>
    <InlineNotice v-if="report" tone="success" :title="t('questionnaireImport.complete')">{{ t('questionnaireImport.completeCopy', { count: report.rewardedCount, reward: txbInputFromMinor(report.rewardTxbMinor) }) }}</InlineNotice>
    <div v-if="!preview" class="csv-drop">
      <PhFileCsv :size="36" /><h4>{{ t('questionnaireImport.csv') }}</h4><p>{{ t('questionnaireImport.csvHint') }}</p>
      <input ref="fileInput" class="sr-only" type="file" accept=".csv,text/csv" @change="handleFile" />
      <button class="button button--secondary" type="button" :disabled="busy" @click="chooseFile"><PhUploadSimple :size="18" />{{ busy ? t('questionnaireImport.uploading') : t('questionnaireImport.choose') }}</button>
    </div>
    <template v-else>
      <div class="csv-meta"><span>{{ t('questionnaireImport.rows', { count: preview.dataRowCount }) }}</span><span>{{ t('questionnaireImport.columns', { count: preview.headers.length }) }}</span><span>{{ t('questionnaireImport.delimiter', { value: preview.delimiter }) }}</span></div>
      <label class="database-field"><span>{{ t('questionnaireImport.codeColumn') }}</span><select v-model="codeColumn" class="compact-select"><option v-for="header in preview.headers" :key="header" :value="header">{{ header }}</option></select></label>
      <div class="csv-preview" tabindex="0"><table><thead><tr><th v-for="header in preview.headers" :key="header">{{ header }}</th></tr></thead><tbody><tr v-for="(row, index) in preview.sampleRows" :key="index"><td v-for="(cell, cellIndex) in row" :key="cellIndex">{{ cell }}</td></tr></tbody></table></div>
      <div v-if="summary" class="csv-summary"><span><strong>{{ summary.matchedCount }}</strong> {{ t('questionnaireImport.matched') }}</span><span><strong>{{ summary.duplicateCount }}</strong> {{ t('questionnaireImport.duplicate') }}</span><span><strong>{{ summary.unknownCount }}</strong> {{ t('questionnaireImport.unknown') }}</span><span><strong>{{ summary.malformedCount }}</strong> {{ t('questionnaireImport.malformed') }}</span><span><strong>{{ summary.alreadyAwardedCount }}</strong> {{ t('questionnaireImport.rewarded') }}</span></div>
      <div class="button-row"><button class="button button--secondary" type="button" :disabled="busy" @click="reset">{{ t('questionnaireImport.startOver') }}</button><button v-if="!summary" class="button button--primary" type="button" :disabled="busy || !codeColumn" @click="analyze">{{ busy ? t('questionnaireImport.analyzing') : t('questionnaireImport.analyze') }}</button><button v-else class="button button--primary" type="button" :disabled="busy || !canSettle" @click="settle"><PhCheckCircle :size="18" />{{ busy ? t('questionnaireImport.queueing') : t('questionnaireImport.confirm') }}</button></div>
    </template>
  </aside>
</template>

<style scoped>
.questionnaire-import { width: min(100%, 680px); max-height: calc(100dvh - 1rem); overflow: auto; position: fixed; right: 0; bottom: 0; z-index: 60; display: grid; gap: 1rem; padding: 1rem; border: 1px solid var(--line-strong); border-radius: var(--radius-panel) var(--radius-panel) 0 0; background: var(--surface); box-shadow: var(--shadow-panel); }
.questionnaire-import > header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.questionnaire-import h3, .questionnaire-import header p { margin: 0; }
.questionnaire-import header p { margin-top: 0.3rem; color: var(--text-muted); font-size: 0.72rem; }
.csv-drop { display: grid; place-items: center; min-height: 280px; padding: 1.2rem; border: 1px dashed var(--line-strong); border-radius: var(--radius-panel); color: var(--accent); text-align: center; }
.csv-drop h4, .csv-drop p { margin: 0; }
.csv-drop h4 { margin-top: 0.7rem; color: var(--text); }
.csv-drop p { max-width: 34ch; margin: 0.4rem 0 1rem; color: var(--text-muted); font-size: 0.74rem; line-height: 1.45; }
.csv-meta, .csv-summary { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.csv-meta span, .csv-summary span { padding: 0.45rem 0.6rem; border-radius: 9px; color: var(--text-muted); background: var(--surface-raised); font-size: 0.68rem; }
.csv-summary strong { color: var(--accent); font-family: var(--font-mono); }
.csv-preview { max-height: 260px; overflow: auto; border: 1px solid var(--line); border-radius: var(--radius-control); }
.csv-preview table { border-collapse: collapse; font-size: 0.68rem; }
.csv-preview th, .csv-preview td { max-width: 200px; padding: 0.5rem; border-bottom: 1px solid var(--line); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.csv-preview th { position: sticky; top: 0; background: var(--surface-raised); }
.database-field { display: grid; gap: 0.4rem; }
.database-field > span { color: var(--text-muted); font-size: 0.75rem; font-weight: 700; }
.sr-only { width: 1px; height: 1px; overflow: hidden; position: absolute; clip: rect(0, 0, 0, 0); }

@media (min-width: 900px) { .questionnaire-import { top: 0; bottom: 0; border-radius: var(--radius-panel) 0 0 var(--radius-panel); } }
</style>
