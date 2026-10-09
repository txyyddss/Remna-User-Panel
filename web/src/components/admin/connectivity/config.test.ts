import { describe, expect, it } from 'vitest'
import { defaultConnectivityConfig, sameConnectivityConfig, validConnectivityConfig } from './config'

describe('connectivity retry configuration', () => {
  it('defaults to ten additional retries with a one-second delay', () => {
    expect(defaultConnectivityConfig.maxRetries).toBe(10)
    expect(defaultConnectivityConfig.retryIntervalSeconds).toBe(1)
    expect(validConnectivityConfig(defaultConnectivityConfig)).toBe(true)
  })

  it.each([
    { maxRetries: 0, retryIntervalSeconds: 1, valid: true },
    { maxRetries: 10, retryIntervalSeconds: 60, valid: true },
    { maxRetries: -1, retryIntervalSeconds: 1, valid: false },
    { maxRetries: 11, retryIntervalSeconds: 1, valid: false },
    { maxRetries: 1.5, retryIntervalSeconds: 1, valid: false },
    { maxRetries: 10, retryIntervalSeconds: 0, valid: false },
    { maxRetries: 10, retryIntervalSeconds: 61, valid: false },
    { maxRetries: 10, retryIntervalSeconds: 1.5, valid: false },
    { maxRetries: Number.NaN, retryIntervalSeconds: 1, valid: false },
  ])('validates retry bounds: $maxRetries retries and $retryIntervalSeconds seconds', ({ maxRetries, retryIntervalSeconds, valid }) => {
    expect(validConnectivityConfig({ ...defaultConnectivityConfig, maxRetries, retryIntervalSeconds })).toBe(valid)
  })

  it('detects independent retry edits for dirty state', () => {
    expect(sameConnectivityConfig(defaultConnectivityConfig, { ...defaultConnectivityConfig })).toBe(true)
    expect(sameConnectivityConfig(defaultConnectivityConfig, { ...defaultConnectivityConfig, maxRetries: 0 })).toBe(false)
    expect(sameConnectivityConfig(defaultConnectivityConfig, { ...defaultConnectivityConfig, retryIntervalSeconds: 2 })).toBe(false)
  })
})
