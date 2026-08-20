import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AffiliateTierProgress from './AffiliateTierProgress.vue'

const tier = { id: 'starter', name: 'Starter', threshold: 0, enabled: true, commissionEnabled: false, commissionBps: 0, reward: { kind: 'none' as const } }

describe('AffiliateTierProgress', () => {
  it('renders progress and the top-tier state', async () => {
    const wrapper = mount(AffiliateTierProgress, { props: { progress: { current: tier, next: { ...tier, id: 'next', name: 'Next', threshold: 3 }, successful: 1, remaining: 2, topTier: false } } })
    expect(wrapper.findComponent({ name: 'Progress' }).exists()).toBe(true)
    await wrapper.setProps({ progress: { current: tier, successful: 8, remaining: 0, topTier: true } })
    expect(wrapper.text()).toContain('top tier')
  })
})
