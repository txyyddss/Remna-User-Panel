import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        {
            path: '/',
            name: 'home',
            component: () => import('@/views/HomeView.vue'),
        },
        {
            path: '/sub',
            name: 'subscription',
            component: () => import('@/views/SubView.vue'),
        },
        {
            path: '/combos',
            name: 'combos',
            component: () => import('@/views/CombosView.vue'),
        },
        {
            path: '/credits',
            name: 'credits',
            component: () => import('@/views/CreditsView.vue'),
        },
        {
            path: '/jellyfin',
            name: 'jellyfin',
            component: () => import('@/views/JellyfinView.vue'),
        },
        {
            path: '/info',
            name: 'info',
            component: () => import('@/views/InfoView.vue'),
        },
        {
            path: '/squads',
            name: 'squads',
            component: () => import('@/views/SquadsView.vue'),
        },
        {
            path: '/ip',
            name: 'ip-change',
            component: () => import('@/views/IPChangeView.vue'),
        },
        {
            path: '/admin',
            name: 'admin',
            component: () => import('@/views/AdminView.vue'),
        },
        {
            path: '/blocked',
            name: 'blocked',
            component: () => import('@/views/BlockedView.vue'),
        },
    ],
})

export default router
