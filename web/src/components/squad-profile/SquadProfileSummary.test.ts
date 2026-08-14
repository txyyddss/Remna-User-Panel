import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import SquadProfileSummary from './SquadProfileSummary.vue'

describe('SquadProfileSummary', () => {
  it('projects international profile facts and preserves extra Markdown', () => {
    const wrapper = mount(SquadProfileSummary, {
      props: {
        profile: { type: 'international_network', portMbps: null, countryCode: 'SG', upstreamCarriers: ['Carrier A'] },
        description: 'Extra route notes',
      },
    })
    expect(wrapper.text()).toContain('International Network')
    expect(wrapper.text()).toContain('Unlimited')
    expect(wrapper.text()).toContain('Extra route notes')
  })

  it('uses carrier marks without CT, CU, or CM prefixes for China routes', () => {
    const wrapper = mount(SquadProfileSummary, {
      props: {
        profile: { type: 'china_optimized', ct: 'CN2', cu: 'CUG', cm: 'CMI', portMbps: null, countryCode: 'CN' },
        presentation: 'member',
      },
    })

    expect(wrapper.findAll('.carrier-logo')).toHaveLength(3)
    expect(wrapper.find('.carrier-logo--telecom').attributes('aria-label')).toBe('China Telecom')
    expect(wrapper.find('.carrier-logo--unicom').attributes('aria-label')).toBe('China Unicom')
    expect(wrapper.find('.carrier-logo--mobile').attributes('aria-label')).toBe('China Mobile')
    expect(wrapper.text()).not.toMatch(/\b(?:CT|CU|CM)\b/)
  })
})
