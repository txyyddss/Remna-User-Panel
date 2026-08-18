import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

const { previewQuestionnaireCsv, analyzeQuestionnaireCsv, settleQuestionnaireCsv, getQuestionnaireImportState, getOperation } = vi.hoisted(() => ({
  previewQuestionnaireCsv: vi.fn(),
  analyzeQuestionnaireCsv: vi.fn(),
  settleQuestionnaireCsv: vi.fn(),
  getQuestionnaireImportState: vi.fn(),
  getOperation: vi.fn(),
}))

vi.mock('@/api/features', () => ({
  featuresApi: { previewQuestionnaireCsv, analyzeQuestionnaireCsv, settleQuestionnaireCsv, getQuestionnaireImportState },
}))
vi.mock('@/api/memberOperations', () => ({ memberOperationsApi: { getOperation } }))

import { useQuestionnaireImport } from './useQuestionnaireImport'

const preview = {
  id: 'import-1',
  questionnaireId: 'questionnaire-1',
  status: 'preview' as const,
  headers: ['Validation code', 'Answer'],
  sampleRows: [['ABC', 'Yes']],
  delimiter: 'comma' as const,
  dataRowCount: 1,
  malformedRowCount: 0,
  createdAt: '2026-08-08T00:00:00Z',
  updatedAt: '2026-08-08T00:00:00Z',
}

describe('questionnaire import workflow', () => {
  afterEach(() => {
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  it('keeps analysis read-only, then polls the durable settlement to completion', async () => {
    vi.useFakeTimers()
    previewQuestionnaireCsv.mockResolvedValue(preview)
    analyzeQuestionnaireCsv.mockResolvedValue({ importId: 'import-1', questionnaireId: 'questionnaire-1', codeColumn: 'Validation code', matchedCount: 1, duplicateCount: 0, unknownCount: 0, malformedCount: 0, alreadyAwardedCount: 0 })
    settleQuestionnaireCsv.mockResolvedValue({
      operation: {
        id: 'operation-1',
        kind: 'questionnaire_settlement',
        status: 'queued',
        errorCode: null,
        createdAt: '2026-08-08T00:00:00Z',
        updatedAt: '2026-08-08T00:00:00Z',
        completedAt: '2026-08-08T00:00:01Z',
      },
      import: { ...preview, status: 'processing' },
    })
    getOperation.mockResolvedValue({
      id: 'operation-1',
      kind: 'questionnaire_settlement',
      status: 'processing',
      errorCode: null,
      createdAt: '2026-08-08T00:00:00Z',
      updatedAt: '2026-08-08T00:00:00Z',
      completedAt: null,
    })
    getQuestionnaireImportState
      .mockResolvedValueOnce({ preview: { ...preview, status: 'processing' } })
      .mockResolvedValueOnce({
        preview: { ...preview, status: 'settled' },
        report: { importId: 'import-1', questionnaireId: 'questionnaire-1', matchedCount: 1, duplicateCount: 0, unknownCount: 0, malformedCount: 0, alreadyAwardedCount: 0, rewardedCount: 1, rewardTxbMinor: '500', settledAt: '2026-08-08T00:00:00Z', replayed: false },
      })
    const scope = effectScope()
    const state = scope.run(() => useQuestionnaireImport(() => 'questionnaire-1'))!

    await state.upload(new File(['Validation code,Answer\nABC,Yes'], 'responses.csv', { type: 'text/csv' }))
    await state.analyze()
    expect(settleQuestionnaireCsv).not.toHaveBeenCalled()
    expect(state.summary.value?.matchedCount).toBe(1)

    await state.settle()
    await vi.advanceTimersByTimeAsync(1500)
    expect(state.preview.value?.status).toBe('processing')
    await vi.advanceTimersByTimeAsync(1500)
    expect(state.preview.value?.status).toBe('settled')
    expect(state.report.value?.rewardedCount).toBe(1)
    scope.stop()
  })
})
