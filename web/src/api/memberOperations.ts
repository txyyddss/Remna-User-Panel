import type {
  ConnectionScan,
  MemberRefundQuote,
  OperationReceipt,
  TrafficResetQuote,
} from './types'
import { request } from './http'

export const memberOperationsApi = {
  requestConnections: (idempotencyKey: string) => request<ConnectionScan>('/api/v1/subscription/connections', {
    method: 'POST',
    headers: { 'Idempotency-Key': idempotencyKey },
  }),
  pollConnections: (scanId: string) => request<ConnectionScan>(`/api/v1/subscription/connections/${encodeURIComponent(scanId)}`),
  dropConnection: (handle: string, idempotencyKey: string) => request<OperationReceipt>('/api/v1/subscription/connections/drop', {
    method: 'POST',
    headers: { 'Idempotency-Key': idempotencyKey },
    body: { handle },
  }),
  getTrafficResetQuote: (purchaseId: string) => request<TrafficResetQuote>(`/api/v1/purchases/${encodeURIComponent(purchaseId)}/traffic-reset`),
  resetPurchaseTraffic: (purchaseId: string, idempotencyKey: string) => request<OperationReceipt>(`/api/v1/purchases/${encodeURIComponent(purchaseId)}/traffic-reset`, {
    method: 'POST',
    headers: { 'Idempotency-Key': idempotencyKey },
  }),
  getPurchaseRefundQuote: (purchaseId: string) => request<MemberRefundQuote>(`/api/v1/purchases/${encodeURIComponent(purchaseId)}/refund`),
  refundPurchase: (purchaseId: string, idempotencyKey: string) => request<OperationReceipt>(`/api/v1/purchases/${encodeURIComponent(purchaseId)}/refund`, {
    method: 'POST',
    headers: { 'Idempotency-Key': idempotencyKey },
  }),
  getOperation: (operationId: string) => request<OperationReceipt>(`/api/v1/operations/${encodeURIComponent(operationId)}`),
}
