import { computed, onScopeDispose, readonly, shallowRef } from 'vue'

import type { QuestionnaireImportPreview, QuestionnaireImportSummary, QuestionnaireSettlementReport } from '@/api/features'
import { featuresApi } from '@/api/features'
import { localizedError, t } from '@/i18n'
import { useDurableCommand } from './useDurableCommand'
import { createLatestRequest } from '@/utils/latestRequest'

export function useQuestionnaireImport(questionnaireId: () => string) {
  const preview = shallowRef<QuestionnaireImportPreview | null>(null)
  const summary = shallowRef<QuestionnaireImportSummary | null>(null)
  const codeColumn = shallowRef('')
  const working = shallowRef(false)
  const settlementImportId = shallowRef<string | null>(null)
  const report = shallowRef<QuestionnaireSettlementReport | null>(null)
  const mutationError = shallowRef<string | null>(null)
  const latestPoll = createLatestRequest()
  const latestMutation = createLatestRequest()
  const settlement = useDurableCommand({
    errorKey: 'errors.settlementQueue',
    onTerminal: () => { void pollSettlement() },
  })
  const busy = computed(() => working.value || settlement.busy.value)
  const error = computed(() => mutationError.value ?? settlement.error.value)
  let pollTimer: ReturnType<typeof setTimeout> | undefined

  function stopPolling(): void {
    latestPoll.invalidate()
    clearPollTimer()
  }

  function clearPollTimer(): void {
    if (pollTimer !== undefined) clearTimeout(pollTimer)
    pollTimer = undefined
  }

  async function upload(file: File): Promise<void> {
    const token = latestMutation.begin()
    stopPolling()
    if (file.size > 5 * 1024 * 1024) {
      working.value = false
      mutationError.value = t('errors.csvLimit')
      return
    }
    working.value = true
    mutationError.value = null
    summary.value = null
    settlementImportId.value = null
    report.value = null
    try {
      const next = await featuresApi.previewQuestionnaireCsv(questionnaireId(), file)
      if (!latestMutation.isCurrent(token)) return
      preview.value = next
      codeColumn.value = next.headers[0] ?? ''
    } catch (caught) {
      if (!latestMutation.isCurrent(token)) return
      mutationError.value = localizedError(caught, 'errors.csvUpload')
    } finally {
      if (latestMutation.isCurrent(token)) working.value = false
    }
  }

  async function analyze(): Promise<void> {
    if (!preview.value || !codeColumn.value) return
    const token = latestMutation.begin()
    const ownerID = questionnaireId()
    const importID = preview.value.id
    const column = codeColumn.value
    working.value = true
    mutationError.value = null
    try {
      const next = await featuresApi.analyzeQuestionnaireCsv(ownerID, importID, column)
      if (!latestMutation.isCurrent(token)) return
      summary.value = next
    } catch (caught) {
      if (!latestMutation.isCurrent(token)) return
      mutationError.value = localizedError(caught, 'errors.csvAnalyze')
    } finally {
      if (latestMutation.isCurrent(token)) working.value = false
    }
  }

  async function settle(): Promise<void> {
    if (!preview.value || !summary.value || !codeColumn.value || settlement.blocksMutations.value) return
    const token = latestMutation.begin()
    const importId = preview.value.id
    const ownerId = questionnaireId()
    mutationError.value = null
    await settlement.execute(importId, `${ownerId}:${importId}`, async (key) => {
      const response = await featuresApi.settleQuestionnaireCsv(ownerId, importId, key)
      if (!latestMutation.isCurrent(token)) return response.operation
      settlementImportId.value = response.import.id
      preview.value = response.import
      schedulePoll(latestPoll.begin())
      return response.operation
    })
  }

  function schedulePoll(token: number): void {
    clearPollTimer()
    pollTimer = setTimeout(() => void pollSettlement(token), 1500)
  }

  async function pollSettlement(expectedToken?: number): Promise<void> {
    const token = expectedToken ?? latestPoll.begin()
    if (!latestPoll.isCurrent(token) || !settlementImportId.value) return
    try {
      const state = await featuresApi.getQuestionnaireImportState(questionnaireId(), settlementImportId.value)
      if (!latestPoll.isCurrent(token)) return
      preview.value = state.preview
      report.value = state.report ?? null
      if (state.preview.status === 'settled' || state.preview.status === 'failed') {
        stopPolling()
        if (state.preview.status === 'failed') mutationError.value = t('errors.settlementFailed')
        return
      }
    } catch {
      if (!latestPoll.isCurrent(token)) return
      // Keep the durable import visible and retry a transient status failure.
    }
    schedulePoll(token)
  }

  function reset(): void {
    latestMutation.invalidate()
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

  onScopeDispose(() => {
    clearPollTimer()
    latestPoll.dispose()
    latestMutation.dispose()
  })

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
