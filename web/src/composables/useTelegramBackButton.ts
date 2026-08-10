import type { ComputedRef } from 'vue'
import { onMounted, onUnmounted, watch } from 'vue'

import { getTelegramWebApp } from '@/utils/telegram'

export function useTelegramBackButton(visible: ComputedRef<boolean>, onBack: () => void): void {
  let backButton: TelegramWebApp['BackButton']
  let retry: number | undefined
  let attempts = 0

  function bind(): void {
    if (retry !== undefined) window.clearTimeout(retry)
    retry = undefined
    const next = getTelegramWebApp()?.BackButton
    if (next && next !== backButton) {
      backButton?.offClick(onBack)
      backButton = next
      backButton.onClick(onBack)
      if (visible.value) backButton.show()
      else backButton.hide()
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
    if (show) backButton?.show()
    else backButton?.hide()
  }, { immediate: true })

  onUnmounted(() => {
    if (retry !== undefined) window.clearTimeout(retry)
    backButton?.offClick(onBack)
    backButton?.hide()
  })
}
