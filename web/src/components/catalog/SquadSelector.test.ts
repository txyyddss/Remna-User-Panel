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
  stockRemaining: 0,
  createdAt: '2026-08-14T00:00:00Z',
  updatedAt: '2026-08-14T00:00:00Z',
}

describe('SquadSelector', () => {
  it('marks a full paid add-on unavailable without affecting included squads', () => {
    const wrapper = mount(SquadSelector, {
      props: { squads: [fullAddon], selectedIds: [], includedIds: [] },
    })

    expect(wrapper.find('.squad-option').classes()).toContain('squad-option--full')
    expect(wrapper.text()).toContain('Full')
  })
})
