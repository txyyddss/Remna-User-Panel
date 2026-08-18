import { getTelegramWebApp, supportsTelegramVersion, tryTelegramCall } from './telegram'

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
