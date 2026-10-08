import { mount } from '@vue/test-utils'
import { defineComponent, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { connectivityApi, type ConnectivityHistory, type ConnectivityAttempt } from '@/api/connectivity'
import { useConnectivityHistory } from './useConnectivityHistory'

vi.mock('@/api/connectivity', () => ({ connectivityApi: { history: vi.fn() } }))
afterEach(() => vi.clearAllMocks())

describe('connectivity history filtering', () => {
  it('discards stale pages after changing hosts and continues with the returned cursor', async () => {
    const pages: Array<(page: ConnectivityHistory) => void> = []
    vi.mocked(connectivityApi.history).mockImplementation(() => new Promise(resolve => { pages.push(resolve) }))
    let history!: ReturnType<typeof useConnectivityHistory>
    const wrapper = mount(defineComponent({ setup() { history = useConnectivityHistory(); return () => null } }))
    const host = '80000000-0000-4000-8000-000000000001'
    history.hostUuid.value = host
    await nextTick()
    const filtered = { id: 'filtered', hostUuid: host } as ConnectivityAttempt
    pages[1]!({ items: [filtered], nextCursor: 'next-page' })
    await Promise.resolve()
    pages[0]!({ items: [{ id: 'stale' } as ConnectivityAttempt], nextCursor: null })
    await Promise.resolve()
    expect(history.items.value.map(item => item.id)).toEqual(['filtered'])
    const more = history.load(true)
    expect(connectivityApi.history).toHaveBeenLastCalledWith(host, 'next-page', expect.any(AbortSignal))
    pages[2]!({ items: [filtered, { id: 'second' } as ConnectivityAttempt], nextCursor: null })
    await more
    expect(history.items.value.map(item => item.id)).toEqual(['filtered', 'second'])
    wrapper.unmount()
  })
})
