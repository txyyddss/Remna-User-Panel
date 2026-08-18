import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ConnectionScan } from '@/api/types'

const { pollConnections, requestConnections } = vi.hoisted(() => ({
  pollConnections: vi.fn(),
  requestConnections: vi.fn(),
}))

vi.mock('@/api/memberOperations', () => ({ memberOperationsApi: { pollConnections, requestConnections } }))
vi.mock('@/utils/browserCompatibility', () => ({ createUuid: () => 'scan-key' }))

import { useConnectionScan } from './useConnectionScan'

const pending: ConnectionScan = {
  id: 'scan-1', isCompleted: false, isFailed: false, progressPercent: 25, nodes: [],
  createdAt: '2026-08-18T00:00:00Z', expiresAt: '2026-08-18T00:15:00Z',
}
const completed: ConnectionScan = { ...pending, isCompleted: true, progressPercent: 100 }

function mountScan() {
  let scan!: ReturnType<typeof useConnectionScan>
  const wrapper = mount(defineComponent({
    setup() {
      scan = useConnectionScan(20)
      return () => h('div')
    },
  }))
  return { scan, wrapper }
}

describe('useConnectionScan', () => {
  afterEach(() => {
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  it('polls an accepted scan until its terminal result', async () => {
    vi.useFakeTimers()
    requestConnections.mockResolvedValue(pending)
    pollConnections.mockResolvedValue(completed)
    const { scan, wrapper } = mountScan()
    await flushPromises()

    expect(scan.polling.value).toBe(true)
    await vi.advanceTimersByTimeAsync(20)
    await flushPromises()

    expect(pollConnections).toHaveBeenCalledWith('scan-1')
    expect(scan.completed.value).toBe(true)
    expect(scan.polling.value).toBe(false)
    wrapper.unmount()
  })

  it('reuses the start key after an ambiguous request failure', async () => {
    requestConnections.mockRejectedValueOnce(new Error('network')).mockResolvedValueOnce(completed)
    const { scan, wrapper } = mountScan()
    await flushPromises()

    await scan.restart()

    expect(requestConnections).toHaveBeenNthCalledWith(1, 'scan-key')
    expect(requestConnections).toHaveBeenNthCalledWith(2, 'scan-key')
    expect(scan.completed.value).toBe(true)
    wrapper.unmount()
  })
})
