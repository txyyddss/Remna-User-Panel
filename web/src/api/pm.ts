import type { components } from './generated'
import { request } from './http'
import type { OperationReceipt, PMDelivery } from './types'

export const pmApi = {
  moderate: (id: string, body: components['schemas']['PMModerationUpdate'], key: string) =>
    request<OperationReceipt>(`/api/v1/admin/pm/${encodeURIComponent(id)}`, { method: 'PUT', body, headers: { 'Idempotency-Key': key } }),
  repair: (id: string, body: components['schemas']['PMTopicRepair'], key: string) =>
    request<OperationReceipt>(`/api/v1/admin/pm/${encodeURIComponent(id)}/topic`, { method: 'POST', body, headers: { 'Idempotency-Key': key } }),
  deliveries: (id: string) => request<{ items: PMDelivery[] }>(`/api/v1/admin/pm/${encodeURIComponent(id)}/deliveries`),
}
