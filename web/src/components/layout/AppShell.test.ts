import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { nextTick } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Session } from '@/api/types'
import { useSessionStore } from '@/stores/session'
import AppShell from './AppShell.vue'

const session: Session = {
  authenticated: true,
  user: {
    id: 'user-1', telegramId: '42', firstName: 'Mira', lastName: 'Lin', telegramUsername: 'mira', username: 'mira',
    role: 'user', onboardingState: 'complete', groupJoined: true, channelJoined: true,
    policyAcceptedAt: '2026-08-08T00:00:00Z', agreementRevision: 1, recoveryReason: '', createdAt: '2026-08-08T00:00:00Z', updatedAt: '2026-08-08T00:00:00Z',
  },
}

describe('AppShell accessibility', () => {
  afterEach(() => {
    delete window.Telegram
    document.body.innerHTML = ''
  })

  it('restores focus after routing and drives Telegram BackButton', async () => {
    const show = vi.fn()
    const hide = vi.fn()
    const onClick = vi.fn()
    const offClick = vi.fn()
    window.Telegram = { WebApp: {
      version: '9.0', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(),
      openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn(),
      BackButton: { isVisible: false, show, hide, onClick, offClick },
    } }
    const pinia = createPinia()
    useSessionStore(pinia).session = session
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/home', component: { template: '<div />' } },
        { path: '/catalog', component: { template: '<div />' } },
        { path: '/balance', component: { template: '<div />' } },
        { path: '/activity', component: { template: '<div />' } },
        { path: '/questionnaire', component: { template: '<div />' } },
        { path: '/emby', component: { template: '<div />' } },
        { path: '/admin/settings', component: { template: '<div />' } },
      ],
    })
    await router.push('/home')
    await router.isReady()
    const wrapper = mount(AppShell, { attachTo: document.body, global: { plugins: [pinia, router] }, slots: { default: '<h1>Content</h1>' } })

    expect(wrapper.get('main').attributes('tabindex')).toBe('-1')
    await router.push('/catalog')
    await nextTick()
    expect(document.activeElement).toBe(wrapper.get('main').element)
    expect(show).toHaveBeenCalled()
    expect(onClick).toHaveBeenCalledOnce()

    wrapper.unmount()
    expect(offClick).toHaveBeenCalledOnce()
  })
})
