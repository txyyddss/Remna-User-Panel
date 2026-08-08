import { computed, readonly, shallowRef } from 'vue'

import { localeMessages, type Locale, type LocaleKey } from './generated'

const storageKey = 'tx-carpool-locale'
const supportedLocales = Object.keys(localeMessages) as Locale[]

function detectLocale(): Locale {
  const stored = globalThis.localStorage?.getItem(storageKey)
  if (stored && supportedLocales.includes(stored as Locale)) return stored as Locale
  const telegramLanguage = window.Telegram?.WebApp?.initDataUnsafe?.user?.language_code?.toLowerCase()
  if (telegramLanguage?.startsWith('zh')) return 'zh-CN'
  return 'en'
}

const locale = shallowRef<Locale>(detectLocale())

function lookup(key: string, source: unknown): unknown {
  return key.split('.').reduce<unknown>((value, segment) => {
    if (!value || typeof value !== 'object') return undefined
    return (value as Record<string, unknown>)[segment]
  }, source)
}

export function setLocale(next: Locale): void {
  if (!supportedLocales.includes(next)) return
  locale.value = next
  globalThis.localStorage?.setItem(storageKey, next)
  document.documentElement.lang = next
}

export function getLocale(): Locale {
  return locale.value
}

export function t(key: LocaleKey | string, variables: Record<string, string | number> = {}): string {
  const value = lookup(key, localeMessages[locale.value]) ?? lookup(key, localeMessages.en)
  if (typeof value !== 'string') return key
  return value.replace(/\{(\w+)\}/g, (_, name: string) => String(variables[name] ?? `{${name}}`))
}

export function localizedError(caught: unknown, fallbackKey = 'common.error'): string {
  if (caught && typeof caught === 'object' && 'code' in caught) {
    const code = String((caught as { code?: unknown }).code ?? '')
    const message = t(`errors.api.${code}`)
    if (message !== `errors.api.${code}`) return message
  }
  if (caught instanceof Error && caught.message) return caught.message
  return t(fallbackKey)
}

export function useI18n() {
  return {
    locale: readonly(locale),
    locales: computed(() => supportedLocales),
    t,
    setLocale,
  }
}
