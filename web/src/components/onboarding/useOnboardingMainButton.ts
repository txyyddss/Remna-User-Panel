import type { ComputedRef } from 'vue'
import { onMounted, onUnmounted, shallowRef, watch } from 'vue'

import { getTelegramWebApp, haptic, supportsTelegramVersion, tryTelegramCall } from '@/utils/telegram'

export interface OnboardingMainAction {
  text: string
  disabled: boolean
  loading: boolean
  run: () => void
}

function readMainButton(): TelegramMainButton | undefined {
  if (!supportsTelegramVersion('6.1')) return undefined
  return getTelegramWebApp()?.MainButton
}

export function useOnboardingMainButton(action: ComputedRef<OnboardingMainAction | null>) {
  const available = shallowRef(false)
  let button: TelegramMainButton | undefined
  let retry: number | undefined
  let attempts = 0

  function handleClick(): void {
    if (!action.value?.disabled) {
      haptic('action')
      action.value?.run()
    }
  }

  function sync(): void {
    const next = readMainButton()
    if (next && next !== button) {
      if (button) tryTelegramCall(() => button?.offClick(handleClick))
      button = next
      if (!tryTelegramCall(() => next.onClick(handleClick))) {
        button = undefined
        return
      }
      available.value = true
    }
    if (!button) return
    const current = action.value
    if (!current) {
      tryTelegramCall(() => button?.hideProgress())
      tryTelegramCall(() => button?.hide())
      return
    }
    tryTelegramCall(() => button?.setText(current.text))
    tryTelegramCall(() => current.disabled ? button?.disable() : button?.enable())
    tryTelegramCall(() => current.loading ? button?.showProgress(true) : button?.hideProgress())
    tryTelegramCall(() => button?.show())
  }

  function bind(): void {
    if (retry !== undefined) window.clearTimeout(retry)
    retry = undefined
    if (!getTelegramWebApp()) {
      if (attempts < 20) {
        attempts += 1
        retry = window.setTimeout(bind, 100)
      }
      return
    }
    sync()
    if (!button && supportsTelegramVersion('6.1') && attempts < 20) {
      attempts += 1
      retry = window.setTimeout(bind, 100)
    }
  }

  onMounted(bind)
  watch(action, sync, { immediate: true })
  onUnmounted(() => {
    if (retry !== undefined) window.clearTimeout(retry)
    if (button) {
      tryTelegramCall(() => button?.offClick(handleClick))
      tryTelegramCall(() => button?.hideProgress())
      tryTelegramCall(() => button?.hide())
    }
  })

  return { available }
}
