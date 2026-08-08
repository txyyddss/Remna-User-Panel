import type { ComputedRef } from 'vue'
import { onMounted, onUnmounted, watch } from 'vue'

import { getTelegramWebApp } from '@/utils/telegram'

export function useTelegramBackButton(visible: ComputedRef<boolean>, onBack: () => void): void {
  const backButton = getTelegramWebApp()?.BackButton

  onMounted(() => {
    backButton?.onClick(onBack)
  })

  watch(visible, (show) => {
    if (show) backButton?.show()
    else backButton?.hide()
  }, { immediate: true })

  onUnmounted(() => {
    backButton?.offClick(onBack)
    backButton?.hide()
  })
}
