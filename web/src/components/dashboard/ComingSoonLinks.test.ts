import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import ComingSoonLinks from './ComingSoonLinks.vue'

describe('ComingSoonLinks navigation', () => {
  it('keeps questionnaire and Emby actions inside Vue Router history', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
    })
    await router.push('/home')
    await router.isReady()
    const wrapper = mount(ComingSoonLinks, { global: { plugins: [router] } })
    const launchURL = window.location.href
    const actions = wrapper.findAll('.home-around__link')

    expect(actions).toHaveLength(2)
    expect(actions[0].attributes('href')).toBeUndefined()
    expect(actions[1].attributes('href')).toBeUndefined()

    await actions[0].trigger('click')
    await new Promise<void>((resolve) => setTimeout(resolve, 0))
    expect(router.currentRoute.value.path).toBe('/questionnaire')
    expect(window.location.href).toBe(launchURL)

    await router.push('/home')
    await actions[1].trigger('click')
    await new Promise<void>((resolve) => setTimeout(resolve, 0))
    expect(router.currentRoute.value.path).toBe('/emby')
    expect(window.location.href).toBe(launchURL)
  })
})
