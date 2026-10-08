import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import type { GroupBoostStatus } from '@/api/features'
import GroupBoostPanel from './GroupBoostPanel.vue'

const openExternalLink = vi.hoisted(() => vi.fn())
vi.mock('@/utils/telegramLinks', () => ({ openExternalLink }))

const ButtonStub = defineComponent({
  props: { label: { type: String, required: true }, disabled: Boolean, loading: Boolean },
  template: '<div role="button" :aria-disabled="disabled">{{ label }}</div>',
})

function panel(boost: GroupBoostStatus, refreshing = false) {
  return mount(GroupBoostPanel, {
    props: { boost, refreshing },
    global: { stubs: { Button: ButtonStub, UButton: ButtonStub } },
  })
}

describe('Group boost information', () => {
  it('shows the minimum multiplier for one active boost', () => {
    const wrapper = panel({ state: 'boosted', count: 1, boostUrl: null })
    expect(wrapper.text()).toContain('Active boosts: 1 · 1× rewards')
  })
  it('explains the requirement and opens the authoritative group link', async () => {
    const url = 'https://t.me/boost?c=123456'
    const wrapper = panel({ state: 'required', count: 0, boostUrl: url })
    expect(wrapper.text()).toContain('at least one active boost')
    expect(wrapper.text()).toContain('One or two boosts give 1×')
    await wrapper.findAll('[role="button"]')[0].trigger('click')
    expect(openExternalLink).toHaveBeenCalledWith(url)
    await wrapper.findAll('[role="button"]')[1].trigger('click')
    expect(wrapper.emitted('refresh')).toHaveLength(1)
  })

  it('distinguishes verification failure from an unboosted user', () => {
    const wrapper = panel({ state: 'unavailable', count: null, boostUrl: null })
    expect(wrapper.text()).toContain('Boost verification unavailable')
    expect(wrapper.findAll('[role="button"]')).toHaveLength(1)
  })

  it('shows the personal count and fractional multiplier', () => {
    const wrapper = panel({ state: 'boosted', count: 3, boostUrl: null })
    expect(wrapper.text()).toContain('Active boosts: 3 · 1.5× rewards')
    expect(wrapper.text()).not.toContain('Boost group')
  })

  it('disables verification while a refresh is pending', () => {
    const wrapper = panel({ state: 'required', count: 0, boostUrl: null }, true)
    expect(wrapper.find('[role="button"]').attributes('aria-disabled')).toBe('true')
  })
})
