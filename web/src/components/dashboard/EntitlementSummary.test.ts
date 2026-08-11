import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import EntitlementSummary from './EntitlementSummary.vue'

describe('EntitlementSummary navigation', () => {
  it('keeps the catalog action in Vue Router history', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
    })
    await router.push('/home')
    await router.isReady()
    const wrapper = mount(EntitlementSummary, { global: { plugins: [router] } })
    const launchURL = window.location.href

    await wrapper.get('.empty-inline button').trigger('click')
    await new Promise<void>((resolve) => setTimeout(resolve, 0))

    expect(router.currentRoute.value.path).toBe('/catalog')
    expect(window.location.href).toBe(launchURL)
  })
})
