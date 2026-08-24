import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'

import type { AbusePolicy } from '@/api/abuse'
import { setLocale } from '@/i18n'

import AbusePolicyCard from './AbusePolicyCard.vue'

const policy: AbusePolicy = {
  globalEnabled: true,
  globalLimit: 120,
  streakSeconds: 30,
  warningValidityDays: 14,
  warningCooldownMinutes: 45,
  revision: 3,
}

function mountCard() {
  return mount(AbusePolicyCard, { props: { policy, busy: false } })
}

describe('AbusePolicyCard', () => {
  afterEach(() => setLocale('en'))

  it('loads, validates, and submits the streak without dropping policy fields', async () => {
    const wrapper = mountCard()
    const streak = wrapper.get<HTMLInputElement>('[data-test="streak-seconds"]')
    expect(streak.element.value).toBe('30')
    expect(streak.attributes()).toMatchObject({ 'aria-valuemin': '1', 'aria-valuemax': '1800' })
    expect(wrapper.text()).toContain('uninterrupted seconds')

    await streak.trigger('wheel', { deltaY: 100 })
    expect(streak.element.value).toBe('30')

    await wrapper.setProps({ policy: { ...policy, streakSeconds: 0 } })
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.emitted('save')).toBeUndefined()
    expect(wrapper.text()).toContain('Enter a whole-number streak from 1 to 1,800 seconds.')

    await wrapper.setProps({ policy: { ...policy, streakSeconds: 75 } })
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({ ...policy, streakSeconds: 75 })
  })
})
