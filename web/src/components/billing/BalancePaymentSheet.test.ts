import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { afterEach, describe, expect, it } from 'vitest'

import type { FeaturePaymentMethod } from '@/api/features'
import { setLocale } from '@/i18n'

import BalancePaymentConfiguration from './BalancePaymentConfiguration.vue'

const methods: FeaturePaymentMethod[] = [
  { id: 'ezpay:profile-one:alipay', provider: 'ezpay', profileId: 'profile-one', providerName: 'Main EZPay', rail: 'alipay', name: 'Alipay', currency: 'CNY', available: true, note: '', mode: 'order' },
  { id: 'ezpay:profile-one:wxpay', provider: 'ezpay', profileId: 'profile-one', providerName: 'Main EZPay', rail: 'wxpay', name: 'WeChat Pay', currency: 'CNY', available: true, note: '', mode: 'order' },
  { id: 'bepusdt:profile-two:usdt.trc20', provider: 'bepusdt', profileId: 'profile-two', providerName: 'USDT Account', rail: 'usdt.trc20', name: 'USDT TRC20', currency: 'USDT', available: true, note: '', mode: 'order' },
]

describe('BalancePaymentConfiguration', () => {
  afterEach(() => {
    setLocale('en')
    document.body.innerHTML = ''
  })

  it('renders provider accounts as tiles separate from their channels', async () => {
    const wrapper = mount(BalancePaymentConfiguration, {
      props: {
        amount: '20.00',
        methods,
        selectedMethodId: 'ezpay:profile-one:alipay',
        stage: 'configure',
        error: null,
        amountValid: true,
        canCreate: true,
        canReissue: false,
      },
      global: {
        stubs: {
          TxbAmountField: true,
          UAlert: { template: '<div><slot /></div>' },
          UFormField: { template: '<div><slot /></div>' },
          UInput: true,
        },
      },
    })

    await nextTick()

    const providerTiles = wrapper.findAll('button.provider-option')
    const channelPicker = wrapper.get('[role="radiogroup"]')
    expect(providerTiles).toHaveLength(2)
    expect(providerTiles[0].text()).toContain('Main EZPay')
    expect(providerTiles[0].text()).not.toContain('Alipay')
    expect(providerTiles[0].attributes('aria-pressed')).toBe('true')
    expect(wrapper.find('button[aria-haspopup="listbox"]').exists()).toBe(false)
    expect(channelPicker.text()).toContain('Alipay')
    expect(wrapper.get('[data-test="payment-submit"]').text()).toContain('Proceed to payment')
    expect(wrapper.get('[data-test="payment-submit"]').text()).not.toContain('Continue')
  })
})
