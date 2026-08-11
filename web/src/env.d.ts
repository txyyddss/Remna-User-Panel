/// <reference types="vite/client" />
import 'vue'

declare global {

interface TelegramWebAppUser {
  id: number
  first_name: string
  last_name?: string
  username?: string
  language_code?: string
}

interface TelegramWebApp {
  version?: string
  platform?: string
  initData: string
  initDataUnsafe: { user?: TelegramWebAppUser }
  colorScheme: 'light' | 'dark'
  themeParams?: Record<string, string | undefined>
  safeAreaInset?: TelegramInsets
  contentSafeAreaInset?: TelegramInsets
  viewportHeight?: number
  viewportStableHeight?: number
  ready(): void
  expand(): void
  close(): void
  openLink(url: string): void
  openTelegramLink(url: string): void
  openInvoice(url: string, callback?: (status: string) => void): void
  isVersionAtLeast?(version: string): boolean
  onEvent?(eventType: string, callback: (...args: unknown[]) => void): void
  offEvent?(eventType: string, callback: (...args: unknown[]) => void): void
  BackButton?: {
    isVisible: boolean
    show(): void
    hide(): void
    onClick(callback: () => void): void
    offClick(callback: () => void): void
  }
  HapticFeedback?: {
    impactOccurred(style: 'light' | 'medium' | 'heavy'): void
    notificationOccurred(type: 'error' | 'success' | 'warning'): void
  }
}

interface TelegramInsets {
  top: number
  right: number
  bottom: number
  left: number
}

interface Window {
  Telegram?: { WebApp: TelegramWebApp }
}
}

declare module 'vue' {
  interface ComponentCustomProperties {
    $t: (key: string, variables?: Record<string, string | number>) => string
  }
}
