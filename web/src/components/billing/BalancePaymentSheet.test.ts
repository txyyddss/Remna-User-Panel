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

  it('keeps the provider account selector separate from its channels', async () => {
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

    const providerPicker = wrapper.get('button[aria-haspopup="listbox"]')
    const channelPicker = wrapper.get('[role="radiogroup"]')
    expect(providerPicker.text()).toContain('Main EZPay')
    expect(providerPicker.text()).not.toContain('Alipay')
    expect(channelPicker.text()).toContain('Alipay')
  })
})
