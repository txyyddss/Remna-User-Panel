import { defineComponent, reactive } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({ state: vi.fn(), quote: vi.fn(), submit: vi.fn(), check: vi.fn(), replace: vi.fn() }))
vi.mock('@/api/ipLookup', () => ({ ipLookupApi: mocks }))
vi.mock('vue-router', () => ({ useRoute: () => ({ query: {} }), useRouter: () => ({ replace: mocks.replace }) }))
const account = reactive({ user: { id: 'member' } })
vi.mock('@/stores/session', () => ({ useSessionStore: () => account }))
import { useIPLookup } from './useIPLookup'

let state: ReturnType<typeof useIPLookup>
const Harness = defineComponent({ setup() { state = useIPLookup(); return () => null } })
const quote = { ip: '150.249.241.62', refresh: false, charge: { currency: 'TXB', minor: '250', display: '2.50 TXB' }, useQuota: false, cacheReportId: '', token: 'quote' }

describe('IP Lookup request safety', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.useFakeTimers()
    account.user.id = 'member'
    mocks.state.mockResolvedValue({ enabled: true, lookupFee: quote.charge, refreshFee: quote.charge, allowance: null })
    mocks.quote.mockResolvedValue(quote)
    mocks.replace.mockResolvedValue(undefined)
    mocks.check.mockResolvedValue({ operation: { status: 'succeeded' }, report: null })
  })
  afterEach(() => { vi.useRealTimers() })

  it('does not charge for loading, quoting or invalid addresses', async () => {
    const wrapper = mount(Harness)
    await flushPromises()
    state.ip.value = 'localhost'
    await vi.advanceTimersByTimeAsync(400)
    expect(mocks.quote).not.toHaveBeenCalled()
    expect(mocks.submit).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('reuses the idempotency key after an ambiguous submission response', async () => {
    const wrapper = mount(Harness)
    await flushPromises()
    state.ip.value = quote.ip
    await vi.advanceTimersByTimeAsync(400)
    mocks.submit.mockRejectedValueOnce(new Error('response lost')).mockResolvedValueOnce({ id: 'operation' })
    await state.submit()
    await state.submit()
    expect(mocks.submit).toHaveBeenCalledTimes(2)
    expect(mocks.submit.mock.calls[0]![1]).toBe(mocks.submit.mock.calls[1]![1])
    expect(mocks.replace).toHaveBeenCalledWith({ query: { check: 'operation' } })
    wrapper.unmount()
  })

  it('suppresses duplicate submissions while waiting for acceptance', async () => {
    const wrapper = mount(Harness)
    await flushPromises()
    state.ip.value = quote.ip
    await vi.advanceTimersByTimeAsync(400)
    let accept!: (value: { id: string }) => void
    mocks.submit.mockReturnValue(new Promise(resolve => { accept = resolve }))
    const first = state.submit()
    await state.submit()
    expect(mocks.submit).toHaveBeenCalledTimes(1)
    accept({ id: 'operation' })
    await first
    wrapper.unmount()
  })

  it('stops polling after the owning view is disposed', async () => {
    const wrapper = mount(Harness)
    await flushPromises()
    state.ip.value = quote.ip
    await vi.advanceTimersByTimeAsync(400)
    mocks.submit.mockResolvedValue({ id: 'operation' })
    mocks.check.mockResolvedValue({ operation: { status: 'processing' }, report: null })
    await state.submit()
    expect(mocks.check).toHaveBeenCalledTimes(1)
    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(5000)
    expect(mocks.check).toHaveBeenCalledTimes(1)
  })
})
