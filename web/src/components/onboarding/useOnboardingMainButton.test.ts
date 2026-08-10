import { mount } from '@vue/test-utils'
import { computed, defineComponent, shallowRef } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useOnboardingMainButton } from './useOnboardingMainButton'

describe('useOnboardingMainButton', () => {
  afterEach(() => {
    delete window.Telegram
  })

  it('synchronizes the native action and cleans it up', async () => {
    const button = {
      show: vi.fn(), hide: vi.fn(), enable: vi.fn(), disable: vi.fn(),
      setText: vi.fn(), showProgress: vi.fn(), hideProgress: vi.fn(),
      onClick: vi.fn(), offClick: vi.fn(),
    }
    window.Telegram = { WebApp: {
      initData: '', initDataUnsafe: {}, colorScheme: 'dark',
      ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(),
      openTelegramLink: vi.fn(), openInvoice: vi.fn(), MainButton: button,
    } as TelegramWebApp }
    const text = shallowRef('Continue')
    const run = vi.fn()
    const host = defineComponent({
      setup() {
        const action = computed(() => ({ text: text.value, disabled: false, loading: false, run }))
        return useOnboardingMainButton(action)
      },
      template: '<span>{{ available }}</span>',
    })

    const wrapper = mount(host)
    expect(button.setText).toHaveBeenCalledWith('Continue')
    expect(button.enable).toHaveBeenCalled()
    expect(button.show).toHaveBeenCalled()
    expect(button.onClick).toHaveBeenCalledWith(expect.any(Function))

    const click = button.onClick.mock.calls[0]?.[0]
    click?.()
    expect(run).toHaveBeenCalledOnce()

    text.value = 'Finish'
    await wrapper.vm.$nextTick()
    expect(button.setText).toHaveBeenLastCalledWith('Finish')

    wrapper.unmount()
    expect(button.offClick).toHaveBeenCalledWith(click)
    expect(button.hide).toHaveBeenCalled()
  })
})
