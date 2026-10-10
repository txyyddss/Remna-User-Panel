import type { components } from './generated'
import { request } from './http'

export type IPLookupState = components['schemas']['IPLookupState']
export type IPLookupQuote = components['schemas']['IPLookupQuoteResponse']
type IPLookupSubmission = components['schemas']['IPLookupQuote']
export type IPLookupCheck = components['schemas']['IPLookupCheck']
export type IPLookupReport = components['schemas']['IPLookupReport']
export type IPLookupAdminSettings = components['schemas']['IPLookupAdminSettings']

export const ipLookupApi = {
  state: () => request<IPLookupState>('/api/v1/ip-lookup', { cache: 'no-store' }),
  quote: (ip: string, refresh: boolean) => request<IPLookupQuote>('/api/v1/ip-lookup/quote', { method: 'POST', body: { ip, refresh }, cache: 'no-store' }),
  submit: (quote: IPLookupQuote, key: string) => {
    const body: IPLookupSubmission = {
      ip: quote.ip, refresh: quote.refresh, charge: quote.charge, useQuota: quote.useQuota,
      purchaseId: quote.purchaseId, remaining: quote.remaining, cacheReportId: quote.cacheReportId,
      configHash: quote.configHash, expiresAt: quote.expiresAt, token: quote.token,
    }
    return request<components['schemas']['OperationReceipt']>('/api/v1/ip-lookup/checks', { method: 'POST', body, headers: { 'Idempotency-Key': key } })
  },
  check: (id: string) => request<IPLookupCheck>(`/api/v1/ip-lookup/checks/${encodeURIComponent(id)}`, { cache: 'no-store' }),
  settings: () => request<IPLookupAdminSettings>('/api/v1/admin/ip-lookup', { cache: 'no-store' }),
  saveSettings: (value: IPLookupAdminSettings) => request<IPLookupAdminSettings>('/api/v1/admin/ip-lookup', { method: 'PUT', body: value }),
}
