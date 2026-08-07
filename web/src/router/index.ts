import { createRouter, createWebHistory } from 'vue-router'

import { useSessionStore } from '@/stores/session'
import { resolveProtectedRoute } from './guards'

const router = createRouter({
  history: createWebHistory(),
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
    { path: '/balance', name: 'balance', component: () => import('@/views/BalanceView.vue') },
    { path: '/games', name: 'games', component: () => import('@/views/ComingSoonView.vue'), meta: { product: 'Games' } },
    { path: '/questionnaire', name: 'questionnaire', component: () => import('@/views/ComingSoonView.vue'), meta: { product: 'Questionnaire' } },
    { path: '/emby', name: 'emby', component: () => import('@/views/ComingSoonView.vue'), meta: { product: 'Emby' } },
    { path: '/admin', redirect: '/admin/settings' },
    {
      path: '/admin/:section(settings|catalog|users|entitlements|payments|backups|audit)',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach(async (to) => {
  const store = useSessionStore()
  await store.bootstrap()
  if (store.status === 'error') return true
  return resolveProtectedRoute(to, store.user) ?? true
})

export default router
