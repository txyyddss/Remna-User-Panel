import { detectTelegramWebApp } from './telegramContext'

export { isTelegramUserAgent, waitForTelegramContext } from './telegramContext'

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

export async function getTelegramInitData(timeoutMs = 8000): Promise<string | undefined> {
  const startedAt = Date.now()

  while (Date.now() - startedAt < timeoutMs) {
    const app = getTelegramWebApp()
    const initData = readTelegramInitData(app)
    if (initData) return initData
    await new Promise<void>((resolve) => window.setTimeout(resolve, 50))
  }

  return undefined
}

const telegramEvents = ['themeChanged', 'viewportChanged']
const safeAreaEvents = ['safeAreaChanged', 'contentSafeAreaChanged']

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

export function initializeTelegram(): () => void {
  const app = getTelegramWebApp()
  const disposeHapticClicks = installHapticClickFeedback()
  if (app) tryTelegramCall(() => app.expand())
  syncTelegramEnvironment()
  const events = supportsTelegramVersion('8.0') ? [...telegramEvents, ...safeAreaEvents] : telegramEvents
  for (const event of events) {
    if (app?.onEvent) tryTelegramCall(() => app.onEvent?.(event, syncTelegramEnvironment))
  }
  window.addEventListener('resize', syncTelegramEnvironment)
  return () => {
    for (const event of events) {
      if (app?.offEvent) tryTelegramCall(() => app.offEvent?.(event, syncTelegramEnvironment))
    }
    window.removeEventListener('resize', syncTelegramEnvironment)
    disposeHapticClicks()
  }
}

export function markTelegramReady(): void {
  syncTelegramEnvironment()
  const app = getTelegramWebApp()
  if (app) tryTelegramCall(() => app.ready())
}

export function openExternalLink(url: string): void {
  const app = getTelegramWebApp()
  if (app && supportsTelegramVersion('6.1')) {
    if (url.startsWith('https://t.me/') && tryTelegramCall(() => app.openTelegramLink(url))) return
    if (tryTelegramCall(() => app.openLink(url))) return
  }
  tryTelegramCall(() => window.open(url, '_blank', 'noopener,noreferrer'))
}

export function openTelegramInvoice(url: string): boolean {
  const app = getTelegramWebApp()
  return Boolean(app && supportsTelegramVersion('6.1') && tryTelegramCall(() => app.openInvoice(url)))
}

export type HapticImpact = 'light' | 'medium' | 'heavy'

export function haptic(type: HapticImpact = 'light'): void {
  const app = getTelegramWebApp()
  if (app?.HapticFeedback && supportsTelegramVersion('6.1')) {
    tryTelegramCall(() => app.HapticFeedback?.impactOccurred(type))
  }
}

function hapticImpactFor(element: Element): HapticImpact {
  const value = element.getAttribute('data-haptic')
  return value === 'medium' || value === 'heavy' ? value : 'light'
}

function handleHapticClick(event: MouseEvent): void {
  if (!(event.target instanceof Element)) return
  const target = event.target.closest('[data-haptic]')
  if (!target || target.hasAttribute('disabled') || target.getAttribute('aria-disabled') === 'true') return
  haptic(hapticImpactFor(target))
}

export function installHapticClickFeedback(): () => void {
  document.addEventListener('click', handleHapticClick, true)
  return () => document.removeEventListener('click', handleHapticClick, true)
}

export function notifyHaptic(type: 'error' | 'success' | 'warning'): void {
  const app = getTelegramWebApp()
  if (app?.HapticFeedback && supportsTelegramVersion('6.1')) {
    tryTelegramCall(() => app.HapticFeedback?.notificationOccurred(type))
  }
}
