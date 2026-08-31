import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CommunityMembershipRows from './CommunityMembershipRows.vue'

function mountRows(overrides: Partial<InstanceType<typeof CommunityMembershipRows>['$props']> = {}) {
  return mount(CommunityMembershipRows, {
    props: { activeCombo: false, groupJoined: false, channelJoined: false, joining: [], ...overrides },
    global: {
      stubs: { UBadge: { template: '<span><slot /></span>' }, UIcon: true },
      mocks: { $t: (key: string) => key },
    },
  })
}

describe('CommunityMembershipRows', () => {
  it('keeps confirmed membership ahead of lapsed eligibility', () => {
    const wrapper = mountRows({ groupJoined: true })

    expect(wrapper.text()).toContain('community.joined')
    expect(wrapper.text()).toContain('community.unavailable')
    expect(wrapper.findAll('.community-row__join')).toHaveLength(0)
  })

  it('shows only eligible spaces as join actions and emits the selected space', async () => {
    const wrapper = mountRows({ activeCombo: true })
    const actions = wrapper.findAll('.community-row__join')

    expect(actions).toHaveLength(2)
    expect(actions[0].text()).toContain('community.join')
    await actions[1].trigger('click')
    expect(wrapper.emitted('join')).toEqual([['channel']])
  })
})
