import { request } from './http'
import type {
  NodeCompensationConfig,
  NodeCompensationConfigWrite,
  NodeCompensationEvent,
  NodeCompensationEventPage,
  NodeCompensationReview,
  NodeCompensationStatus,
} from './contracts/compensation'

export const compensationApi = {
  config: () => request<NodeCompensationConfig>('/api/v1/admin/node-compensation/config'),
  saveConfig: (body: NodeCompensationConfigWrite) => request<NodeCompensationConfig>('/api/v1/admin/node-compensation/config', {
    method: 'PUT', body,
  }),
  events: (status: NodeCompensationStatus | '', cursor?: string) => request<NodeCompensationEventPage>(
    '/api/v1/admin/node-compensation/events', { query: { status: status || undefined, cursor, limit: 25 } },
  ),
  approve: (id: string, body: NodeCompensationReview, key: string) => request<NodeCompensationEvent>(
    `/api/v1/admin/node-compensation/events/${encodeURIComponent(id)}/approve`, {
      method: 'POST', body, headers: { 'Idempotency-Key': key },
    },
  ),
  dismiss: (id: string, body: NodeCompensationReview, key: string) => request<NodeCompensationEvent>(
    `/api/v1/admin/node-compensation/events/${encodeURIComponent(id)}/dismiss`, {
      method: 'POST', body: { revision: body.revision, reason: body.reason }, headers: { 'Idempotency-Key': key },
    },
  ),
}
