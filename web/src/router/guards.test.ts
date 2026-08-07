import type { RouteLocationNormalized } from 'vue-router'
import { describe, expect, it } from 'vitest'

import type { User } from '@/api/types'
import { resolveProtectedRoute } from './guards'

function route(path: string): RouteLocationNormalized {
  return { path } as RouteLocationNormalized
}

function user(overrides: Partial<User> = {}): User {
  return {
    id: 'user-1',
    telegramId: '42',
    firstName: 'Mira',
    lastName: 'Lin',
    telegramUsername: 'miralin',
    username: 'mira',
    role: 'user',
    onboardingState: 'complete',
    groupJoined: true,
    channelJoined: true,
    policyAcceptedAt: '2026-08-07T00:00:00Z',
    createdAt: '2026-08-07T00:00:00Z',
    updatedAt: '2026-08-07T00:00:00Z',
    ...overrides,
  }
}

describe('protected route decisions', () => {
  it('resumes incomplete onboarding', () => {
    expect(resolveProtectedRoute(route('/catalog'), user({ onboardingState: 'username' }))).toBe('/onboarding')
  })

  it('keeps completed users out of onboarding', () => {
    expect(resolveProtectedRoute(route('/onboarding'), user())).toBe('/home')
  })

  it('rejects non-admin users from every admin section', () => {
    expect(resolveProtectedRoute(route('/admin/payments'), user())).toBe('/home')
  })

  it('lets an admin bootstrap settings before onboarding completes', () => {
    expect(resolveProtectedRoute(route('/admin/settings'), user({ role: 'admin', onboardingState: 'membership' }))).toBeNull()
  })
})
