import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
const mocks = vi.hoisted(() => ({ getTrafficResetAutomation: vi.fn(), updateTrafficResetAutomation: vi.fn(), getTag: vi.fn() }))
vi.mock('@/api/memberOperations', () => ({ memberOperationsApi: mocks }))
vi.mock('@/api/preferences', () => ({ preferencesApi: mocks }))
import { useSettings } from './useSettings'

let settingsState!: ReturnType<typeof useSettings>
const SettingsHarness = defineComponent({
  setup() {
    settingsState = useSettings()
    return () => null
  },
})

describe('Settings reset automation', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    mocks.getTrafficResetAutomation.mockResolvedValue({ enabled: true, updatedAt: '2026-10-07T00:00:00Z' })
    mocks.getTag.mockResolvedValue({ groupJoined: false, editable: false, tag: '', reasonCode: 'GROUP_NOT_JOINED' })
  })
  it('moves account loading and immediate persistence into Settings', async () => {
    const wrapper = mount(SettingsHarness)
    await flushPromises(); expect(settingsState.automation.value?.enabled).toBe(true)
    mocks.updateTrafficResetAutomation.mockResolvedValue({ enabled: false, updatedAt: '2026-10-07T00:01:00Z' })
    await settingsState.setAutomation(false); await settingsState.setAutomation(false)
    expect(mocks.updateTrafficResetAutomation).toHaveBeenCalledExactlyOnceWith(false)
    expect(settingsState.automation.value?.enabled).toBe(false); wrapper.unmount()
  })
  it('preserves the confirmed reset preference when a save fails', async () => {
    const wrapper = mount(SettingsHarness)
    await flushPromises(); mocks.updateTrafficResetAutomation.mockRejectedValue(new Error('offline'))
    await settingsState.setAutomation(false)
    expect(settingsState.automation.value?.enabled).toBe(true)
    expect(settingsState.automationError.value).toBeTruthy(); wrapper.unmount()
  })
})
