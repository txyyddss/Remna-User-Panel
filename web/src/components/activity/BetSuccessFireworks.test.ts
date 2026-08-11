import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import BetSuccessFireworks from './BetSuccessFireworks.vue'

describe('BetSuccessFireworks', () => {
  it('renders three compact, non-interactive bursts', () => {
    const wrapper = mount(BetSuccessFireworks)

    expect(wrapper.find('.bet-fireworks').attributes('aria-hidden')).toBe('true')
    expect(wrapper.findAll('.bet-fireworks__burst')).toHaveLength(3)
    expect(wrapper.findAll('.bet-fireworks__spark')).toHaveLength(36)
    expect(wrapper.get('.bet-fireworks').classes()).toContain('bet-fireworks')
  })
})
