import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import StatisticsPaymentDonut from './StatisticsPaymentDonut.vue'

describe('StatisticsPaymentDonut', () => {
  it('renders one generic donut for every terminal payment state', () => {
    const wrapper = mount(StatisticsPaymentDonut, {
      props: { items: [
        { id: 'ezpay:paid', label: 'Paid', value: 3 },
        { id: 'stars:expired', label: 'Expired', value: 1 },
      ] },
      global: { stubs: { UIcon: true } },
    })

    expect(wrapper.find('svg[role="img"]').exists()).toBe(true)
    expect(wrapper.find('.statistics-donut').exists()).toBe(true)
    expect(wrapper.findAll('.statistics-ring-segment')).toHaveLength(2)
    expect(wrapper.find('.statistics-payment-rings').exists()).toBe(false)
  })

  it('renders a specific empty state when no terminal payments exist', () => {
    const wrapper = mount(StatisticsPaymentDonut, {
      props: { items: [] },
      global: { stubs: { UIcon: true } },
    })

    expect(wrapper.find('.statistics-donut').exists()).toBe(false)
    expect(wrapper.text()).toContain('No terminal payments recorded')
  })
})
