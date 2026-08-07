/// <reference types="vite/client" />

interface TelegramWebAppUser {
  id: number
  first_name: string
  last_name?: string
  username?: string
  language_code?: string
}

interface TelegramWebApp {
  initData: string
  initDataUnsafe: { user?: TelegramWebAppUser }
  colorScheme: 'light' | 'dark'
  ready(): void
  expand(): void
  close(): void
  openLink(url: string): void
  openTelegramLink(url: string): void
  openInvoice(url: string, callback?: (status: string) => void): void
  HapticFeedback?: {
    impactOccurred(style: 'light' | 'medium' | 'heavy'): void
    notificationOccurred(type: 'error' | 'success' | 'warning'): void
  }
}

interface Window {
  Telegram?: { WebApp: TelegramWebApp }
}
