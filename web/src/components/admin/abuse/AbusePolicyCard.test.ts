/* eslint-disable vue/one-component-per-file */
import { mount } from '@vue/test-utils'
import { computed, defineComponent } from 'vue'
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

const FormStub = defineComponent({
  name: 'AbusePolicyFormStub',
  props: {
    state: { type: Object, required: true },
    validate: { type: Function, required: true },
  },
  emits: ['submit'],
  setup(props) {
    const validate = props.validate as (value: Partial<AbusePolicy>) => Array<{ name?: string }>
    const zeroStreakInvalid = computed(() => validate({ ...(props.state as AbusePolicy), streakSeconds: 0 }).some(error => error.name === 'streakSeconds'))
    const maxStreakValid = computed(() => validate({ ...(props.state as AbusePolicy), streakSeconds: 1800 }).length === 0)
    return { zeroStreakInvalid, maxStreakValid }
  },
  template: '<div data-test="policy-form" :data-zero-streak-invalid="zeroStreakInvalid" :data-max-streak-valid="maxStreakValid" @click="$emit(\'submit\')"><slot /></div>',
})
const InputNumberStub = defineComponent({
  name: 'AbusePolicyInputNumberStub',
  props: {
    modelValue: { type: Number, required: true },
    min: { type: Number, required: true },
    max: { type: Number, required: true },
    disableWheelChange: { type: Boolean, required: true },
  },
  emits: ['update:modelValue'],
  template: '<span :data-model-value="modelValue" :data-min="min" :data-max="max" :data-disable-wheel-change="disableWheelChange" @click.stop="$emit(\'update:modelValue\', 75)" />',
})

function mountCard() {
  return mount(AbusePolicyCard, {
    props: { policy, busy: false },
    global: {
      stubs: {
        UForm: FormStub,
        UFormField: { props: ['name', 'label', 'description'], template: '<label>{{ label }}<small>{{ description }}</small><slot /></label>' },
        UInputNumber: InputNumberStub,
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
    const streak = wrapper.get('[data-test="streak-seconds"]')
    expect(streak.attributes()).toMatchObject({ 'data-model-value': '30', 'data-min': '1', 'data-max': '1800', 'data-disable-wheel-change': 'true' })
    expect(wrapper.text()).toContain('uninterrupted seconds')

    const form = wrapper.get('[data-test="policy-form"]')
    expect(form.attributes('data-zero-streak-invalid')).toBe('true')
    expect(form.attributes('data-max-streak-valid')).toBe('true')

    await streak.trigger('click')
    await form.trigger('click')
    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({ ...policy, streakSeconds: 75 })
  })
})
