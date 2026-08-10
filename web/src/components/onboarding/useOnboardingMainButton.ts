import type { ComputedRef } from 'vue'
import { onMounted, onUnmounted, shallowRef, watch } from 'vue'

import { getTelegramWebApp } from '@/utils/telegram'

interface TelegramMainButton {
  show(): void
  hide(): void
  enable(): void
  disable(): void
  setText(text: string): void
  showProgress(leaveActive?: boolean): void
  hideProgress(): void
  onClick(callback: () => void): void
  offClick(callback: () => void): void
}

export interface OnboardingMainAction {
  text: string
  disabled: boolean
  loading: boolean
  run: () => void
}

function readMainButton(): TelegramMainButton | undefined {
  return (getTelegramWebApp() as TelegramWebApp & { MainButton?: TelegramMainButton } | undefined)?.MainButton
}

export function useOnboardingMainButton(action: ComputedRef<OnboardingMainAction | null>) {
  const available = shallowRef(false)
  let button: TelegramMainButton | undefined
  let retry: number | undefined
  let attempts = 0

  function handleClick(): void {
    if (!action.value?.disabled) action.value?.run()
  }

  function sync(): void {
    const next = readMainButton()
    if (next && next !== button) {
      button?.offClick(handleClick)
      button = next
      button.onClick(handleClick)
      available.value = true
    }
    if (!button) return
    const current = action.value
    if (!current) {
      button.hideProgress()
      button.hide()
      return
    }
    button.setText(current.text)
    if (current.disabled) button.disable()
    else button.enable()
    if (current.loading) button.showProgress(true)
    else button.hideProgress()
    button.show()
  }

  function bind(): void {
    if (retry !== undefined) window.clearTimeout(retry)
    retry = undefined
    sync()
    if (!button && attempts < 20) {
      attempts += 1
      retry = window.setTimeout(bind, 100)
    }
  }

  onMounted(bind)
  watch(action, sync, { immediate: true })
  onUnmounted(() => {
    if (retry !== undefined) window.clearTimeout(retry)
    button?.offClick(handleClick)
    button?.hideProgress()
    button?.hide()
  })

  return { available }
}
