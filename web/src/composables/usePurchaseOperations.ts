import { computed, readonly, shallowRef } from 'vue'

import { ApiError } from '@/api/client'
import { memberOperationsApi } from '@/api/memberOperations'
import type { MemberRefundQuote } from '@/api/types'
import { localizedError, t } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { notifyHaptic } from '@/utils/telegram'
import { operationIsActive, useOperationReceipt } from './useOperationReceipt'

export type PurchaseOperationKind = 'refund'

export function usePurchaseOperations(purchaseId: () => string) {
  const refundQuote = shallowRef<MemberRefundQuote | null>(null)
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

  async function loadQuote(): Promise<void> {
    quoteLoading.value = true
    mutationError.value = null
    refundQuote.value = null
    try {
      refundQuote.value = await memberOperationsApi.getPurchaseRefundQuote(purchaseId())
    } catch (caught) {
      mutationError.value = localizedError(caught, 'purchaseOperations.errors.quoteUnavailable')
    } finally {
      quoteLoading.value = false
    }
  }

  async function refreshAfterConflict(): Promise<void> {
    refundQuote.value = null
    try {
      refundQuote.value = await memberOperationsApi.getPurchaseRefundQuote(purchaseId())
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
      const receipt = await memberOperationsApi.refundPurchase(purchaseId(), key)
      keys.delete(actionId)
      activeKind.value = kind
      operation.track(receipt)
      notifyHaptic('success')
      return true
    } catch (caught) {
      if (caught instanceof ApiError && (caught.status === 409 || caught.status === 422)) {
        keys.delete(actionId)
        await refreshAfterConflict()
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
    refundQuote.value = null
    mutationError.value = null
    activeKind.value = null
    keys.clear()
    operation.reset()
  }

  return {
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
    start,
    refresh: operation.refresh,
    dismissOperation,
    reset,
  }
}
