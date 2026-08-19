import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

const api = vi.hoisted(() => ({ listIPBlocks: vi.fn(), dropConnection: vi.fn(), unblockIP: vi.fn(), getOperation: vi.fn() }))
vi.mock('@/api/memberOperations', () => ({ memberOperationsApi: api }))
vi.mock('@/utils/browserCompatibility', () => ({ createUuid: () => 'stable-key' }))

import { useConnectionBlocks } from './useConnectionBlocks'

function mountBlocks() {
  let blocks!: ReturnType<typeof useConnectionBlocks>
  const wrapper = mount(defineComponent({ setup() { blocks = useConnectionBlocks(); return () => h('div') } }))
  return { blocks, wrapper }
}

describe('useConnectionBlocks', () => {
  afterEach(() => { vi.useRealTimers(); vi.clearAllMocks() })

  it('reuses an idempotency key after an ambiguous block failure', async () => {
    api.listIPBlocks.mockResolvedValue({ items: [] })
    api.dropConnection.mockRejectedValueOnce(new Error('network')).mockResolvedValueOnce({ id: 'op', kind: 'connection_block', status: 'failed' })
    const { blocks, wrapper } = mountBlocks()
    await flushPromises()

    await blocks.block('handle')
    await blocks.block('handle')

    expect(api.dropConnection).toHaveBeenNthCalledWith(1, 'handle', 'stable-key')
    expect(api.dropConnection).toHaveBeenNthCalledWith(2, 'handle', 'stable-key')
    wrapper.unmount()
  })

  it('polls an accepted unblock and refreshes the active list', async () => {
    vi.useFakeTimers()
    api.listIPBlocks.mockResolvedValue({ items: [] })
    api.unblockIP.mockResolvedValue({ id: 'op', kind: 'connection_unblock', status: 'queued' })
    api.getOperation.mockResolvedValue({ id: 'op', kind: 'connection_unblock', status: 'succeeded' })
    const { blocks, wrapper } = mountBlocks()
    await flushPromises()

    await blocks.unblock('block-1')
    await vi.advanceTimersByTimeAsync(1500)
    await flushPromises()

    expect(api.unblockIP).toHaveBeenCalledWith('block-1', 'stable-key')
    expect(api.getOperation).toHaveBeenCalledWith('op')
    expect(api.listIPBlocks).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
})
