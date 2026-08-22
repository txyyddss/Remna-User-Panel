import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { SquadProduct } from '@/api/types'
import SquadSelector from './SquadSelector.vue'

const fullAddon: SquadProduct = {
  id: 'squad-1',
  remnaSquadUuid: '00000000-0000-4000-8000-000000000001',
  name: 'North transit',
  description: '',
  profile: null,
  price: { currency: 'TXB', minor: '100', display: '1.00 TXB' },
  visible: true,
  upstreamPresent: true,
  activationRequired: false,
  accessibleNodes: [],
  stockRemaining: 0,
  createdAt: '2026-08-14T00:00:00Z',
  updatedAt: '2026-08-14T00:00:00Z',
}

const occupiedAddon: SquadProduct = {
  ...fullAddon,
  id: 'squad-2',
  name: 'Harbor transit',
  stockLimit: 10,
  stockRemaining: 4,
}

const node = {
  uuid: '00000000-0000-4000-8000-000000000003',
  name: 'Harbor relay',
  countryCode: 'SG',
  consumptionMultiplier: 1,
  providerName: 'Transit provider',
}

describe('SquadSelector', () => {
  it('marks a full paid add-on unavailable without affecting included squads', () => {
    const wrapper = mount(SquadSelector, {
      props: { squads: [fullAddon], selectedIds: [], includedIds: [], featuredIds: [], orderedIds: [fullAddon.id] },
    })

    expect(wrapper.find('.squad-option').classes()).toContain('squad-option--full')
    expect(wrapper.text()).toContain('Full')
  })

  it('shows bounded squad occupancy as a whole percentage without exact counts', () => {
    const wrapper = mount(SquadSelector, {
      props: { squads: [occupiedAddon], selectedIds: [], includedIds: [], featuredIds: [], orderedIds: [occupiedAddon.id] },
    })

    expect(wrapper.text()).toContain('60%')
    expect(wrapper.text()).not.toContain('occupied')
    expect(wrapper.text()).not.toContain('6/10')
  })

  it('reserves zero and one hundred percent for empty and full squads', () => {
    const wrapper = mount(SquadSelector, {
      props: {
        squads: [
          { ...occupiedAddon, id: 'squad-low', stockLimit: 1_000, stockRemaining: 999 },
          { ...occupiedAddon, id: 'squad-high', stockLimit: 1_000, stockRemaining: 1 },
        ],
        selectedIds: [],
        includedIds: [],
        featuredIds: [],
        orderedIds: ['squad-low', 'squad-high'],
      },
    })

    expect(wrapper.findAll('.squad-option__occupancy').map(node => node.text()))
      .toEqual(['1%', '99%'])
  })

  it('uses the prepared composition order and opens the exact node without toggling its squad', async () => {
    const squad = { ...occupiedAddon, id: 'featured', name: 'Featured transit', accessibleNodes: [node] }
    const wrapper = mount(SquadSelector, {
      props: { squads: [occupiedAddon, squad], selectedIds: [], includedIds: [], featuredIds: [squad.id], orderedIds: [squad.id, occupiedAddon.id] },
      global: { stubs: {
        Tooltip: { template: '<span><slot /></span>' },
        Button: { emits: ['click'], template: '<div v-bind="$attrs" @click="$emit(\'click\', $event)"><slot /></div>' },
      } },
    })

    const options = wrapper.findAll('.squad-option')
    const featuredTag = options[0]!.get('.squad-option__featured-tag')
    const nameCopy = options[0]!.get('.squad-profile-summary__name-copy')
    expect(options[0]!.classes()).toContain('squad-option--featured')
    expect(wrapper.findAll('.squad-option__featured-tag')).toHaveLength(1)
    expect(featuredTag.text()).toBe('Featured')
    expect(nameCopy.element.children[0]).toBe(featuredTag.element)
    expect(nameCopy.get('strong').text()).toBe('Featured transit')
    expect(options[1]!.find('.squad-option__featured-tag').exists()).toBe(false)
    await wrapper.get('[aria-label="View Geocheck result"]').trigger('click')

    expect(wrapper.emitted('openGeocheck')).toEqual([[node]])
    expect(wrapper.emitted('toggle')).toBeUndefined()
  })
})
