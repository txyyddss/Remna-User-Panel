import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { nextTick } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Session } from '@/api/types'
import { useSessionStore } from '@/stores/session'
import AppShell from './AppShell.vue'

function session(role: Session['user']['role'] = 'user'): Session {
  return {
    authenticated: true,
    user: {
      id: 'user-1', telegramId: '42', firstName: 'Mira', lastName: 'Lin', telegramUsername: 'mira', username: 'mira',
      role, onboardingState: 'complete', groupJoined: true, channelJoined: true,
      policyAcceptedAt: '2026-08-08T00:00:00Z', agreementRevision: 1, recoveryReason: '', createdAt: '2026-08-08T00:00:00Z', updatedAt: '2026-08-08T00:00:00Z',
    },
  }
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
    const impactOccurred = vi.fn()
    window.Telegram = { WebApp: {
      version: '9.0', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(),
      openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn(),
      HapticFeedback: { impactOccurred, notificationOccurred: vi.fn() },
      BackButton: { isVisible: false, show, hide, onClick, offClick },
    } }
    const pinia = createPinia()
    useSessionStore(pinia).session = session()
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
    onClick.mock.calls[0]?.[0]?.()
    expect(impactOccurred).toHaveBeenCalledWith('light')

    const launchURL = window.location.href
    await wrapper.get('.side-rail a[href="/catalog"]').trigger('click')
    await nextTick()
    expect(router.currentRoute.value.path).toBe('/catalog')
    expect(window.location.href).toBe(launchURL)

    await router.push('/home')
    await nextTick()
    const bottomNavigationItems = wrapper.findAll('.bottom-nav__item')
    expect(bottomNavigationItems).toHaveLength(3)
    await bottomNavigationItems[1].trigger('click')
    await vi.waitFor(() => expect(router.currentRoute.value.path).toBe('/catalog'))
    expect(window.location.href).toBe(launchURL)

    wrapper.unmount()
    expect(offClick).toHaveBeenCalledOnce()
  })

  it('adds the admin entry to mobile navigation for administrators', async () => {
    const pinia = createPinia()
    useSessionStore(pinia).session = session('admin')
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/home', component: { template: '<div />' } },
        { path: '/catalog', component: { template: '<div />' } },
        { path: '/activity', component: { template: '<div />' } },
        { path: '/admin/settings', component: { template: '<div />' } },
      ],
    })
    await router.push('/home')
    await router.isReady()
    const wrapper = mount(AppShell, { global: { plugins: [pinia, router] }, slots: { default: '<h1>Content</h1>' } })

    const bottomNavigationItems = wrapper.findAll('.bottom-nav__item')
    expect(bottomNavigationItems).toHaveLength(4)
    expect(bottomNavigationItems[3]?.text()).toContain('Admin')

    await bottomNavigationItems[3]?.trigger('click')
    await vi.waitFor(() => expect(router.currentRoute.value.path).toBe('/admin/settings'))
    expect(wrapper.find('.bottom-nav').classes()).toContain('bottom-nav--admin')

    wrapper.unmount()
  })
})
