import { mount } from '@vue/test-utils'
import type { ComputedRef } from 'vue'
import { describe, expect, it, vi } from 'vitest'

const { useTelegramBackButton } = vi.hoisted(() => ({ useTelegramBackButton: vi.fn() }))
vi.mock('@/composables/useTelegramBackButton', () => ({ useTelegramBackButton }))

import ConnectionBlockDialog from './ConnectionBlockDialog.vue'

const target = {
  nodeName: 'Tokyo', countryCode: 'JP',
  connection: { ip: '203.0.113.8', lastSeen: '2026-08-18T00:00:00Z', handle: 'signed-handle-value' },
}

describe('ConnectionBlockDialog', () => {
  it('owns native Back and includes shared-IP and three-day warnings', async () => {
    const wrapper = mount(ConnectionBlockDialog, {
      props: { open: true, target },
      global: { stubs: { CountryFlag: true, InlineNotice: { template: '<div><slot /></div>' }, Button: { template: '<div role="button" tabindex="0" v-bind="$attrs"><slot /></div>' },
        Modal: { template: '<div><slot name="body" /><slot name="footer" /></div>' } } },
    })
    const [visible, onBack] = useTelegramBackButton.mock.calls[0] as [ComputedRef<boolean>, () => void]

    expect(visible.value).toBe(true)
    expect(wrapper.text()).toContain('72')
    expect(wrapper.find('[data-haptic="destructive"]').exists()).toBe(true)
    expect(wrapper.find('[data-haptic="dismiss"]').exists()).toBe(true)
    onBack()
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('update:open')).toEqual([[false]])
  })
})
