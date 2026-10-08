import { shallowMount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import type { ConnectivitySnapshot } from '@/api/connectivity'
import { setLocale, t } from '@/i18n'
import { defaultConnectivityConfig } from './config'
import ConnectivityResults from './ConnectivityResults.vue'

beforeEach(() => setLocale('en'))
function render(snapshot: ConnectivitySnapshot) {
  return shallowMount(ConnectivityResults, { props: { snapshot }, global: { stubs: { InlineNotice: { template: '<div><slot /></div>' }, ConnectivityAttemptSummary: true } } })
}
function snapshot(): ConnectivitySnapshot {
  return { config: { ...defaultConnectivityConfig, remnawaveUserId: 41 }, user: null, run: null, stale: true, errorCode: 'CONNECTIVITY_SUBSCRIPTION_UNAVAILABLE', hosts: [] }
}

describe('connectivity results during subscription failure', () => {
  it('keeps restored host identities without inventing an endpoint', () => {
    const hostUuid = '80000000-0000-4000-8000-000000000001'
    const wrapper = render({ ...snapshot(), hosts: [{ hostUuid, remark: '', address: '', port: 0, latest: null }] })
    expect(wrapper.text()).toContain(hostUuid)
    expect(wrapper.find('.connectivity-results__identity small').exists()).toBe(false)
    expect(wrapper.text()).toContain(t('hostConnectivity.stale'))
    wrapper.unmount()
  })

  it('shows the actionable setup failure without a contradictory empty-state prompt', () => {
    const wrapper = render(snapshot())
    expect(wrapper.text()).toContain(t('hostConnectivity.codes.CONNECTIVITY_SUBSCRIPTION_UNAVAILABLE'))
    expect(wrapper.text()).not.toContain(t('hostConnectivity.noHosts'))
    expect(wrapper.find('.connectivity-results__list').exists()).toBe(false)
    wrapper.unmount()
  })
})
