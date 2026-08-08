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
    app?.ready()
    const initData = readTelegramInitData(app)
    if (initData) return initData
    await new Promise<void>((resolve) => window.setTimeout(resolve, 50))
  }

  return undefined
}

export function initializeTelegram(): void {
  const app = getTelegramWebApp()
  app?.ready()
  app?.expand()
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
