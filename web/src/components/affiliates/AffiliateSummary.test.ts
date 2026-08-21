import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import AffiliateSummary from './AffiliateSummary.vue'

const overview = { inviteLink: 'https://t.me/txbot?start=42', totalCommission: { minor: '1250', currency: 'TXB' as const, display: '12.50 TXB' }, registeredCount: 4, successfulCount: 2, conversionBps: 5000, tierProgress: {} as never }

describe('AffiliateSummary', () => {
  it('copies the server-built link and disables copy when discovery is unavailable', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } })
    const wrapper = mount(AffiliateSummary, { props: { overview } })
    expect(wrapper.get('button').attributes('data-haptic')).toBe('copy')
    await wrapper.get('button').trigger('click')
    expect(writeText).toHaveBeenCalledWith(overview.inviteLink)
    await wrapper.setProps({ overview: { ...overview, inviteLink: undefined } })
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
  })
})
