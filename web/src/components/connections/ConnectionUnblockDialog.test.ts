import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

vi.mock('@/composables/useTelegramBackButton', () => ({ useTelegramBackButton: vi.fn() }))

import ConnectionUnblockDialog from './ConnectionUnblockDialog.vue'

describe('ConnectionUnblockDialog', () => {
  it('shows the exact expiry and emits confirmation', async () => {
    const wrapper = mount(ConnectionUnblockDialog, {
      props: { open: true, block: { id: 'block-1', ip: '2001:db8::1', nodeUuid: 'node', status: 'active',
        createdAt: '2026-08-19T00:00:00Z', expiresAt: '2026-08-22T00:00:00Z' } },
      global: { stubs: { InlineNotice: true, Icon: true,
        Button: { emits: ['click'], template: '<span data-button @click="$emit(\'click\')"><slot /></span>' },
        Modal: { template: '<div><slot name="body" /><slot name="footer" /></div>' } } },
    })

    expect(wrapper.text()).toContain('2001:db8::1')
    await wrapper.findAll('[data-button]').at(-1)?.trigger('click')
    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })
})
