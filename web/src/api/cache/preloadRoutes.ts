import type { Session } from '../types'
import type { RequestOptions } from '../http'

export interface PreloadResource { path: string; options?: RequestOptions }

export function memberResources(session: Session, locale: string): PreloadResource[] {
  const onboarding = { path: '/api/v1/onboarding/content', options: { query: { locale } } }
  if (session.user.onboardingState !== 'complete') return [onboarding]
  const end = new Date()
  const start = new Date(end)
  start.setUTCDate(start.getUTCDate() - 6)
  return [
    ...['dashboard', 'catalog', 'balance', 'statistics', 'statistics/nodes', 'activity',
      'affiliates', 'coupons/wallet', 'questionnaires/active', 'emby/account',
      'me/abuse-records', 'subscription/ip-blocks', 'me/traffic-reset-automation',
      'community/membership/check'].map(path => ({ path: `/api/v1/${path}` })),
    { path: '/api/v1/affiliates/referrals', options: { query: { page: 1 } } },
    { path: '/api/v1/dashboard/node-usage', options: { query: { start: start.toISOString().slice(0, 10), end: end.toISOString().slice(0, 10) } } },
    { path: '/api/v1/community/membership/check', options: { method: 'POST' } },
    onboarding,
  ]
}

export function adminResources(): PreloadResource[] {
  return [
    ...['settings', 'payment-profiles', 'activity-settings', 'combos', 'squad-products',
      'activity-games', 'lucky-draw', 'affiliates', 'coupons', 'questionnaires',
      'onboarding/content/welcome', 'onboarding/content/agreements', 'users',
      'emby-accounts', 'node-compensation/config', 'abuse/policy', 'abuse/nodes',
      'abuse/rules', 'abuse/punishments', 'abuse/records', 'abuse/statistics',
      'abuse/whitelist', 'backups', 'jobs', 'database/tables', 'audit-events',
    ].map(path => ({ path: `/api/v1/admin/${path}` })),
    { path: '/api/v1/admin/node-compensation/events', options: { query: { limit: 25 } } },
    // Admin settings includes amount limits even before member onboarding is complete.
    { path: '/api/v1/balance' },
  ]
}
