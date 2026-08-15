import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { Combo } from '@/api/types'
import ComboOption from './ComboOption.vue'

const combo: Combo = {
  id: 'combo-1',
  name: 'Weekend North',
  description: 'A compact monthly plan for a small device set.',
  price: { currency: 'TXB', minor: '1880', display: '18.80 TXB' },
  validityDays: 30,
  trafficLimitBytes: '107374182400',
  resetStrategy: 'MONTH',
  active: true,
  includedSquads: [],
  rolloverMinRemainingBps: 0,
  createdAt: '2026-08-07T00:00:00Z',
  updatedAt: '2026-08-07T00:00:00Z',
}

describe('ComboOption', () => {
  it('renders server-formatted price and entitlement values', () => {
    const wrapper = mount(ComboOption, { props: { combo, selected: false } })
    expect(wrapper.text()).toContain('18.80 TXB')
    expect(wrapper.text()).toContain('100 GB')
    expect(wrapper.text()).toContain('30 days')
  })

  it('emits the stable combo id on selection', async () => {
    const wrapper = mount(ComboOption, { props: { combo, selected: false } })
    await wrapper.trigger('click')
    expect(wrapper.emitted('select')).toEqual([['combo-1']])
  })
})
