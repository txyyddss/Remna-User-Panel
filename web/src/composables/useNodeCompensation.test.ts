import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { NodeCompensationEvent } from '@/api/contracts/compensation'

const mocks = vi.hoisted(() => ({
  approve: vi.fn(),
  createUuid: vi.fn(() => 'stable-review-key'),
  notifyHaptic: vi.fn(),
}))

vi.mock('@/api/compensation', () => ({
  compensationApi: {
    config: vi.fn(), events: vi.fn(), saveConfig: vi.fn(),
    approve: mocks.approve, dismiss: vi.fn(),
  },
}))
vi.mock('@/utils/browserCompatibility', () => ({ createUuid: mocks.createUuid }))
vi.mock('@/utils/telegram', () => ({ notifyHaptic: mocks.notifyHaptic }))

import { useNodeCompensation } from './useNodeCompensation'

const event: NodeCompensationEvent = {
  id: 'event-1', nodeUuid: 'node-1', nodeName: 'Node', status: 'pending_review',
  offlineObservedAt: '2026-08-23T00:00:00Z', recoveredObservedAt: '2026-08-23T02:00:00Z',
  observedDurationSeconds: 7200, thresholdMinutes: 60, multiplierBps: 10_000,
  proposedExtensionMinutes: 120, finalExtensionMinutes: null, capped: false,
  squads: [], frozenRecipientCount: 1, eligibleRecipientCount: null,
  skippedRecipientCount: null, reviewedBy: null, reviewedAt: null,
  reviewReason: null, ineligibleReason: null, revision: 1, operation: null,
}

describe('useNodeCompensation review durability', () => {
  beforeEach(() => vi.clearAllMocks())

  it('reuses one idempotency key after an ambiguous failure', async () => {
    mocks.approve.mockRejectedValueOnce(new Error('network')).mockResolvedValueOnce({ ...event, status: 'queued' })
    const state = useNodeCompensation()

    expect(await state.review(event, 'approve', 120, 'reviewed')).toBe(false)
    expect(await state.review(event, 'approve', 120, 'reviewed')).toBe(true)
    expect(mocks.createUuid).toHaveBeenCalledOnce()
    expect(mocks.approve.mock.calls.map((call) => call[2])).toEqual(['stable-review-key', 'stable-review-key'])
  })
})
