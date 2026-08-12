import { mount } from '@vue/test-utils'
import { nextTick, shallowRef } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeaturePaymentMethod } from '@/api/features'
import { setLocale } from '@/i18n'

const chooseMethod = vi.hoisted(() => vi.fn())

vi.mock('@/composables/usePaymentOrder', () => ({
  usePaymentOrder: () => ({
    amount: shallowRef('20.00'),
    selectedMethodId: shallowRef('ezpay:alipay'),
    stage: shallowRef('configure'),
    order: shallowRef(null),
    qrDataUrl: shallowRef(null),
    error: shallowRef(null),
    canReissue: shallowRef(false),
    amountValid: shallowRef(true),
    canCreate: shallowRef(true),
    reset: vi.fn(),
    hydrateReissueOrder: vi.fn(() => false),
    chooseMethod,
    createOrder: vi.fn(),
    cancelOrder: vi.fn(),
    openPaymentTarget: vi.fn(),
    stopPolling: vi.fn(),
  }),
}))

import BalancePaymentSheet from './BalancePaymentSheet.vue'

const methods: FeaturePaymentMethod[] = [
  { id: 'ezpay:alipay', provider: 'ezpay', rail: 'alipay', name: 'Alipay', currency: 'CNY', available: true, note: '', mode: 'order' },
  { id: 'ezpay:wxpay', provider: 'ezpay', rail: 'wxpay', name: 'WeChat Pay', currency: 'CNY', available: true, note: '', mode: 'order' },
  { id: 'bepusdt:usdt.trc20', provider: 'bepusdt', rail: 'usdt.trc20', name: 'USDT 路 TRC20', currency: 'USDT', available: true, note: '', mode: 'order' },
]

describe('BalancePaymentSheet', () => {
  afterEach(() => {
    setLocale('en')
    vi.clearAllMocks()
    document.body.innerHTML = ''
  })

  it('keeps the EZPay provider label separate from its Alipay channel', async () => {
    const wrapper = mount(BalancePaymentSheet, {
      props: { open: false, methods },
      global: {
        stubs: {
          Alert: true,
          Badge: true,
          Icon: true,
          Modal: { template: '<div><slot name="body" /></div>' },
          Button: { template: '<span v-bind="$attrs"><slot /></span>' },
          TxbAmountField: true,
        },
      },
    })

    await wrapper.setProps({ open: true })
    await nextTick()

    const providerPicker = wrapper.get('.provider-picker')
    const channelPicker = wrapper.get('.channel-picker')
    expect(providerPicker.text()).toContain('EZPay')
    expect(providerPicker.text()).not.toContain('Alipay')
    expect(channelPicker.text()).toContain('Alipay')
  })
})
