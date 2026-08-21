import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import TrafficResetAutomationControl from './TrafficResetAutomationControl.vue'

const SwitchStub = {
  props: ['modelValue', 'disabled'],
  emits: ['update:modelValue'],
  template: '<div role="switch" tabindex="0" :aria-checked="String(modelValue)" :aria-disabled="String(disabled)" @click="$emit(\'update:modelValue\', !modelValue)" />',
}

describe('TrafficResetAutomationControl', () => {
  it('stays controlled while emitting an immediate preference update', async () => {
    const wrapper = mount(TrafficResetAutomationControl, {
      props: { enabled: false, loading: false, saving: false, error: null },
      global: { stubs: { Switch: SwitchStub, InlineNotice: true } },
    })

    await wrapper.get('[role="switch"]').trigger('click')

    expect(wrapper.emitted('update')).toEqual([[true]])
    expect(wrapper.get('[role="switch"]').attributes('aria-checked')).toBe('false')
    await wrapper.setProps({ enabled: true })
    expect(wrapper.get('[role="switch"]').attributes('aria-checked')).toBe('true')
  })
})
