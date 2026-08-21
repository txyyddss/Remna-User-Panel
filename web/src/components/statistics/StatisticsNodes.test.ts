import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { StatisticsNodesSnapshot } from '@/api/types'
import StatisticsNodes from './StatisticsNodes.vue'

const snapshot: StatisticsNodesSnapshot = {
  generatedAt: '2026-08-19T12:00:00Z', stale: false,
  nodes: [{ uuid: '373f14bc-089a-4c3a-91c3-3421e7c83367', name: 'Tokyo', countryCode: 'JP', online: true, usersOnline: 1, rxBytesPerSec: '2', txBytesPerSec: '3', xrayVersion: '1.0', multiplier: 1 }],
}

describe('StatisticsNodes', () => {
  it('emits the selected node from its Geocheck action', async () => {
    const wrapper = mount(StatisticsNodes, {
      props: { snapshot, loading: false },
      global: {
        stubs: {
          CountryFlag: { template: '<span />' },
          Icon: { template: '<svg />' },
          Tooltip: { template: '<span><slot /></span>' },
          Button: { emits: ['click'], template: '<div v-bind="$attrs" @click="$emit(\'click\', $event)"><slot /></div>' },
        },
        mocks: { $t: (key: string) => key },
      },
    })
    const action = wrapper.find('[aria-label="statistics.geocheck.open"]')
    expect(action.attributes('data-haptic')).toBe('open')
    expect(wrapper.text()).not.toContain('statistics.nodeXrayVersion')
    expect(wrapper.text()).not.toContain('statistics.nodeMultiplier')
    await action.trigger('click')
    expect(wrapper.emitted('openGeocheck')).toEqual([[snapshot.nodes[0]]])
  })
})
