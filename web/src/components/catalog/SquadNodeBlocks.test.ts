import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import SquadNodeBlocks from './SquadNodeBlocks.vue'

describe('SquadNodeBlocks', () => {
  it('renders each node provider and a decimal lowercase-x multiplier', () => {
    const wrapper = mount(SquadNodeBlocks, { props: { nodes: [
      {
        uuid: '00000000-0000-4000-8000-000000000001',
        name: 'Tokyo relay',
        countryCode: 'JP',
        consumptionMultiplier: 1.5,
        providerName: 'Transit provider',
      },
      {
        uuid: '00000000-0000-4000-8000-000000000002',
        name: 'Osaka relay',
        countryCode: 'JP',
        consumptionMultiplier: 0.75,
        providerName: null,
      },
    ] } })

    expect(wrapper.text()).toContain('Transit provider')
    expect(wrapper.text()).toContain('1.5x traffic')
    expect(wrapper.text()).toContain('Unavailable')
  })
})
