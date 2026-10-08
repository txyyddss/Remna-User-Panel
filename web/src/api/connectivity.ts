import type { components } from './generated'
import { request } from './http'

export type ConnectivityConfig = components['schemas']['ConnectivityConfig']
export type ConnectivityUser = components['schemas']['ConnectivityUser']
export type ConnectivityAttempt = components['schemas']['ConnectivityAttempt']
export type ConnectivityRun = components['schemas']['ConnectivityRun']
export type ConnectivityHostResult = components['schemas']['ConnectivityHostResult']
export type ConnectivitySnapshot = components['schemas']['ConnectivitySnapshot']
export type ConnectivityHistory = components['schemas']['ConnectivityHistory']

const base = '/api/v1/admin/host-connectivity'
export const connectivityApi = {
  snapshot: (signal?: AbortSignal) => request<ConnectivitySnapshot>(base, { cache: 'no-store', signal }),
  resolve: (username: string, signal?: AbortSignal) => request<ConnectivityUser>(`${base}/test-user/resolve`, { method: 'POST', body: { username }, signal }),
  check: (signal?: AbortSignal) => request<ConnectivityRun>(`${base}/checks`, { method: 'POST', signal }),
  history: (hostUuid?: string, cursor?: string, signal?: AbortSignal) => request<ConnectivityHistory>(`${base}/history`, { cache: 'no-store', signal, query: { hostUuid: hostUuid || undefined, cursor, limit: 50 } }),
  save: (config: ConnectivityConfig) => request<void>('/api/v1/admin/settings/connectivity.config', { method: 'PUT', body: { value: JSON.stringify(config) } }),
}
