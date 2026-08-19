import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AdminUserIPBlocks from './AdminUserIPBlocks.vue'

describe('AdminUserIPBlocks', () => {
  it('offers unblock but no administrator block control', async () => {
    const wrapper = mount(AdminUserIPBlocks, {
      props: { busy: false, items: [{ id: 'block-1', ip: '203.0.113.4', nodeUuid: 'node', status: 'active',
        createdAt: '2026-08-19T00:00:00Z', expiresAt: '2026-08-22T00:00:00Z' }] },
      global: { stubs: { StatusBadge: true, UIcon: true,
        UButton: { template: '<span data-button @click="$emit(\'click\')">unblock</span>' } } },
    })

    expect(wrapper.findAll('[data-button]')).toHaveLength(1)
    await wrapper.find('[data-button]').trigger('click')
    expect(wrapper.emitted('unblock')?.[0]?.[0]).toMatchObject({ id: 'block-1' })
    expect(wrapper.emitted('block')).toBeUndefined()
  })
})
