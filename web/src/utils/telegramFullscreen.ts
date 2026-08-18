import { readonly, shallowRef } from 'vue'

const fullscreen = shallowRef(false)

export function syncTelegramFullscreen(app: TelegramWebApp | undefined): void {
  fullscreen.value = Boolean(app?.isFullscreen)
}

export function telegramFullscreenState() {
  return readonly(fullscreen)
}

export function requestTelegramFullscreen(app: TelegramWebApp | undefined): boolean {
  if (!app?.requestFullscreen) return false
  try {
    app.requestFullscreen()
    return true
  } catch {
    return false
  }
}
