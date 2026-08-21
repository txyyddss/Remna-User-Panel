import { detectTelegramWebApp } from './telegramContext'
import { installHapticClickFeedback } from './telegramHaptics'
import { requestTelegramFullscreen, syncTelegramFullscreen } from './telegramFullscreen'

export { isTelegramUserAgent, waitForTelegramContext } from './telegramContext'
export { haptic, installHapticClickFeedback, notifyHaptic, notifyBetOutcome, selectionHaptic, type HapticImpact, type HapticIntent } from './telegramHaptics'
export { telegramFullscreenState } from './telegramFullscreen'
export { openExternalLink, openTelegramInvoice } from './telegramLinks'

export function getTelegramWebApp(): TelegramWebApp | undefined {
  return window.Telegram?.WebApp
}

function versionParts(value: string | undefined): number[] | undefined {
  if (!value || !/^\d+(?:\.\d+)*$/.test(value)) return undefined
  return value.split('.').map(Number)
}

export function supportsTelegramVersion(minimum: string): boolean {
  const app = getTelegramWebApp()
  if (!app) return false
  try {
    if (typeof app.isVersionAtLeast === 'function') return app.isVersionAtLeast(minimum)
  } catch {
    return false
  }
  const current = versionParts(app.version)
  const required = versionParts(minimum)
  if (!current || !required) return false
  const length = Math.max(current.length, required.length)
  for (let index = 0; index < length; index += 1) {
    const actual = current[index] ?? 0
    const expected = required[index] ?? 0
    if (actual !== expected) return actual > expected
  }
  return true
}

export function tryTelegramCall(callback: () => void): boolean {
  try {
    callback()
    return true
  } catch {
    return false
  }
}

export function isTelegramWebAppDetected(): boolean {
  return detectTelegramWebApp(getTelegramWebApp())
}

function decodeTelegramData(value: string): string {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

function readTelegramDataFromLocation(source: string): string | undefined {
  const normalized = source.replace(/^#/, '').replace(/^\?/, '')
  const marker = normalized.search(/(?:^|&)tgWebAppData=/)
  if (marker < 0) return undefined
  const start = normalized.indexOf('=', marker) + 1
  let payload = normalized.slice(start)
  const clientParameter = payload.search(/&tgWebApp(?:Version|Platform|ThemeParams|StartParam)=/i)
  if (clientParameter >= 0) payload = payload.slice(0, clientParameter)
  return payload ? decodeTelegramData(payload) : undefined
}

function readTelegramInitData(app: TelegramWebApp | undefined): string | undefined {
  const direct = app?.initData?.trim()
  if (direct) return direct
  const webViewData = window.Telegram?.WebView?.initParams?.tgWebAppData?.trim()
  if (webViewData) return decodeTelegramData(webViewData)
  const sources = [window.location.hash, window.location.search]
  for (const source of sources) {
    const launchData = readTelegramDataFromLocation(source)
    if (launchData) return launchData
  }
  return undefined
}

export async function getTelegramInitData(timeoutMs = 3000): Promise<string | undefined> {
  const startedAt = Date.now()

  while (Date.now() - startedAt < timeoutMs) {
    const app = getTelegramWebApp()
    const initData = readTelegramInitData(app)
    if (initData) return initData
    await new Promise<void>((resolve) => window.setTimeout(resolve, 50))
  }

  return undefined
}

const telegramEvents = ['themeChanged', 'viewportChanged', 'fullscreenChanged', 'fullscreenFailed'] as const
const safeAreaEvents = ['safeAreaChanged', 'contentSafeAreaChanged'] as const

function syncTelegramEnvironment(): void {
  const app = getTelegramWebApp()
  if (!app) return
  const root = document.documentElement
  root.classList.add('dark')
  for (const [key, value] of Object.entries(app.themeParams ?? {})) {
    if (value) root.style.setProperty(`--tg-theme-${key.replace(/_/g, '-')}`, value)
  }
  if (app.colorScheme === 'dark') {
    const theme = app.themeParams ?? {}
    if (theme.bg_color) root.style.setProperty('--canvas', theme.bg_color)
    if (theme.secondary_bg_color) root.style.setProperty('--surface', theme.secondary_bg_color)
    if (theme.text_color) root.style.setProperty('--text', theme.text_color)
    if (theme.hint_color) root.style.setProperty('--text-muted', theme.hint_color)
  }
  const setInsets = (prefix: string, insets?: TelegramInsets) => {
    if (!insets) return
    for (const edge of ['top', 'right', 'bottom', 'left'] as const) {
      root.style.setProperty(`${prefix}-${edge}`, `${insets[edge]}px`)
    }
  }
  setInsets('--tg-safe-area-inset', app.safeAreaInset)
  setInsets('--tg-content-safe-area-inset', app.contentSafeAreaInset)
  root.style.setProperty('--tg-viewport-height', `${app.viewportHeight ?? window.innerHeight}px`)
  root.style.setProperty('--tg-viewport-stable-height', `${app.viewportStableHeight ?? window.innerHeight}px`)
}

function syncTelegramColors(app: TelegramWebApp): void {
  const theme = app.themeParams ?? {}
  const background = theme.bg_color
  const surface = theme.secondary_bg_color
  if (background) {
    tryTelegramCall(() => app.setBackgroundColor?.(background))
    tryTelegramCall(() => app.setHeaderColor?.(background))
  }
  if (surface) tryTelegramCall(() => app.setBottomBarColor?.(surface))
}

export function initializeTelegram(): () => void {
  const disposeHapticClicks = installHapticClickFeedback()
  let app: TelegramWebApp | undefined
  let events: readonly string[] = []
  let retry: number | undefined
  let attempts = 0

  const syncAll = () => {
    const current = getTelegramWebApp()
    if (current) syncTelegramColors(current)
    syncTelegramFullscreen(current)
    syncTelegramEnvironment()
  }

  const unbind = () => {
    if (!app?.offEvent) return
    for (const event of events) tryTelegramCall(() => app?.offEvent?.(event, syncAll))
    events = []
  }

  const bind = () => {
    retry = undefined
    const next = getTelegramWebApp()
    if (!next) {
      if (attempts < 30) {
        attempts += 1
        retry = window.setTimeout(bind, 100)
      }
      return
    }
    if (next !== app) {
      unbind()
      app = next
      events = supportsTelegramVersion('8.0') ? [...telegramEvents, ...safeAreaEvents] : telegramEvents.slice(0, 2)
      for (const event of events) if (app.onEvent) tryTelegramCall(() => app?.onEvent?.(event, syncAll))
      tryTelegramCall(() => app?.expand())
    }
    syncAll()
  }

  bind()
  window.addEventListener('resize', syncTelegramEnvironment)
  return () => {
    if (retry !== undefined) window.clearTimeout(retry)
    unbind()
    window.removeEventListener('resize', syncTelegramEnvironment)
    disposeHapticClicks()
    syncTelegramFullscreen(undefined)
  }
}
export function markTelegramReady(): void {
  syncTelegramEnvironment()
  const app = getTelegramWebApp()
  if (app) {
    tryTelegramCall(() => app.ready())
    syncTelegramFullscreen(app)
    if (supportsTelegramVersion('8.0')) requestTelegramFullscreen(app)
  }
}
