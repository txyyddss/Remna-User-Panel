import { request } from './http'

export interface AbuseRecord { id: string; occurredAt: string; reason: string; measuredQPS: number; qpsLimit: number; action: string; expiresAt?: string }
export interface AbusePage { items: AbuseRecord[]; nextCursor: string }
export interface AbusePolicy { globalEnabled: boolean; globalLimit: number; warningValidityDays: number; warningCooldownMinutes: number; revision: number }
export interface AbuseRule { id: string; name: string; expression: string; qpsLimit: number; enabled: boolean; revision: number }
export interface AbuseNode { uuid: string; name: string; lastReportAt?: string; rotatedAt: string }
export interface AbusePunishment { action: string; enabled: boolean; incidentThreshold: number; durationMinutes: number; allNodes: boolean; revision: number }

const key = () => crypto.randomUUID()
export const abuseApi = {
  records: () => request<AbusePage>('/api/v1/me/abuse-records'),
  adminRecords: () => request<AbusePage>('/api/v1/admin/abuse/records'),
  policy: () => request<AbusePolicy>('/api/v1/admin/abuse/policy'),
  savePolicy: (body: AbusePolicy) => request<AbusePolicy>('/api/v1/admin/abuse/policy', { method: 'PUT', headers: { 'Idempotency-Key': key() }, body }),
  nodes: () => request<AbuseNode[]>('/api/v1/admin/abuse/nodes'),
  copyNodeKey: (id: string) => request<{ key: string }>(`/api/v1/admin/abuse/nodes/${encodeURIComponent(id)}/key`),
  rotateNodeKey: (id: string) => request<{ key: string }>(`/api/v1/admin/abuse/nodes/${encodeURIComponent(id)}/rotate`, { method: 'POST', headers: { 'Idempotency-Key': key() } }),
  rules: () => request<AbuseRule[]>('/api/v1/admin/abuse/rules'),
  saveRule: (body: AbuseRule) => request<AbuseRule>(body.id ? `/api/v1/admin/abuse/rules/${encodeURIComponent(body.id)}` : '/api/v1/admin/abuse/rules', { method: body.id ? 'PUT' : 'POST', headers: { 'Idempotency-Key': key() }, body }),
  deleteRule: (id: string, revision: number) => request<void>(`/api/v1/admin/abuse/rules/${encodeURIComponent(id)}?revision=${encodeURIComponent(String(revision))}`, { method: 'DELETE', headers: { 'Idempotency-Key': key() } }),
  punishments: () => request<AbusePunishment[]>('/api/v1/admin/abuse/punishments'),
  savePunishment: (body: AbusePunishment) => request<AbusePunishment>(`/api/v1/admin/abuse/punishments/${encodeURIComponent(body.action)}`, { method: 'PUT', headers: { 'Idempotency-Key': key() }, body }),
  whitelist: () => request<string[]>('/api/v1/admin/abuse/whitelist'),
  setWhitelist: (id: string, enabled: boolean) => request<void>(`/api/v1/admin/abuse/whitelist/${encodeURIComponent(id)}`, { method: 'PUT', headers: { 'Idempotency-Key': key() }, body: { enabled } }),
  statistics: () => request<Record<string, number>>('/api/v1/admin/abuse/statistics'),
}
