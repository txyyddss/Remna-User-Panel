import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { abuseApi, type AbusePolicy, type AbusePunishment, type AbuseRule } from './abuse'

const fetchMock = vi.fn()

function jsonResponse(payload: unknown): Response {
  return new Response(JSON.stringify(payload), { status: 200, headers: { 'Content-Type': 'application/json' } })
}

function requestAt(index = 0): [string, RequestInit] {
  return fetchMock.mock.calls[index] as [string, RequestInit]
}

describe('abuse administration API mutations', () => {
  beforeEach(() => vi.stubGlobal('fetch', fetchMock))
  afterEach(() => {
    fetchMock.mockReset()
    vi.unstubAllGlobals()
  })

  it('creates a domain rule without sending the route-owned id', async () => {
    const draft: AbuseRule = { id: '', name: 'Streaming', expression: '(^|\\.)example\\.com$', qpsLimit: 25, enabled: true, revision: 0 }
    fetchMock.mockResolvedValue(jsonResponse({ ...draft, id: 'rule-1' }))

    await abuseApi.saveRule(draft)

    const [url, options] = requestAt()
    expect(url).toBe('/api/v1/admin/abuse/rules')
    expect(options.method).toBe('POST')
    expect(JSON.parse(String(options.body))).toEqual({ name: draft.name, expression: draft.expression, qpsLimit: draft.qpsLimit, enabled: draft.enabled, revision: draft.revision })
  })

  it('updates a domain rule with its encoded id only in the path', async () => {
    const rule: AbuseRule = { id: 'rule/1', name: 'Downloads', expression: 'downloads\\.example$', qpsLimit: 80, enabled: false, revision: 4 }
    fetchMock.mockResolvedValue(jsonResponse({ ...rule, revision: 5 }))

    await abuseApi.saveRule(rule)

    const [url, options] = requestAt()
    expect(url).toBe('/api/v1/admin/abuse/rules/rule%2F1')
    expect(options.method).toBe('PUT')
    expect(JSON.parse(String(options.body))).toEqual({ name: rule.name, expression: rule.expression, qpsLimit: rule.qpsLimit, enabled: rule.enabled, revision: rule.revision })
  })

  it('updates a punishment with its action only in the path', async () => {
    const punishment: AbusePunishment = { action: 'ip_ban', enabled: true, incidentThreshold: 3, durationMinutes: 90, allNodes: true, revision: 2 }
    fetchMock.mockResolvedValue(jsonResponse({ ...punishment, revision: 3 }))

    await abuseApi.savePunishment(punishment)

    const [url, options] = requestAt()
    expect(url).toBe('/api/v1/admin/abuse/punishments/ip_ban')
    expect(options.method).toBe('PUT')
    expect(JSON.parse(String(options.body))).toEqual({ enabled: punishment.enabled, incidentThreshold: punishment.incidentThreshold, durationMinutes: punishment.durationMinutes, allNodes: punishment.allNodes, revision: punishment.revision })
  })

  it('sends the required streak with every policy field', async () => {
    const policy: AbusePolicy = { globalEnabled: true, globalLimit: 50, streakSeconds: 75, warningValidityDays: 7, warningCooldownMinutes: 30, revision: 4 }
    fetchMock.mockResolvedValue(jsonResponse({ ...policy, revision: 5 }))

    await abuseApi.savePolicy(policy)

    const [url, options] = requestAt()
    expect(url).toBe('/api/v1/admin/abuse/policy')
    expect(options.method).toBe('PUT')
    expect(JSON.parse(String(options.body))).toEqual(policy)
  })
})
