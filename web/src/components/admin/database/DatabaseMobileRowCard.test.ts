import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { DatabaseRow } from '@/api/features'
import DatabaseMobileRowCard from './DatabaseMobileRowCard.vue'

const ButtonStub = {
  inheritAttrs: false,
  props: { icon: { type: String, default: '' } },
  emits: ['click'],
  template: '<div role="button" :data-test="icon.includes(\'pencil\') ? \'edit-action\' : \'delete-action\'" @click="$emit(\'click\')" />',
}

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
        stubs: { UButton: ButtonStub },
      },
    })

    expect(wrapper.get('code').text()).toBe('{"id":"user-1"}')
    expect(wrapper.text()).not.toContain('hidden')

    await wrapper.get('[data-test="edit-action"]').trigger('click')
    await wrapper.get('[data-test="delete-action"]').trigger('click')

    expect(wrapper.emitted('edit')).toEqual([[row]])
    expect(wrapper.emitted('delete')).toEqual([[row]])
  })
})
