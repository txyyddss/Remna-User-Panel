import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import type { Money, Purchase } from '@/api/types'
import CatalogPaymentStep from './CatalogPaymentStep.vue'

const balance: Money = { currency: 'TXB', minor: '0', display: '0 TXB' }

async function mountStep(purchase: Purchase | null = null) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  })
  await router.push('/catalog')
  await router.isReady()
  return { router, wrapper: mount(CatalogPaymentStep, {
    global: {
      plugins: [router],
      stubs: { CatalogConfirmation: { template: '<div data-test="shared-confirmation" />' } },
    },
    props: { balance, quote: null, purchase, purchasing: false, needsBalance: true },
  }) }
}

describe('CatalogPaymentStep navigation', () => {
  it('keeps the add-balance action in Vue Router history', async () => {
    const { router, wrapper } = await mountStep()
    const launchURL = window.location.href

    await wrapper.get('button').trigger('click')
    await new Promise<void>((resolve) => setTimeout(resolve, 0))

    expect(router.currentRoute.value.fullPath).toBe('/home?topUp=1')
    expect(window.location.href).toBe(launchURL)
  })

  it('uses the shared confirmation screen after a purchase', async () => {
    const { wrapper } = await mountStep({ comboName: 'Combo' } as Purchase)

    expect(wrapper.find('[data-test="shared-confirmation"]').exists()).toBe(true)
    wrapper.unmount()
  })
})
