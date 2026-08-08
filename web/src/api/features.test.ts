import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { featuresApi } from './features'

const fetchMock = vi.fn()

function jsonResponse(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), { status, headers: { 'Content-Type': 'application/json' } })
}

describe('forward feature API contracts', () => {
  beforeEach(() => vi.stubGlobal('fetch', fetchMock))
  afterEach(() => {
    fetchMock.mockReset()
    vi.unstubAllGlobals()
  })

  it('reviews a database mutation before applying the exact review hash and confirmation', async () => {
    const input = { action: 'update' as const, table: 'users', key: { id: 'user-1' }, values: { role: 'admin' }, recordHash: 'old-hash', reason: 'Repair imported role' }
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ ...input, before: { role: 'user' }, after: { role: 'admin' }, changedColumns: ['role'], reviewHash: 'review-hash', requiredConfirmation: 'EDIT users', rescueBackupRequired: true, warning: 'Direct edits bypass hooks.' }))
      .mockResolvedValueOnce(jsonResponse({ row: { key: input.key, values: input.values, recordHash: 'new-hash' }, deleted: false, rescueBackupId: 'backup-1' }))

    const review = await featuresApi.reviewDatabaseMutation(input)
    expect(review.requiredConfirmation).toBe('EDIT users')
    const [reviewUrl, reviewOptions] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(reviewUrl).toBe('/api/v1/admin/database/mutations/review')
    expect(JSON.parse(String(reviewOptions.body))).toEqual(input)

    await featuresApi.applyDatabaseMutation({ ...input, reviewHash: review.reviewHash, confirmation: review.requiredConfirmation })
    const [applyUrl, applyOptions] = fetchMock.mock.calls[1] as [string, RequestInit]
    expect(applyUrl).toBe('/api/v1/admin/database/mutations')
    expect(JSON.parse(String(applyOptions.body))).toEqual({ ...input, reviewHash: 'review-hash', confirmation: 'EDIT users' })
  })
})
