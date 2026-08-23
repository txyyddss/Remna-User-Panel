import { describe, expect, it } from 'vitest'

import { compensationStatusChoices, fromCompensationStatus, toCompensationStatus } from './statusFilter'

describe('compensation status filter', () => {
  it('uses a non-empty select sentinel and omits it from API queries', () => {
    expect(compensationStatusChoices).not.toContain('')
    expect(fromCompensationStatus('')).toBe('all')
    expect(toCompensationStatus('all')).toBe('')
    expect(toCompensationStatus('pending_review')).toBe('pending_review')
  })
})
