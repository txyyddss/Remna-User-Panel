import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import ComingSoonLinks from './ComingSoonLinks.vue'

describe('ComingSoonLinks navigation', () => {
  it('keeps member-tool actions inside Vue Router history', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
    })
    await router.push('/home')
    await router.isReady()
    const wrapper = mount(ComingSoonLinks, { global: { plugins: [router] } })
    const launchURL = window.location.href
    const actions = wrapper.findAll('.home-around__link')
    const destinations = ['/affiliates', '/questionnaire', '/community', '/emby', '/statistics', '/abuse-records']

    expect(actions).toHaveLength(destinations.length)
    for (const [index, destination] of destinations.entries()) {
      expect(actions[index].attributes('href')).toBeUndefined()
      await actions[index].trigger('click')
      await new Promise<void>((resolve) => setTimeout(resolve, 0))
      expect(router.currentRoute.value.path).toBe(destination)
      expect(window.location.href).toBe(launchURL)
      await router.push('/home')
    }
  })
})
