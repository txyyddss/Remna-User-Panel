import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
const mocks = vi.hoisted(() => ({ getTrafficResetAutomation: vi.fn(), updateTrafficResetAutomation: vi.fn(), getTag: vi.fn() }))
vi.mock('@/api/memberOperations', () => ({ memberOperationsApi: mocks }))
vi.mock('@/api/preferences', () => ({ preferencesApi: mocks }))
import { useSettings } from './useSettings'

describe('Settings reset automation', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    mocks.getTrafficResetAutomation.mockResolvedValue({ enabled: true, updatedAt: '2026-10-07T00:00:00Z' })
    mocks.getTag.mockResolvedValue({ groupJoined: false, editable: false, tag: '', reasonCode: 'GROUP_NOT_JOINED' })
  })
  it('moves account loading and immediate persistence into Settings', async () => {
    let state!: ReturnType<typeof useSettings>
    const wrapper = mount(defineComponent({ setup() { state = useSettings(); return () => null } }))
    await flushPromises(); expect(state.automation.value?.enabled).toBe(true)
    mocks.updateTrafficResetAutomation.mockResolvedValue({ enabled: false, updatedAt: '2026-10-07T00:01:00Z' })
    await state.setAutomation(false); await state.setAutomation(false)
    expect(mocks.updateTrafficResetAutomation).toHaveBeenCalledExactlyOnceWith(false)
    expect(state.automation.value?.enabled).toBe(false); wrapper.unmount()
  })
  it('preserves the confirmed reset preference when a save fails', async () => {
    let state!: ReturnType<typeof useSettings>
    const wrapper = mount(defineComponent({ setup() { state = useSettings(); return () => null } }))
    await flushPromises(); mocks.updateTrafficResetAutomation.mockRejectedValue(new Error('offline'))
    await state.setAutomation(false)
    expect(state.automation.value?.enabled).toBe(true)
    expect(state.automationError.value).toBeTruthy(); wrapper.unmount()
  })
})
