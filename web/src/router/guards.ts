import type { RouteLocationNormalized } from 'vue-router'

import type { User } from '@/api/types'

export function resolveProtectedRoute(to: RouteLocationNormalized, user: User | null): string | null {
  if (!user) return null

  const wantsAdmin = to.path.startsWith('/admin')
  if (wantsAdmin && user.role !== 'admin') return '/home'

  if (user.onboardingState !== 'complete' && !wantsAdmin && to.path !== '/onboarding') {
    return '/onboarding'
  }

  if (user.onboardingState === 'complete' && to.path === '/onboarding') return '/home'
  if (to.path === '/') return user.onboardingState === 'complete' ? '/home' : '/onboarding'

  return null
}
