export function getTelegramWebApp(): TelegramWebApp | undefined {
  return window.Telegram?.WebApp
}

export function isTelegramWebAppDetected(): boolean {
  if (getTelegramWebApp()) return true
  return [window.location.hash, window.location.search].some((source) => new URLSearchParams(source.replace(/^#/, '').replace(/^\?/, '')).has('tgWebAppData'))
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

const telegramEvents = ['themeChanged', 'safeAreaChanged', 'contentSafeAreaChanged', 'viewportChanged']

function syncTelegramEnvironment(): void {
  const app = getTelegramWebApp()
  if (!app) return
  const root = document.documentElement
  root.classList.add('dark')
  for (const [key, value] of Object.entries(app.themeParams ?? {})) {
    if (value) root.style.setProperty(`--tg-theme-${key.replaceAll('_', '-')}`, value)
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
  app?.expand()
  syncTelegramEnvironment()
  for (const event of telegramEvents) app?.onEvent?.(event, syncTelegramEnvironment)
  window.addEventListener('resize', syncTelegramEnvironment)
  return () => {
    for (const event of telegramEvents) app?.offEvent?.(event, syncTelegramEnvironment)
    window.removeEventListener('resize', syncTelegramEnvironment)
  }
}

export function markTelegramReady(): void {
  syncTelegramEnvironment()
  getTelegramWebApp()?.ready()
}

export function openExternalLink(url: string): void {
  const app = getTelegramWebApp()
  if (url.startsWith('https://t.me/') && app) {
    app.openTelegramLink(url)
    return
  }
  if (app) {
    app.openLink(url)
    return
  }
  window.open(url, '_blank', 'noopener,noreferrer')
}

export function haptic(type: 'light' | 'medium' | 'heavy' = 'light'): void {
  getTelegramWebApp()?.HapticFeedback?.impactOccurred(type)
}

export function notifyHaptic(type: 'error' | 'success' | 'warning'): void {
  getTelegramWebApp()?.HapticFeedback?.notificationOccurred(type)
}
