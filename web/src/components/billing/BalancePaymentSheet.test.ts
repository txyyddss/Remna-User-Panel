import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { afterEach, describe, expect, it } from 'vitest'

import type { FeaturePaymentMethod } from '@/api/features'
import { setLocale } from '@/i18n'

import BalancePaymentConfiguration from './BalancePaymentConfiguration.vue'

const methods: FeaturePaymentMethod[] = [
  { id: 'ezpay:profile-one:alipay', provider: 'ezpay', profileId: 'profile-one', providerName: 'Renminbi payment', rail: 'alipay', name: 'Alipay', currency: 'CNY', cryptoCurrency: '', network: '', networkName: '', available: true, note: '', mode: 'order' },
  { id: 'ezpay:profile-one:wxpay', provider: 'ezpay', profileId: 'profile-one', providerName: 'Renminbi payment', rail: 'wxpay', name: 'WeChat Pay', currency: 'CNY', cryptoCurrency: '', network: '', networkName: '', available: true, note: '', mode: 'order' },
  { id: 'bepusdt:profile-two:usdt.trc20', provider: 'bepusdt', profileId: 'profile-two', providerName: 'USDT Account', rail: 'usdt.trc20', name: 'USDT TRC20', currency: 'USD', cryptoCurrency: 'USDT', network: 'tron', networkName: 'TRC20', available: true, note: '', mode: 'order' },
  { id: 'coupon', provider: 'coupon', profileId: '', rail: '', name: 'Coupon', currency: 'TXB', cryptoCurrency: '', network: '', networkName: '', available: true, note: '', mode: 'coupon_redemption' },
]

describe('BalancePaymentConfiguration', () => {
  afterEach(() => {
    setLocale('en')
    document.body.innerHTML = ''
  })

  it('separates provider selection from channel selection', async () => {
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
    expect(providerTiles).toHaveLength(3)
    expect(providerTiles[0].text()).toContain('Renminbi payment')
    expect(providerTiles[0].text()).not.toContain('Alipay')
    expect(wrapper.findAll('button.provider-option small')).toHaveLength(0)
    expect(providerTiles[0].attributes('aria-pressed')).toBe('true')
    expect(wrapper.find('button[aria-haspopup="listbox"]').exists()).toBe(false)
    expect(wrapper.get('[data-test="choose-channel"]').text()).toContain('Choose channel')
    expect(wrapper.find('[role="radiogroup"]').exists()).toBe(false)

    await wrapper.get('[data-test="choose-channel"]').trigger('click')
    await nextTick()

    const channelPicker = wrapper.get('[role="radiogroup"]')
    expect(channelPicker.text()).toContain('Alipay')
    const paymentSubmit = wrapper.get('[data-test="payment-submit"]')
    expect(paymentSubmit.attributes('class')?.split(/\s+/)).toContain('payment-submit')
    expect(paymentSubmit.text()).toContain('Proceed to payment')
    expect(paymentSubmit.text()).not.toContain('Continue')
  })

  it('keeps the BEPUSDT channel visible when order creation is rejected', async () => {
    const wrapper = mount(BalancePaymentConfiguration, {
      props: {
        amount: '20.00',
        methods,
        selectedMethodId: 'bepusdt:profile-two:usdt.trc20',
        stage: 'creating',
        error: null,
        amountValid: true,
        canCreate: false,
        canReissue: false,
      },
      global: {
        stubs: {
          TxbAmountField: true,
          UAlert: { props: ['description'], template: '<div>{{ description }}</div>' },
          UFormField: { template: '<div><slot /></div>' },
          UInput: true,
        },
      },
    })

    expect(wrapper.get('[data-test="payment-submit"]').text()).toContain('Creating order')
    await wrapper.setProps({
      stage: 'configure',
      error: 'BEPUSDT rejected the order',
      canCreate: true,
      canReissue: true,
    })

    expect(wrapper.get('[data-test="payment-submit"]').text()).toContain('Reissue payment')
    expect(wrapper.text()).toContain('BEPUSDT rejected the order')
    expect(wrapper.find('[data-test="choose-channel"]').exists()).toBe(false)
  })
})
