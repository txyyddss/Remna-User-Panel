import { mount } from '@vue/test-utils'
import type { ComputedRef } from 'vue'
import { describe, expect, it, vi } from 'vitest'

const { useTelegramBackButton } = vi.hoisted(() => ({ useTelegramBackButton: vi.fn() }))
vi.mock('@/composables/useTelegramBackButton', () => ({ useTelegramBackButton }))

import ConnectionDropDialog from './ConnectionDropDialog.vue'

const target = {
  nodeName: 'Tokyo', countryCode: 'JP',
  connection: { ip: '203.0.113.8', lastSeen: '2026-08-18T00:00:00Z', handle: 'signed-handle-value' },
}

describe('ConnectionDropDialog', () => {
  it('owns native Back while open and closes through its model', async () => {
    const wrapper = mount(ConnectionDropDialog, {
      props: { open: true, target },
      global: {
        stubs: {
          CountryFlag: true,
          InlineNotice: { template: '<div><slot /></div>' },
          UButton: true,
          UModal: { template: '<div><slot name="body" /><slot name="footer" /></div>' },
        },
      },
    })
    const [visible, onBack] = useTelegramBackButton.mock.calls[0] as [ComputedRef<boolean>, () => void]

    expect(visible.value).toBe(true)
    onBack()
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('update:open')).toEqual([[false]])
  })
})
