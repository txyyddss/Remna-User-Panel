import { onMounted, onScopeDispose } from 'vue'
import { getTelegramWebApp, tryTelegramCall } from '@/utils/telegram'

export function useTelegramSettingsButton(navigate: () => void): void {
  let button: TelegramWebApp['SettingsButton']
  let timer: ReturnType<typeof setTimeout> | undefined
  let disposed = false
  let attempts = 0
  function connect(): void {
    if (disposed) return
    button = getTelegramWebApp()?.SettingsButton
    if (button) {
      tryTelegramCall(() => { button?.onClick(navigate); button?.show() })
    } else if (attempts++ < 30) timer = setTimeout(connect, 100)
  }
  onMounted(connect)
  onScopeDispose(() => {
    disposed = true
    if (timer) clearTimeout(timer)
    if (button) tryTelegramCall(() => { button?.offClick(navigate); button?.hide() })
  })
}
