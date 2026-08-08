export function getTelegramWebApp(): TelegramWebApp | undefined {
  return window.Telegram?.WebApp
}

export function isTelegramWebAppDetected(): boolean {
  if (getTelegramWebApp()) return true
  return [window.location.hash, window.location.search].some((source) => new URLSearchParams(source.replace(/^#/, '').replace(/^\?/, '')).has('tgWebAppData'))
}

function readTelegramInitData(app: TelegramWebApp | undefined): string | undefined {
  const direct = app?.initData?.trim()
  if (direct) return direct
  const sources = [window.location.hash, window.location.search]
  for (const source of sources) {
    const params = new URLSearchParams(source.replace(/^#/, '').replace(/^\?/, ''))
    const encoded = params.get('tgWebAppData')
    if (encoded) return encoded
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
