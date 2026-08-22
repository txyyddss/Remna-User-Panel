import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import SquadNodeBlocks from './SquadNodeBlocks.vue'

describe('SquadNodeBlocks', () => {
  it('renders one Geocheck action per node and emits the exact selected node', async () => {
    const nodes = [
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
    ]
    const wrapper = mount(SquadNodeBlocks, {
      props: { nodes },
      global: { stubs: {
        Tooltip: { template: '<span><slot /></span>' },
        Button: { emits: ['click'], template: '<div v-bind="$attrs" @click="$emit(\'click\', $event)"><slot /></div>' },
      } },
    })

    expect(wrapper.text()).toContain('Transit provider')
    expect(wrapper.text()).toContain('1.5x')
    expect(wrapper.text()).not.toContain('Accessible nodes')
    expect(wrapper.text()).toContain('Unavailable')
    const actions = wrapper.findAll('[aria-label="View Geocheck result"]')
    expect(actions).toHaveLength(nodes.length)
    expect(actions.every(action => action.attributes('data-haptic') === 'open')).toBe(true)

    await actions[1]!.trigger('click')
    expect(wrapper.emitted('openGeocheck')).toEqual([[nodes[1]]])
  })
})
