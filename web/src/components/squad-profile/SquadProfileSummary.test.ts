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
})
