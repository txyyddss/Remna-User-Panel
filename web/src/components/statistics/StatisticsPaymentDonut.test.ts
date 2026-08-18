import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import StatisticsPaymentDonut from './StatisticsPaymentDonut.vue'

describe('StatisticsPaymentDonut', () => {
  it('renders EZPay outside BEPusdt as concentric SVG rings', () => {
    const wrapper = mount(StatisticsPaymentDonut, {
      props: { items: [
        { id: 'ezpay:paid', label: 'Paid', value: 3 },
        { id: 'bepusdt:expired', label: 'Expired', value: 1 },
      ] },
      global: { stubs: { UIcon: true } },
    })

    expect(wrapper.find('svg[role="img"]').exists()).toBe(true)
    expect(wrapper.findAll('.statistics-ring-segment.statistics-payment-ring--outer')).toHaveLength(1)
    expect(wrapper.findAll('.statistics-ring-segment.statistics-payment-ring--inner')).toHaveLength(1)
    expect(wrapper.find('.statistics-payment-ring--outer').attributes('r')).toBe('50')
    expect(wrapper.find('.statistics-payment-ring--inner').attributes('r')).toBe('32')
  })

  it('renders a specific empty state when no terminal payments exist', () => {
    const wrapper = mount(StatisticsPaymentDonut, {
      props: { items: [] },
      global: { stubs: { UIcon: true } },
    })

    expect(wrapper.find('.statistics-payment-rings').exists()).toBe(false)
    expect(wrapper.text()).toContain('No terminal payments recorded')
  })
})
