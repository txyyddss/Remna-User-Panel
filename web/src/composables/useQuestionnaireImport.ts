import { onScopeDispose, readonly, shallowRef } from 'vue'

import type { QuestionnaireImportPreview, QuestionnaireImportSummary, QuestionnaireSettlementReport } from '@/api/features'
import { featuresApi } from '@/api/features'

export function useQuestionnaireImport(questionnaireId: () => string) {
  const preview = shallowRef<QuestionnaireImportPreview | null>(null)
  const summary = shallowRef<QuestionnaireImportSummary | null>(null)
  const codeColumn = shallowRef('')
  const busy = shallowRef(false)
  const settlementImportId = shallowRef<string | null>(null)
  const report = shallowRef<QuestionnaireSettlementReport | null>(null)
  const error = shallowRef<string | null>(null)
  let pollTimer: ReturnType<typeof setTimeout> | undefined

  function stopPolling(): void {
    if (pollTimer !== undefined) clearTimeout(pollTimer)
    pollTimer = undefined
  }

  async function upload(file: File): Promise<void> {
    if (file.size > 5 * 1024 * 1024) {
      error.value = 'CSV files are limited to 5 MiB.'
      return
    }
    busy.value = true
    error.value = null
    summary.value = null
    settlementImportId.value = null
    report.value = null
    stopPolling()
    try {
      preview.value = await featuresApi.previewQuestionnaireCsv(questionnaireId(), file)
      codeColumn.value = preview.value.headers[0] ?? ''
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'The CSV could not be uploaded.'
    } finally {
      busy.value = false
    }
  }

  async function analyze(): Promise<void> {
    if (!preview.value || !codeColumn.value) return
    busy.value = true
    error.value = null
    try {
      summary.value = await featuresApi.analyzeQuestionnaireCsv(questionnaireId(), preview.value.id, codeColumn.value)
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'The CSV could not be analyzed.'
    } finally {
      busy.value = false
    }
  }

  async function settle(): Promise<void> {
    if (!preview.value || !summary.value || !codeColumn.value) return
    busy.value = true
    error.value = null
    try {
      const response = await featuresApi.settleQuestionnaireCsv(questionnaireId(), preview.value.id)
      settlementImportId.value = response.id
      preview.value = response
      schedulePoll()
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'Settlement could not be queued.'
    } finally {
      busy.value = false
    }
  }

  function schedulePoll(): void {
    stopPolling()
    pollTimer = setTimeout(() => void pollSettlement(), 1500)
  }

  async function pollSettlement(): Promise<void> {
    if (!settlementImportId.value) return
    try {
      const state = await featuresApi.getQuestionnaireImportState(questionnaireId(), settlementImportId.value)
      preview.value = state.preview
      report.value = state.report ?? null
      if (state.preview.status === 'settled' || state.preview.status === 'failed') {
        stopPolling()
        if (state.preview.status === 'failed') error.value = 'Settlement failed. The import is preserved for review.'
        return
      }
    } catch {
      // Keep the durable import visible and retry a transient status failure.
    }
    schedulePoll()
  }

  function reset(): void {
    preview.value = null
    summary.value = null
    codeColumn.value = ''
    settlementImportId.value = null
    report.value = null
    error.value = null
    stopPolling()
  }

  onScopeDispose(stopPolling)

  return {
    preview: readonly(preview),
    summary: readonly(summary),
    codeColumn,
    busy: readonly(busy),
    settlementImportId: readonly(settlementImportId),
    report: readonly(report),
    error: readonly(error),
    upload,
    analyze,
    settle,
    reset,
  }
}
