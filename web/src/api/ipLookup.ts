import type { components } from './generated'
import { request } from './http'

export type IPLookupState = components['schemas']['IPLookupState']
export type IPLookupQuote = components['schemas']['IPLookupQuote']
export type IPLookupCheck = components['schemas']['IPLookupCheck']
export type IPLookupReport = components['schemas']['IPLookupReport']
export type IPLookupProvider = components['schemas']['IPLookupProviderResult']
export type IPLookupAdminSettings = components['schemas']['IPLookupAdminSettings']

export const ipLookupApi = {
  state: () => request<IPLookupState>('/api/v1/ip-lookup', { cache: 'no-store' }),
  quote: (ip: string, refresh: boolean) => request<IPLookupQuote>('/api/v1/ip-lookup/quote', { method: 'POST', body: { ip, refresh }, cache: 'no-store' }),
  submit: (quote: IPLookupQuote, key: string) => request<components['schemas']['OperationReceipt']>('/api/v1/ip-lookup/checks', { method: 'POST', body: quote, headers: { 'Idempotency-Key': key } }),
  check: (id: string) => request<IPLookupCheck>(`/api/v1/ip-lookup/checks/${encodeURIComponent(id)}`, { cache: 'no-store' }),
  settings: () => request<IPLookupAdminSettings>('/api/v1/admin/ip-lookup', { cache: 'no-store' }),
  saveSettings: (value: IPLookupAdminSettings) => request<IPLookupAdminSettings>('/api/v1/admin/ip-lookup', { method: 'PUT', body: value }),
}
