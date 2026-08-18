import { onScopeDispose, shallowRef, watch } from 'vue'

import { api } from '@/api/client'
import type { OperationReceipt, PaymentOperation } from '@/api/types'
import { createUuid } from '@/utils/browserCompatibility'
import { operationIsActive, useOperationReceipt } from './useOperationReceipt'

export type PaymentCommandKind = 'create' | 'cancel'

export interface PaymentCommand {
  kind: PaymentCommandKind
  paymentOrderId: string
}

interface PaymentOrderOperationOptions {
  onTerminal: (command: PaymentCommand, receipt: OperationReceipt) => void | Promise<void>
}

export function usePaymentOrderOperations(options: PaymentOrderOperationOptions) {
  const command = shallowRef<PaymentCommand | null>(null)
  const tracker = useOperationReceipt()
  let createKey: string | undefined
  let cancelKey: string | undefined

  function accept(kind: PaymentCommandKind, result: PaymentOperation): void {
    command.value = { kind, paymentOrderId: result.paymentOrderId }
    tracker.track(result.operation)
  }

  async function queueCreate(methodId: string, txbMinor: string): Promise<void> {
    createKey ??= createUuid()
    const result = await api.createPaymentOrder(methodId, txbMinor, createKey)
    createKey = undefined
    accept('create', result)
  }

  async function queueCancellation(paymentOrderId: string): Promise<void> {
    cancelKey ??= createUuid()
    const result = await api.cancelPaymentOrder(paymentOrderId, cancelKey)
    cancelKey = undefined
    accept('cancel', result)
  }

  watch(tracker.receipt, (receipt) => {
    const current = command.value
    if (!receipt || !current || operationIsActive(receipt.status)) return
    command.value = null
    void options.onTerminal(current, receipt)
  })

  function reset(): void {
    command.value = null
    tracker.reset()
  }

  onScopeDispose(reset)

  return {
    receipt: tracker.receipt,
    error: tracker.error,
    queueCreate,
    queueCancellation,
    retry: tracker.refresh,
    reset,
  }
}
