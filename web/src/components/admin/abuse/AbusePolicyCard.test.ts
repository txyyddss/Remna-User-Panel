import { mount } from '@vue/test-utils'
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
  return mount(AbusePolicyCard, {
    props: { policy, busy: false },
    global: {
      stubs: {
        UForm: { name: 'UForm', props: ['state', 'validate'], emits: ['submit'], template: '<div><slot /></div>' },
        UFormField: { props: ['name', 'label', 'description'], template: '<label>{{ label }}<small>{{ description }}</small><slot /></label>' },
        UInputNumber: { name: 'UInputNumber', props: ['modelValue', 'min', 'max', 'disableWheelChange'], emits: ['update:modelValue'], template: '<span />' },
        USwitch: { props: ['modelValue'], emits: ['update:modelValue'], template: '<span />' },
        UButton: { template: '<span><slot /></span>' },
      },
    },
  })
}

describe('AbusePolicyCard', () => {
  afterEach(() => setLocale('en'))

  it('loads, validates, and submits the streak without dropping policy fields', async () => {
    const wrapper = mountCard()
    const streak = wrapper.findAllComponents({ name: 'UInputNumber' }).find(component => component.props('max') === 1800)
    expect(streak).toBeDefined()
    if (!streak) throw new Error('streak input missing')
    expect(streak.props()).toMatchObject({ modelValue: 30, min: 1, max: 1800, disableWheelChange: true })
    expect(wrapper.text()).toContain('uninterrupted seconds')

    const form = wrapper.findComponent({ name: 'UForm' })
    const validate = form.props('validate') as (value: Partial<AbusePolicy>) => Array<{ name?: string }>
    expect(validate({ ...policy, streakSeconds: 0 })).toContainEqual(expect.objectContaining({ name: 'streakSeconds' }))
    expect(validate({ ...policy, streakSeconds: 1800 })).toEqual([])

    streak.vm.$emit('update:modelValue', 75)
    form.vm.$emit('submit')
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({ ...policy, streakSeconds: 75 })
  })
})
