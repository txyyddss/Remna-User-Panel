import { computed, readonly, shallowRef } from 'vue'

import { ApiError } from '@/api/client'
import { memberOperationsApi } from '@/api/memberOperations'
import type { MemberRefundQuote, TrafficResetAutomation, TrafficResetQuote } from '@/api/types'
import { localizedError, t } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { notifyHaptic, selectionHaptic } from '@/utils/telegram'
import { operationIsActive, useOperationReceipt } from './useOperationReceipt'

export type PurchaseOperationKind = 'reset' | 'refund'

export function usePurchaseOperations(purchaseId: () => string) {
  const resetQuote = shallowRef<TrafficResetQuote | null>(null)
  const refundQuote = shallowRef<MemberRefundQuote | null>(null)
  const resetAutomation = shallowRef<TrafficResetAutomation | null>(null)
  const resetAutomationLoading = shallowRef(false)
  const resetAutomationSaving = shallowRef(false)
  const resetAutomationError = shallowRef<string | null>(null)
  const refundEligibilityLoading = shallowRef(true)
  const quoteLoading = shallowRef(false)
  const mutating = shallowRef(false)
  const mutationError = shallowRef<string | null>(null)
  const activeKind = shallowRef<PurchaseOperationKind | null>(null)
  const keys = new Map<string, string>()
  const operation = useOperationReceipt()

  const error = computed(() => mutationError.value ?? operation.error.value)
  const blocksMutations = computed(() => {
    const status = operation.receipt.value?.status
    return operationIsActive(status) || status === 'pending_review' || status === 'partial'
  })

  async function loadRefundEligibility(): Promise<void> {
    refundEligibilityLoading.value = true
    try {
      refundQuote.value = await memberOperationsApi.getPurchaseRefundQuote(purchaseId())
    } catch {
      refundQuote.value = null
    } finally {
      refundEligibilityLoading.value = false
    }
  }

  async function loadQuote(kind: PurchaseOperationKind): Promise<void> {
    quoteLoading.value = true
    mutationError.value = null
    if (kind === 'reset') {
      resetQuote.value = null
      resetAutomationLoading.value = true
      resetAutomationError.value = null
    } else refundQuote.value = null
    try {
      if (kind === 'reset') {
        const [quoteResult, automationResult] = await Promise.allSettled([
          memberOperationsApi.getTrafficResetQuote(purchaseId()),
          memberOperationsApi.getTrafficResetAutomation(),
        ])
        if (automationResult.status === 'fulfilled') resetAutomation.value = automationResult.value
        else resetAutomationError.value = localizedError(automationResult.reason, 'purchaseOperations.automation.loadFailed')
        if (quoteResult.status === 'rejected') throw quoteResult.reason
        resetQuote.value = quoteResult.value
      } else refundQuote.value = await memberOperationsApi.getPurchaseRefundQuote(purchaseId())
    } catch (caught) {
      mutationError.value = localizedError(caught, 'purchaseOperations.errors.quoteUnavailable')
    } finally {
      quoteLoading.value = false
      resetAutomationLoading.value = false
    }
  }

  async function setResetAutomation(enabled: boolean): Promise<void> {
    if (resetAutomationSaving.value) return
    resetAutomationSaving.value = true
    resetAutomationError.value = null
    selectionHaptic()
    try {
      resetAutomation.value = await memberOperationsApi.updateTrafficResetAutomation(enabled)
    } catch (caught) {
      resetAutomationError.value = localizedError(caught, 'purchaseOperations.automation.saveFailed')
      notifyHaptic('error')
    } finally {
      resetAutomationSaving.value = false
    }
  }

  async function refreshAfterConflict(kind: PurchaseOperationKind): Promise<void> {
    if (kind === 'reset') resetQuote.value = null
    else refundQuote.value = null
    try {
      if (kind === 'reset') resetQuote.value = await memberOperationsApi.getTrafficResetQuote(purchaseId())
      else refundQuote.value = await memberOperationsApi.getPurchaseRefundQuote(purchaseId())
    } catch {
      // Preserve the mutation error; the member can explicitly request a new quote.
    }
  }

  async function start(kind: PurchaseOperationKind): Promise<boolean> {
    if (mutating.value || blocksMutations.value) return false
    mutating.value = true
    mutationError.value = null
    const actionId = `${kind}:${purchaseId()}`
    try {
      const key = keys.get(actionId) ?? createUuid()
      keys.set(actionId, key)
      const receipt = kind === 'reset'
        ? await memberOperationsApi.resetPurchaseTraffic(purchaseId(), key)
        : await memberOperationsApi.refundPurchase(purchaseId(), key)
      keys.delete(actionId)
      activeKind.value = kind
      operation.track(receipt)
      notifyHaptic('success')
      return true
    } catch (caught) {
      if (caught instanceof ApiError && (caught.status === 409 || caught.status === 422)) {
        keys.delete(actionId)
        await refreshAfterConflict(kind)
      }
      mutationError.value = caught instanceof ApiError && ['OPERATION_CONFLICT', 'PURCHASE_OPERATION_INELIGIBLE'].includes(caught.code)
        ? t('purchaseOperations.quoteChanged')
        : localizedError(caught, 'purchaseOperations.errors.operationFailed')
      notifyHaptic('error')
      return false
    } finally {
      mutating.value = false
    }
  }

  function dismissOperation(): void {
    if (blocksMutations.value) return
    activeKind.value = null
    mutationError.value = null
    operation.reset()
  }

  function reset(): void {
    resetQuote.value = null
    refundQuote.value = null
    resetAutomation.value = null
    resetAutomationError.value = null
    mutationError.value = null
    activeKind.value = null
    keys.clear()
    operation.reset()
  }

  return {
    resetQuote: readonly(resetQuote),
    resetAutomation: readonly(resetAutomation),
    resetAutomationLoading: readonly(resetAutomationLoading),
    resetAutomationSaving: readonly(resetAutomationSaving),
    resetAutomationError: readonly(resetAutomationError),
    refundQuote: readonly(refundQuote),
    refundEligibilityLoading: readonly(refundEligibilityLoading),
    quoteLoading: readonly(quoteLoading),
    mutating: readonly(mutating),
    activeKind: readonly(activeKind),
    receipt: operation.receipt,
    polling: operation.polling,
    checking: operation.checking,
    terminal: operation.terminal,
    error,
    blocksMutations,
    loadRefundEligibility,
    loadQuote,
    setResetAutomation,
    start,
    refresh: operation.refresh,
    dismissOperation,
    reset,
  }
}
