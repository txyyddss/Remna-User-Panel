import type { ComputedRef } from 'vue'
import { onMounted, onUnmounted, watch } from 'vue'

import { getTelegramWebApp, haptic, supportsTelegramVersion, tryTelegramCall } from '@/utils/telegram'

type BackAction = () => void | Promise<void>

export function useTelegramBackButton(visible: ComputedRef<boolean>, onBack: BackAction): void {
  let backButton: TelegramWebApp['BackButton']
  let retry: number | undefined
  let attempts = 0

  function handleBack(): void {
    haptic()
    try {
      void Promise.resolve(onBack()).catch(() => undefined)
    } catch {
      // Telegram can dispatch the event while the route component is tearing down.
    }
  }

  function bind(): void {
    if (retry !== undefined) window.clearTimeout(retry)
    retry = undefined
    const app = getTelegramWebApp()
    if (!app) {
      if (attempts < 20) {
        attempts += 1
        retry = window.setTimeout(bind, 100)
      }
      return
    }
    if (!supportsTelegramVersion('6.1')) return
    const next = app.BackButton
    if (next && next !== backButton) {
      if (backButton) tryTelegramCall(() => backButton?.offClick(handleBack))
      backButton = next
      if (!tryTelegramCall(() => next.onClick(handleBack))) {
        backButton = undefined
        return
      }
      tryTelegramCall(() => visible.value ? next.show() : next.hide())
      return
    }
    if (!next && attempts < 20) {
      attempts += 1
      retry = window.setTimeout(bind, 100)
    }
  }

  onMounted(() => {
    bind()
  })

  watch(visible, (show) => {
    bind()
    if (backButton) tryTelegramCall(() => show ? backButton?.show() : backButton?.hide())
  }, { immediate: true })

  onUnmounted(() => {
    if (retry !== undefined) window.clearTimeout(retry)
    if (backButton) {
      tryTelegramCall(() => backButton?.offClick(handleBack))
      tryTelegramCall(() => backButton?.hide())
    }
  })
}
