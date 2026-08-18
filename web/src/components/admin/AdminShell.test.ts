import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import type { Session } from '@/api/types'
import { useSessionStore } from '@/stores/session'
import AdminShell from './AdminShell.vue'

function session(onboardingState: Session['user']['onboardingState']): Session {
  return {
    authenticated: true,
    user: {
      id: 'admin-1',
      telegramId: '42',
      firstName: 'Ada',
      lastName: 'Admin',
      telegramUsername: 'adaadmin',
      username: null,
      role: 'admin',
      onboardingState,
      groupJoined: false,
      channelJoined: false,
      policyAcceptedAt: null,
      agreementRevision: 0,
      recoveryReason: '',
      createdAt: '2026-08-08T00:00:00Z',
      updatedAt: '2026-08-08T00:00:00Z',
    },
  }
}

async function mountShell(onboardingState: Session['user']['onboardingState']) {
  const pinia = createPinia()
  useSessionStore(pinia).session = session(onboardingState)
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  })
  await router.push('/admin/settings')
  await router.isReady()

  return {
    router,
    wrapper: mount(AdminShell, {
      global: { plugins: [pinia, router] },
      slots: { default: '<div>Admin panel</div>' },
    }),
  }
}

describe('AdminShell', () => {
  it('offers optional signup to an admin without a completed user account', async () => {
    const { wrapper } = await mountShell('intro')

    expect(wrapper.get('.button-row button').text()).toContain('Set up user account')
  })

  it('removes the signup entry after the admin completes onboarding', async () => {
    const { wrapper } = await mountShell('complete')

    expect(wrapper.find('.button-row button').exists()).toBe(false)
  })

  it('exposes the combined user account section', async () => {
    const { wrapper } = await mountShell('complete')
    const items = wrapper.getComponent({ name: 'Select' }).props('items') as Array<{ value: string }>

    expect(items.map(({ value }) => value)).toEqual(expect.arrayContaining(['users']))
    expect(items.map(({ value }) => value)).not.toContain('emby')
  })

  it('keeps the setup button inside the router history', async () => {
    const { router, wrapper } = await mountShell('intro')
    const launchURL = window.location.href

    await wrapper.get('.button-row button').trigger('click')
    await new Promise<void>((resolve) => setTimeout(resolve, 0))

    expect(router.currentRoute.value.path).toBe('/onboarding')
    expect(window.location.href).toBe(launchURL)
  })
})
