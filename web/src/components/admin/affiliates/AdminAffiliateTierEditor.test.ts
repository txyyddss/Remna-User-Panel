import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AdminAffiliateTierEditor from './AdminAffiliateTierEditor.vue'

const base = { id: 'one', name: 'Starter', threshold: 0, enabled: true, commissionEnabled: false, commissionBps: 0, reward: { kind: 'none' as const } }

describe('AdminAffiliateTierEditor', () => {
  it('shows only fields for the selected reward union', async () => {
    const wrapper = mount(AdminAffiliateTierEditor, { props: { tier: base, coupons: [{ id: 'coupon', name: 'Welcome' }] as never, index: 0, count: 2 } })
    expect(wrapper.findAllComponents({ name: 'InputNumber' })).toHaveLength(1)
    await wrapper.setProps({ tier: { ...base, reward: { kind: 'subscription_extension', extensionDays: 7 } } })
    expect(wrapper.findAllComponents({ name: 'InputNumber' }).length).toBeGreaterThan(1)
  })
})
