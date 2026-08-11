import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { DatabaseRow } from '@/api/features'
import DatabaseMobileRowCard from './DatabaseMobileRowCard.vue'

const row: DatabaseRow = {
  key: { id: 'user-1' },
  values: { id: 'user-1', role: 'admin', secret: 'hidden' },
  recordHash: 'hash-1',
}

describe('DatabaseMobileRowCard', () => {
  it('shows only the record key and emits row actions', async () => {
    const wrapper = mount(DatabaseMobileRowCard, {
      props: { row },
      global: {
        stubs: {
          UButton: {
            inheritAttrs: false,
            template: '<button v-bind="$attrs" @click="$emit(\'click\')"><slot /></button>',
          },
        },
      },
    })

    expect(wrapper.get('code').text()).toBe('{"id":"user-1"}')
    expect(wrapper.text()).not.toContain('hidden')

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')

    expect(wrapper.emitted('edit')).toEqual([[row]])
    expect(wrapper.emitted('delete')).toEqual([[row]])
  })
})
