import { createRouter } from 'vue-router'

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
    { path: '/catalog', name: 'catalog', component: () => import('@/views/CatalogView.vue') },
    { path: '/payment-result', name: 'payment-result', component: () => import('@/views/PaymentResultView.vue'), meta: { immersive: true, browserPublic: true } },
    { path: '/balance', redirect: { name: 'home', query: { topUp: '1' } } },
    { path: '/activity', name: 'activity', component: () => import('@/views/ActivityView.vue') },
    { path: '/games', redirect: '/activity' },
    { path: '/questionnaire', name: 'questionnaire', component: () => import('@/views/QuestionnaireView.vue') },
    { path: '/emby', name: 'emby', component: () => import('@/views/EmbyView.vue') },
    { path: '/admin', redirect: '/admin/settings' },
    {
      path: '/admin/:section(settings|catalog|activity|coupons|questionnaires|onboarding|users|emby|entitlements|payments|backups|database|audit)',
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
  return resolveProtectedRoute(to, store.user) ?? true
})

export default router
