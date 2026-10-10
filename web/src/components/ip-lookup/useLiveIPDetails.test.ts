import { defineComponent, reactive, shallowRef, nextTick } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { IPLookupLiveDetails } from '@/api/ipLookup'

const mocks = vi.hoisted(() => ({ details: vi.fn() }))
vi.mock('@/api/ipLookup', () => ({ ipLookupApi: mocks }))
const account = reactive({ user: { id: 'member' } })
vi.mock('@/stores/session', () => ({ useSessionStore: () => account }))
import { useLiveIPDetails } from './useLiveIPDetails'

const ip = shallowRef('')
const key = shallowRef('')
let state: ReturnType<typeof useLiveIPDetails>
const Harness = defineComponent({ setup() { state = useLiveIPDetails(() => ip.value, () => key.value); return () => null } })

describe('live IP details ownership and request boundaries', () => {
  beforeEach(() => { vi.resetAllMocks(); account.user.id = 'member' })
  it('queries the requested address once per report identity and refreshes without submission', async () => {
    ip.value = '8.8.8.9'
    key.value = 'cached-report:date'
    mocks.details.mockImplementation(async (address: string) => ({ ip: address }))
    const wrapper = mount(Harness)
    await flushPromises()
    expect(mocks.details).toHaveBeenCalledTimes(1)
    expect(mocks.details.mock.calls[0]![0]).toBe('8.8.8.9')
    key.value = 'cached-report:date'
    await nextTick()
    expect(mocks.details).toHaveBeenCalledTimes(1)
    await state.refresh()
    expect(mocks.details).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
  it('aborts old address/account requests and rejects late results', async () => {
    ip.value = '1.1.1.1'
    key.value = 'report'
    const resolvers: Array<(value: IPLookupLiveDetails) => void> = []
    mocks.details.mockImplementation(() => new Promise(resolve => resolvers.push(resolve)))
    const wrapper = mount(Harness)
    const firstSignal = mocks.details.mock.calls[0]![1] as AbortSignal
    ip.value = '8.8.8.8'
    await nextTick()
    expect(firstSignal.aborted).toBe(true)
    resolvers[0]!({ ip: '1.1.1.1' } as IPLookupLiveDetails)
    await flushPromises()
    expect(state.details.value).toBeNull()
    const secondSignal = mocks.details.mock.calls[1]![1] as AbortSignal
    account.user.id = 'replacement'
    await nextTick()
    expect(secondSignal.aborted).toBe(true)
    wrapper.unmount()
    expect((mocks.details.mock.calls[2]![1] as AbortSignal).aborted).toBe(true)
  })
})
