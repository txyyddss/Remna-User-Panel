/* eslint-disable vue/one-component-per-file, vue/require-default-prop */

import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { DatabaseMutationReview, DatabaseTable } from '@/api/features'
import DatabaseRecordEditor from './DatabaseRecordEditor.vue'

const table: DatabaseTable = {
  name: 'users',
  columns: [
    { name: 'id', declaredType: 'TEXT', nullable: false, primaryKeyPosition: 1, editable: true, sensitive: false },
    { name: 'role', declaredType: 'TEXT', nullable: false, primaryKeyPosition: 0, editable: true, sensitive: false },
  ],
  highRisk: true,
  supportsRowId: false,
  warning: 'Direct edits bypass hooks.',
}

const review: DatabaseMutationReview = {
  action: 'insert',
  table: 'users',
  key: undefined,
  before: null,
  after: { id: 'user-1', role: 'admin' },
  changedColumns: ['id', 'role'],
  reviewHash: 'review-1',
  requiredConfirmation: 'EDIT users',
  rescueBackupRequired: true,
  warning: 'Review this change.',
}

const DrawerStub = defineComponent({
  props: { title: String, description: String },
  template: '<div class="drawer"><slot name="close" /><h2>{{ title }}</h2><slot name="body" /><footer><slot name="footer" /></footer></div>',
})

const ButtonStub = defineComponent({
  inheritAttrs: false,
  props: { type: String, form: String, label: String, disabled: Boolean },
  emits: ['click'],
  template: '<button v-bind="$attrs" :type="type || \'button\'" :form="form" :disabled="disabled" @click="$emit(\'click\')">{{ label }}</button>',
})

const InputStub = defineComponent({
  inheritAttrs: false,
  props: { modelValue: { type: String, default: '' } },
  emits: ['update:modelValue'],
  template: '<input v-bind="$attrs" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
})

const TextareaStub = defineComponent({
  inheritAttrs: false,
  props: { modelValue: { type: String, default: '' } },
  emits: ['update:modelValue'],
  template: '<textarea v-bind="$attrs" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
})

function mountEditor(reviewState: DatabaseMutationReview | null = null) {
  return mount(DatabaseRecordEditor, {
    props: { table, action: 'insert', review: reviewState, busy: false },
    global: {
      stubs: {
        UDrawer: DrawerStub,
        UButton: ButtonStub,
        UInput: InputStub,
        UTextarea: TextareaStub,
        UFormField: { template: '<div><slot /></div>' },
        UAlert: { template: '<div />' },
        UCheckbox: { template: '<div />' },
        UIcon: { template: '<span />' },
        SwitchField: { template: '<div />' },
      },
    },
  })
}

describe('DatabaseRecordEditor', () => {
  it('submits a review from the sticky-footer form action', async () => {
    const wrapper = mountEditor()
    await wrapper.find('textarea').setValue('Repair imported role')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('review')).toHaveLength(1)
    expect(wrapper.find('footer button[type="submit"]').attributes('form')).toBe('database-record-form')
  })

  it('renders the diff and applies only after the required confirmation', async () => {
    const wrapper = mountEditor(review)

    expect(wrapper.find('.database-diff').exists()).toBe(true)
    expect(wrapper.find('footer button[type="submit"]').attributes('disabled')).toBeDefined()

    await wrapper.find('input[autocomplete="off"]').setValue('EDIT users')
    expect(wrapper.find('footer button[type="submit"]').attributes('disabled')).toBeUndefined()
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('apply')).toHaveLength(1)
    expect(wrapper.emitted('apply')?.[0]?.[0]).toMatchObject({ confirmation: 'EDIT users', reason: '' })
  })
})
