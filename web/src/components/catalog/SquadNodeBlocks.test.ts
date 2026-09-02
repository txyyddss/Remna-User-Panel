import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import SquadNodeBlocks from './SquadNodeBlocks.vue'

const nodes = [
  { uuid: '00000000-0000-4000-8000-000000000001', name: 'Tokyo relay', countryCode: 'JP', consumptionMultiplier: 1.5, providerName: 'Transit provider' },
  { uuid: '00000000-0000-4000-8000-000000000002', name: 'Osaka relay', countryCode: 'JP', consumptionMultiplier: 0.75, providerName: null },
]
const manyNodes = [...nodes, ...[3, 4, 5].map((id) => ({ uuid: `00000000-0000-4000-8000-${String(id).padStart(12, '0')}`, name: `Relay ${id}`, countryCode: 'JP', consumptionMultiplier: 1, providerName: null }))]

describe('SquadNodeBlocks', () => {
  it('shows every node by default as an anonymous clickable node', async () => {
    const wrapper = mount(SquadNodeBlocks, {
      props: { nodes },
      global: { stubs: {
        Button: { emits: ['click'], template: '<div role="button" tabindex="0" v-bind="$attrs" @click="$emit(\'click\', $event)"><slot /></div>' },
      } },
    })

    expect(wrapper.findAll('.squad-node-list__node')).toHaveLength(2)
    expect(wrapper.text()).toContain('1.5x')
    expect(wrapper.text()).toContain('0.75x')
    expect(wrapper.text()).not.toContain('Tokyo relay')
    expect(wrapper.text()).not.toContain('Transit provider')
    expect(wrapper.text()).not.toContain('1/2')
    expect(wrapper.get('.squad-node-list__node').attributes('aria-label')).toBe('View Geocheck result for node 1 of 2')

    await wrapper.get('.squad-node-list__node').trigger('click')
    expect(wrapper.emitted('openGeocheck')).toEqual([[nodes[0]]])

    await wrapper.findAll('.squad-node-list__node')[1].trigger('click')
    expect(wrapper.emitted('openGeocheck')?.[1]).toEqual([nodes[1]])
  })

  it('keeps long node lists open but allows them to be folded', async () => {
    const wrapper = mount(SquadNodeBlocks, {
      props: { nodes: manyNodes },
      global: { stubs: {
        Button: { emits: ['click'], template: '<div role="button" tabindex="0" v-bind="$attrs" @click="$emit(\'click\', $event)"><slot /></div>' },
      } },
    })

    expect(wrapper.findAll('.squad-node-list__node')).toHaveLength(5)
    await wrapper.get('.squad-node-list__toggle').trigger('click')
    expect(wrapper.findAll('.squad-node-list__node')).toHaveLength(4)
    await wrapper.get('.squad-node-list__toggle').trigger('click')
    expect(wrapper.findAll('.squad-node-list__node')).toHaveLength(5)
  })
})
