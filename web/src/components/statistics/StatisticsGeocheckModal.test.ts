import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import StatisticsGeocheckModal from './StatisticsGeocheckModal.vue'

vi.mock('@/composables/useTelegramBackButton', () => ({ useTelegramBackButton: () => undefined }))

const result = { nodeUuid: '373f14bc-089a-4c3a-91c3-3421e7c83367', checkedAt: '2026-08-19T12:00:00Z', image: { format: 'svg' as const, mediaType: 'image/svg+xml' as const, encoding: 'base64' as const, data: 'PHN2Zy8+' } }

const stubs = {
  UModal: { props: ['open'], template: '<div v-if="open"><slot name="body" /></div>' },
  UButton: { template: '<div><slot /></div>' },
  UTooltip: { template: '<span><slot /></span>' },
  UIcon: { template: '<svg />' },
}

describe('StatisticsGeocheckModal', () => {
  it('renders the server image as a data URI', () => {
    const wrapper = mount(StatisticsGeocheckModal, {
      props: { open: true, node: { uuid: result.nodeUuid, name: 'Tokyo', countryCode: 'JP', online: true, usersOnline: 0, rxBytesPerSec: '0', txBytesPerSec: '0', xrayVersion: '', multiplier: 1 }, result, loading: false, error: null },
      global: { stubs, mocks: { $t: (key: string) => key } },
    })
    expect(wrapper.find('img').attributes('src')).toBe('data:image/svg+xml;base64,PHN2Zy8+')
  })

  it('shows loading and unavailable states', async () => {
    const wrapper = mount(StatisticsGeocheckModal, { props: { open: true, node: null, result: null, loading: true, error: null }, global: { stubs, mocks: { $t: (key: string) => key } } })
    expect(wrapper.text()).toContain('statistics.geocheck.loading')
    await wrapper.setProps({ loading: false })
    expect(wrapper.text()).toContain('statistics.geocheck.unavailable')
  })

  it('resets the zoom when the modal closes', async () => {
    const wrapper = mount(StatisticsGeocheckModal, {
      props: { open: true, node: { uuid: result.nodeUuid, name: 'Tokyo', countryCode: 'JP', online: true, usersOnline: 0, rxBytesPerSec: '0', txBytesPerSec: '0', xrayVersion: '', multiplier: 1 }, result, loading: false, error: null },
      global: { stubs, mocks: { $t: (key: string) => key } },
    })
    await wrapper.find('.statistics-geocheck__canvas').trigger('dblclick')
    expect(wrapper.find('img').attributes('style')).toContain('scale(2)')

    await wrapper.setProps({ open: false })
    await wrapper.setProps({ open: true })
    expect(wrapper.find('img').attributes('style')).toContain('scale(1)')
  })
})
