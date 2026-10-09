import { shallowMount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { setLocale } from '@/i18n'
import type { UptimeTimeline } from '@/api/subscription'
import UptimeBar from './UptimeBar.vue'
import { h, type FunctionalComponent } from 'vue'

beforeEach(() => { setLocale('en'); vi.useFakeTimers(); vi.setSystemTime('2026-10-09T12:00:00Z') })
afterEach(() => vi.useRealTimers())
const timeline: UptimeTimeline = { from: '2026-10-08T12:00:00Z', to: '2026-10-09T12:00:00Z', state: 'outage', segments: [
  { from: '2026-10-08T12:00:00Z', to: '2026-10-09T11:00:00Z', state: 'operational' },
  { from: '2026-10-09T11:00:00Z', to: '2026-10-09T12:00:00Z', state: 'outage' },
] }
const buttonStub: FunctionalComponent = (props, { slots }) => h('button', props, slots.default?.())
const stubs = { UButton: buttonStub, StatusBadge: { props: ['tone', 'label'], template: '<span :data-tone="tone">{{ label }}</span>' } }

describe('uptime inspection', () => {
  it('lets keyboard users inspect earlier periods without using color alone', async () => {
    const wrapper = shallowMount(UptimeBar, { props: { timeline, label: 'Tokyo transit', host: true }, global: { stubs } })
    const button = wrapper.get('button')
    await button.trigger('focus')
    expect(wrapper.get('[role="status"]').text()).toContain('Complete outage')
    await button.trigger('keydown.left')
    expect(wrapper.get('[role="status"]').text()).toContain('All operational')
    expect(button.attributes('aria-label')).toContain('Tokyo transit')
    wrapper.unmount()
  })
  it('neutralizes the current badge after refresh data becomes stale', async () => {
    const wrapper = shallowMount(UptimeBar, { props: { timeline, label: 'Tokyo transit', host: true }, global: { stubs } })
    expect(wrapper.get('[data-tone]').attributes('data-tone')).toBe('danger')
    await vi.advanceTimersByTimeAsync(105_000)
    expect(wrapper.get('[data-tone]').attributes('data-tone')).toBe('neutral')
    expect(wrapper.text()).toContain('No data')
    wrapper.unmount()
  })
})
