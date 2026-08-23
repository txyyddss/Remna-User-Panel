import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { DatabaseColumn, DatabaseRow } from '@/api/features'
import DatabaseMobileRowCard from './DatabaseMobileRowCard.vue'

const ButtonStub = {
  inheritAttrs: false,
  props: { icon: { type: String, default: '' }, label: { type: String, default: '' } },
  emits: ['click'],
  template: `<div role="button" :data-test="icon.includes('pencil') ? 'edit-action' : icon.includes('trash') ? 'delete-action' : 'expand-action'" @click="$emit('click')">{{ label }}</div>`,
}

const columns: DatabaseColumn[] = [
  { name: 'tenant_id', declaredType: 'TEXT', nullable: false, primaryKeyPosition: 1, editable: false, sensitive: false },
  { name: 'id', declaredType: 'TEXT', nullable: false, primaryKeyPosition: 2, editable: false, sensitive: false },
  { name: 'display_name', declaredType: 'TEXT', nullable: false, primaryKeyPosition: 0, editable: true, sensitive: false },
  { name: 'note', declaredType: 'TEXT', nullable: true, primaryKeyPosition: 0, editable: true, sensitive: false },
  { name: 'payload', declaredType: 'BLOB', nullable: false, primaryKeyPosition: 0, editable: true, sensitive: false },
  { name: 'status', declaredType: 'TEXT', nullable: false, primaryKeyPosition: 0, editable: true, sensitive: false },
  { name: 'secret', declaredType: 'TEXT', nullable: false, primaryKeyPosition: 0, editable: false, sensitive: true },
]
const row: DatabaseRow = {
  key: { tenant_id: 'team-1', id: 'user-1' },
  values: {
    tenant_id: 'team-1', id: 'user-1', display_name: 'Ada', note: null,
    payload: { blobBase64: 'private-bytes' }, status: 'active', secret: 'never-render-this',
  },
  recordHash: 'hash-1',
}

describe('DatabaseMobileRowCard', () => {
  it('renders structured keys and the first three safe schema fields', () => {
    const wrapper = mount(DatabaseMobileRowCard, {
      props: { row, columns },
      global: { stubs: { Button: ButtonStub }, directives: { autoAnimate: {} } },
    })

    expect(wrapper.findAll('.database-row-card__key-list dt').map((item) => item.text())).toEqual(['tenant_id', 'id'])
    expect(wrapper.findAll('.database-row-card__preview dt').map((item) => item.text())).toEqual(['display_name', 'note', 'payload'])
    expect(wrapper.text()).toContain('Ada')
    expect(wrapper.text()).toContain('NULL')
    expect(wrapper.text()).toContain('[BLOB]')
    expect(wrapper.text()).not.toContain('status')
    expect(wrapper.text()).not.toContain('secret')
    expect(wrapper.text()).not.toContain('never-render-this')
  })

  it('expands safe fields and exposes labeled row actions', async () => {
    const wrapper = mount(DatabaseMobileRowCard, {
      props: { row, columns },
      global: { stubs: { Button: ButtonStub }, directives: { autoAnimate: {} } },
    })

    await wrapper.get('[data-test="expand-action"]').trigger('click')
    expect(wrapper.text()).toContain('status')
    expect(wrapper.text()).not.toContain('secret')
    expect(wrapper.text()).not.toContain('never-render-this')
    expect(wrapper.get('[data-test="edit-action"]').text()).toBe('Edit row')
    expect(wrapper.get('[data-test="delete-action"]').text()).toBe('Delete row')
    await wrapper.get('[data-test="edit-action"]').trigger('click')
    await wrapper.get('[data-test="delete-action"]').trigger('click')

    expect(wrapper.emitted('edit')).toEqual([[row]])
    expect(wrapper.emitted('delete')).toEqual([[row]])
  })
})
