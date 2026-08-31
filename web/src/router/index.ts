import { createRouter } from 'vue-router'

import { api } from '@/api/client'
import { useSessionStore } from '@/stores/session'
import { isTelegramWebAppDetected } from '@/utils/telegram'
import { resolveProtectedRoute } from './guards'
import { createAppHistory } from './history'
import { beginRouteRecovery, completeRouteRecovery, isRouteChunkError } from './recovery'

const router = createRouter({
  history: createAppHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    { path: '/', name: 'root', component: () => import('@/views/EntryView.vue') },
    {
      path: '/onboarding',
      name: 'onboarding',
      component: () => import('@/views/OnboardingView.vue'),
      meta: { immersive: true },
    },
    { path: '/home', name: 'home', component: () => import('@/views/HomeView.vue') },
    { path: '/connections', name: 'connections', component: () => import('@/views/ConnectionsView.vue') },
    { path: '/catalog', name: 'catalog', component: () => import('@/views/CatalogView.vue') },
    { path: '/payment-result', name: 'payment-result', component: () => import('@/views/PaymentResultView.vue'), meta: { immersive: true, browserPublic: true } },
    { path: '/balance', redirect: { name: 'home', query: { topUp: '1' } } },
    { path: '/activity', name: 'activity', component: () => import('@/views/ActivityView.vue') },
    { path: '/games', redirect: '/activity' },
    { path: '/statistics', name: 'statistics', component: () => import('@/views/StatisticsView.vue') },
    { path: '/affiliates', name: 'affiliates', component: () => import('@/views/AffiliatesView.vue') },
    { path: '/questionnaire', name: 'questionnaire', component: () => import('@/views/QuestionnaireView.vue') },
    { path: '/community', name: 'community', component: () => import('@/views/CommunityView.vue') },
    { path: '/emby', name: 'emby', component: () => import('@/views/EmbyView.vue') },
    { path: '/abuse-records', name: 'abuse-records', component: () => import('@/views/AbuseRecordsView.vue') },
    { path: '/admin', redirect: '/admin/settings' },
    { path: '/admin/emby', redirect: '/admin/users' },
    { path: '/admin/entitlements', redirect: '/admin/users' },
    {
      path: '/admin/users/:userId',
      name: 'admin-user',
      component: () => import('@/views/AdminUserView.vue'),
      meta: { adminSection: 'users' },
    },
    {
      path: '/admin/:section(settings|catalog|activity|affiliates|coupons|questionnaires|onboarding|users|compensation|abuse|backups|database|audit)',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.afterEach((to) => {
  completeRouteRecovery(to.fullPath)
})

router.onError((error, to) => {
  if (!isRouteChunkError(error) || !beginRouteRecovery(to.fullPath)) return
  if (isTelegramWebAppDetected()) {
    void router.replace(to.fullPath).catch(() => undefined)
    return
  }
  try {
    window.location.reload()
  } catch {
    // A WebView may reject reload while its host is changing state.
  }
})

router.beforeEach(async (to) => {
  const store = useSessionStore()
  if (to.meta.browserPublic === true && !isTelegramWebAppDetected()) return true
  await store.bootstrap()
  if (store.status === 'error') return true
  const protectedRedirect = resolveProtectedRoute(to, store.user)
  if (protectedRedirect) return protectedRedirect
  if (to.name !== 'catalog' || store.user?.role === 'admin') return true

  try {
    const dashboard = await api.getDashboard()
    if (dashboard.activePurchase?.autoRenewEnabled || dashboard.queuedPurchase?.autoRenewEnabled) {
      return { name: 'home', query: { autoRenewBlocked: '1' } }
    }
  } catch {
    // The catalog request remains server-enforced when a dashboard refresh is unavailable.
  }
  return true
})

export default router
