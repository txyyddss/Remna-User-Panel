const telegramLaunchParameter = /^tgWebApp(?:Data|Version|Platform|ThemeParams|StartParam)$/i
const telegramPlatforms = new Set(['android', 'ios', 'macos', 'tdesktop', 'web', 'weba', 'webk', 'unigram'])

export function isTelegramUserAgent(userAgent: string): boolean {
  return /\bTelegram\b|Telegram(?:[-/]|Android|iOS|Desktop|Web)/i.test(userAgent)
}

function hasTelegramLaunchParameters(): boolean {
  return [window.location.hash, window.location.search].some((source) => {
    const normalized = source.replace(/^#/, '').replace(/^\?/, '')
    return [...new URLSearchParams(normalized).keys()].some((key) => telegramLaunchParameter.test(key))
  })
}

function hasKnownTelegramPlatform(platform: string | undefined): boolean {
  return Boolean(platform && telegramPlatforms.has(platform.toLowerCase()))
}

export function detectTelegramWebApp(app: TelegramWebApp | undefined): boolean {
  if (app?.initData?.trim()) return true
  if (hasTelegramLaunchParameters()) return true
  if (hasKnownTelegramPlatform(app?.platform)) return true
  return typeof navigator !== 'undefined' && isTelegramUserAgent(navigator.userAgent)
}
