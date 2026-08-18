import { computed, onScopeDispose, readonly, shallowRef } from 'vue'

import type { QuestionnaireImportPreview, QuestionnaireImportSummary, QuestionnaireSettlementReport } from '@/api/features'
import { featuresApi } from '@/api/features'
import { localizedError, t } from '@/i18n'
import { useDurableCommand } from './useDurableCommand'

export function useQuestionnaireImport(questionnaireId: () => string) {
  const preview = shallowRef<QuestionnaireImportPreview | null>(null)
  const summary = shallowRef<QuestionnaireImportSummary | null>(null)
  const codeColumn = shallowRef('')
  const working = shallowRef(false)
  const settlementImportId = shallowRef<string | null>(null)
  const report = shallowRef<QuestionnaireSettlementReport | null>(null)
  const mutationError = shallowRef<string | null>(null)
  const settlement = useDurableCommand({
    errorKey: 'errors.settlementQueue',
    onTerminal: () => pollSettlement(),
  })
  const busy = computed(() => working.value || settlement.submitting.value)
  const error = computed(() => mutationError.value ?? settlement.error.value)
  let pollTimer: ReturnType<typeof setTimeout> | undefined

  function stopPolling(): void {
    if (pollTimer !== undefined) clearTimeout(pollTimer)
    pollTimer = undefined
  }

  async function upload(file: File): Promise<void> {
    if (file.size > 5 * 1024 * 1024) {
      mutationError.value = t('errors.csvLimit')
      return
    }
    working.value = true
    mutationError.value = null
    summary.value = null
    settlementImportId.value = null
    report.value = null
    stopPolling()
    try {
      preview.value = await featuresApi.previewQuestionnaireCsv(questionnaireId(), file)
      codeColumn.value = preview.value.headers[0] ?? ''
    } catch (caught) {
      mutationError.value = localizedError(caught, 'errors.csvUpload')
    } finally {
      working.value = false
    }
  }

  async function analyze(): Promise<void> {
    if (!preview.value || !codeColumn.value) return
    working.value = true
    mutationError.value = null
    try {
      summary.value = await featuresApi.analyzeQuestionnaireCsv(questionnaireId(), preview.value.id, codeColumn.value)
    } catch (caught) {
      mutationError.value = localizedError(caught, 'errors.csvAnalyze')
    } finally {
      working.value = false
    }
  }

  async function settle(): Promise<void> {
    if (!preview.value || !summary.value || !codeColumn.value) return
    const importId = preview.value.id
    const ownerId = questionnaireId()
    mutationError.value = null
    await settlement.execute(importId, `${ownerId}:${importId}`, async (key) => {
      const response = await featuresApi.settleQuestionnaireCsv(ownerId, importId, key)
      settlementImportId.value = response.import.id
      preview.value = response.import
      schedulePoll()
      return response.operation
    })
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
        if (state.preview.status === 'failed') mutationError.value = t('errors.settlementFailed')
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
    mutationError.value = null
    working.value = false
    stopPolling()
    settlement.reset()
  }

  onScopeDispose(stopPolling)

  return {
    preview: readonly(preview),
    summary: readonly(summary),
    codeColumn,
    busy: readonly(busy),
    settlementImportId: readonly(settlementImportId),
    operationReceipt: settlement.receipt,
    operationChecking: settlement.checking,
    report: readonly(report),
    error: readonly(error),
    upload,
    analyze,
    settle,
    refreshOperation: settlement.refresh,
    reset,
  }
}
