import { describe, expect, it } from 'vitest'

import { adminReasonSchema, identifierSchema, isValid, txbInputSchema, usernameSchema } from './validation'

describe('form validation schemas', () => {
  it('accepts only canonical permanent usernames', () => {
    expect(isValid(usernameSchema, 'member')).toBe(true)
    expect(isValid(usernameSchema, 'Member_1')).toBe(false)
  })

  it('accepts TXB input with at most two decimal places', () => {
    expect(isValid(txbInputSchema, '12.50')).toBe(true)
    expect(isValid(txbInputSchema, '01')).toBe(false)
    expect(isValid(txbInputSchema, '1.005')).toBe(false)
  })

  it('bounds reasons and identifiers', () => {
    expect(isValid(adminReasonSchema, 'Manual reconciliation')).toBe(true)
    expect(isValid(adminReasonSchema, 'no')).toBe(false)
    expect(isValid(identifierSchema, 'payment:01_test')).toBe(true)
    expect(isValid(identifierSchema, '../payment')).toBe(false)
  })
})
