import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import SquadNodeBlocks from './SquadNodeBlocks.vue'

const nodes = [
  { uuid: '00000000-0000-4000-8000-000000000001', name: 'Tokyo relay', countryCode: 'JP', consumptionMultiplier: 1.5, providerName: 'Transit provider' },
  { uuid: '00000000-0000-4000-8000-000000000002', name: 'Osaka relay', countryCode: 'JP', consumptionMultiplier: 0.75, providerName: null },
]

describe('SquadNodeBlocks', () => {
  it('shows one anonymous clickable node and switches with desktop controls', async () => {
    const wrapper = mount(SquadNodeBlocks, {
      props: { nodes },
      global: { stubs: {
        Button: { emits: ['click'], template: '<div role="button" tabindex="0" v-bind="$attrs" @click="$emit(\'click\', $event)" />' },
      } },
    })

    expect(wrapper.findAll('.squad-node-carousel__node')).toHaveLength(1)
    expect(wrapper.text()).toContain('1.5x')
    expect(wrapper.text()).not.toContain('Tokyo relay')
    expect(wrapper.text()).not.toContain('Transit provider')
    expect(wrapper.get('.squad-node-carousel__node').attributes('aria-label')).toBe('View Geocheck result for node 1 of 2')

    await wrapper.get('.squad-node-carousel__node').trigger('click')
    expect(wrapper.emitted('openGeocheck')).toEqual([[nodes[0]]])

    await wrapper.get('[aria-label="Next node"]').trigger('click')
    expect(wrapper.text()).toContain('0.75x')
    expect(wrapper.text()).not.toContain('Osaka relay')
    await wrapper.get('.squad-node-carousel__node').trigger('click')
    expect(wrapper.emitted('openGeocheck')?.[1]).toEqual([nodes[1]])
  })
})
