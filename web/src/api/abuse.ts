import { request } from './http'
import { createUuid } from '@/utils/browserCompatibility'

export interface AbuseRecord { id: string; occurredAt: string; reason: string; measuredQPS: number; qpsLimit: number; action: string; expiresAt?: string }
export interface AbusePage { items: AbuseRecord[]; nextCursor: string }
export interface AbusePolicy { globalEnabled: boolean; globalLimit: number; streakSeconds: number; warningValidityDays: number; warningCooldownMinutes: number; revision: number }
export interface AbuseRule { id: string; name: string; expression: string; qpsLimit: number; enabled: boolean; revision: number }
export interface AbuseNode { uuid: string; name: string; lastReportAt?: string; rotatedAt: string }
export interface AbusePunishment { action: string; enabled: boolean; incidentThreshold: number; durationMinutes: number; allNodes: boolean; revision: number }

const key = () => createUuid()

function saveRule(value: AbuseRule): Promise<AbuseRule> {
  const { id, ...body } = value
  const path = id ? `/api/v1/admin/abuse/rules/${encodeURIComponent(id)}` : '/api/v1/admin/abuse/rules'
  return request<AbuseRule>(path, { method: id ? 'PUT' : 'POST', headers: { 'Idempotency-Key': key() }, body })
}

function savePunishment(value: AbusePunishment): Promise<AbusePunishment> {
  const { action, ...body } = value
  const path = `/api/v1/admin/abuse/punishments/${encodeURIComponent(action)}`
  return request<AbusePunishment>(path, { method: 'PUT', headers: { 'Idempotency-Key': key() }, body })
}

export const abuseApi = {
  records: (cursor = '') => request<AbusePage>(cursor ? `/api/v1/me/abuse-records?cursor=${encodeURIComponent(cursor)}` : '/api/v1/me/abuse-records'),
  adminRecords: (cursor = '') => request<AbusePage>(cursor ? `/api/v1/admin/abuse/records?cursor=${encodeURIComponent(cursor)}` : '/api/v1/admin/abuse/records'),
  policy: () => request<AbusePolicy>('/api/v1/admin/abuse/policy'),
  savePolicy: (body: AbusePolicy) => request<AbusePolicy>('/api/v1/admin/abuse/policy', { method: 'PUT', headers: { 'Idempotency-Key': key() }, body }),
  nodes: () => request<AbuseNode[]>('/api/v1/admin/abuse/nodes'),
  copyNodeKey: (id: string) => request<{ key: string }>(`/api/v1/admin/abuse/nodes/${encodeURIComponent(id)}/key`),
  rotateNodeKey: (id: string) => request<{ key: string }>(`/api/v1/admin/abuse/nodes/${encodeURIComponent(id)}/rotate`, { method: 'POST', headers: { 'Idempotency-Key': key() } }),
  rules: () => request<AbuseRule[]>('/api/v1/admin/abuse/rules'),
  saveRule,
  deleteRule: (id: string, revision: number) => request<void>(`/api/v1/admin/abuse/rules/${encodeURIComponent(id)}?revision=${encodeURIComponent(String(revision))}`, { method: 'DELETE', headers: { 'Idempotency-Key': key() } }),
  punishments: () => request<AbusePunishment[]>('/api/v1/admin/abuse/punishments'),
  savePunishment,
  whitelist: () => request<string[]>('/api/v1/admin/abuse/whitelist'),
  setWhitelist: (id: string, enabled: boolean) => request<void>(`/api/v1/admin/abuse/whitelist/${encodeURIComponent(id)}`, { method: 'PUT', headers: { 'Idempotency-Key': key() }, body: { enabled } }),
  statistics: () => request<Record<string, number>>('/api/v1/admin/abuse/statistics'),
}
