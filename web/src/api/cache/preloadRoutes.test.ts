import { describe, expect, it } from 'vitest'
import type { Session } from '../types'
import { adminResources, memberResources } from './preloadRoutes'
import { isReadRequest } from './policy'

describe('initial page preload inventory', () => {
  const session = (role: string, onboardingState: string) => ({ user: { role, onboardingState } }) as Session

  it('keeps incomplete members on localized onboarding data', () => {
    expect(memberResources(session('member', 'agreement'), 'zh-CN')).toEqual([
      { path: '/api/v1/onboarding/content', options: { query: { locale: 'zh-CN' } } },
    ])
  })

  it('covers member pages with their initial queries and no admin resources', () => {
    const resources = memberResources(session('member', 'complete'), 'en')
    const paths = resources.map(resource => resource.path)
    expect(paths).toEqual(expect.arrayContaining([
      '/api/v1/dashboard', '/api/v1/catalog', '/api/v1/statistics', '/api/v1/activity',
      '/api/v1/affiliates', '/api/v1/emby/account', '/api/v1/questionnaires/active',
      '/api/v1/me/abuse-records', '/api/v1/subscription/ip-blocks',
    ]))
    expect(paths.some(path => path.includes('/admin/'))).toBe(false)
    expect(resources.find(resource => resource.path.endsWith('/referrals'))?.options?.query).toEqual({ page: 1 })
    expect(resources.every(resource => isReadRequest(resource.path, resource.options?.method ?? 'GET'))).toBe(true)
  })

  it('covers admin sections without running operational commands', () => {
    const resources = adminResources()
    expect(resources.map(resource => resource.path)).toEqual(expect.arrayContaining([
      '/api/v1/admin/settings', '/api/v1/admin/users', '/api/v1/admin/combos',
      '/api/v1/admin/activity-games', '/api/v1/admin/affiliates', '/api/v1/admin/coupons',
      '/api/v1/admin/questionnaires', '/api/v1/admin/onboarding/content/agreements',
      '/api/v1/admin/node-compensation/events', '/api/v1/admin/abuse/records',
      '/api/v1/admin/backups', '/api/v1/admin/database/tables', '/api/v1/admin/audit-events',
    ]))
    expect(resources.every(resource => !resource.options?.method || resource.options.method === 'GET')).toBe(true)
  })
})
