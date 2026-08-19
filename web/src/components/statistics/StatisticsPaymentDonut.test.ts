import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import StatisticsPaymentDonut from './StatisticsPaymentDonut.vue'

describe('StatisticsPaymentDonut', () => {
  it('renders EZPay outside BEPusdt and ignores failed payments', () => {
    const wrapper = mount(StatisticsPaymentDonut, {
      props: { items: [
        { id: 'ezpay:paid', label: 'Paid', value: 3 },
        { id: 'bepusdt:expired', label: 'Expired', value: 1 },
        { id: 'bepusdt:failed', label: 'Failed', value: 9 },
      ] },
      global: { stubs: { UIcon: true } },
    })

    expect(wrapper.find('svg[role="img"]').exists()).toBe(true)
    expect(wrapper.findAll('.statistics-concentric-ring--outer.statistics-ring-segment')).toHaveLength(1)
    expect(wrapper.findAll('.statistics-concentric-ring--inner.statistics-ring-segment')).toHaveLength(1)
  })

  it('renders a specific empty state when no terminal payments exist', () => {
    const wrapper = mount(StatisticsPaymentDonut, {
      props: { items: [] },
      global: { stubs: { UIcon: true } },
    })

    expect(wrapper.find('.statistics-concentric-rings').exists()).toBe(false)
    expect(wrapper.text()).toContain('No terminal payments recorded')
  })
})
