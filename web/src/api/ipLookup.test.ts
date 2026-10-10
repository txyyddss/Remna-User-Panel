import { describe, expect, it, vi } from 'vitest'
import type { IPLookupQuote } from './ipLookup'

const request = vi.hoisted(() => vi.fn().mockResolvedValue({ id: 'operation' }))
vi.mock('./http', () => ({ request }))
import { ipLookupApi } from './ipLookup'

describe('IP Lookup quote transport', () => {
  it('submits only signed action fields, excluding cached report presentation', async () => {
    const signed = {
      ip: '8.8.8.8', refresh: false, charge: { currency: 'TXB' as const, minor: '250', display: '2.50 TXB' },
      useQuota: false, purchaseId: '', remaining: 0, cacheReportId: '', configHash: 'config', expiresAt: 123, token: 'signature',
    }
    const quote: IPLookupQuote = { ...signed, cacheMatch: 'none', cachedReport: null }
    await ipLookupApi.submit(quote, 'idempotency-key')
    expect(request).toHaveBeenCalledWith('/api/v1/ip-lookup/checks', {
      method: 'POST', body: signed, headers: { 'Idempotency-Key': 'idempotency-key' },
    })
    expect(request.mock.calls[0]![1].body).not.toHaveProperty('cacheMatch')
    expect(request.mock.calls[0]![1].body).not.toHaveProperty('cachedReport')
  })
})
