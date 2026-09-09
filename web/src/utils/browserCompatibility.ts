export const browserCapabilityCode = 'BROWSER_CAPABILITY_UNSUPPORTED'
export { copyText } from './browserClipboard'

export class BrowserCapabilityError extends Error {
  readonly code = browserCapabilityCode

  constructor() {
    super(browserCapabilityCode)
    this.name = 'BrowserCapabilityError'
  }
}

function browserCrypto(): Crypto | undefined {
  try {
    return globalThis.crypto
  } catch {
    return undefined
  }
}

export function secureRandomBytes(length: number): Uint8Array {
  const crypto = browserCrypto()
  if (!crypto || typeof crypto.getRandomValues !== 'function') {
    throw new BrowserCapabilityError()
  }

  try {
    return crypto.getRandomValues(new Uint8Array(length))
  } catch {
    throw new BrowserCapabilityError()
  }
}

export function createUuid(): string {
  const crypto = browserCrypto()

  if (typeof crypto?.randomUUID === 'function') {
    try {
      return crypto.randomUUID()
    } catch {
      // Some constrained WebViews expose the method but reject it at runtime.
    }
  }

  const bytes = secureRandomBytes(16)
  bytes[6] = (bytes[6] & 0x0f) | 0x40
  bytes[8] = (bytes[8] & 0x3f) | 0x80
  const value = [...bytes].map((byte) => byte.toString(16).padStart(2, '0')).join('')
  return `${value.slice(0, 8)}-${value.slice(8, 12)}-${value.slice(12, 16)}-${value.slice(16, 20)}-${value.slice(20)}`
}

/**
 * Clone JSON/API editing state without requiring structuredClone.
 * API DTOs used by the panel are JSON-compatible values.
 */
export function cloneForEditing<T>(value: T): T {
  try {
    const clone = globalThis.structuredClone
    if (typeof clone === 'function') return clone(value)
  } catch {
    // Fall through for WebViews that expose structuredClone but reject a value.
  }

  try {
    return JSON.parse(JSON.stringify(value)) as T
  } catch {
    throw new BrowserCapabilityError()
  }
}

function installConstructorFallback(name: string, fallback: unknown): void {
  if (typeof window === 'undefined') return
  const host = window as unknown as Record<string, unknown>
  if (typeof host[name] === 'function') return
  if (typeof fallback !== 'function') return

  try {
    Object.defineProperty(window, name, {
      configurable: true,
      value: fallback,
      writable: true,
    })
  } catch {
    // An immutable host object cannot be made compatible here.
  }
}

export function installBrowserCompatibility(): void {
  if (typeof window === 'undefined') return
  installConstructorFallback('PointerEvent', window.MouseEvent)
  installConstructorFallback('FocusEvent', window.Event)
}

export function supportsNativePointerEvents(): boolean {
  if (typeof window === 'undefined') return false

  return 'onpointerdown' in window
    && typeof window.PointerEvent === 'function'
    && window.PointerEvent !== window.MouseEvent
}

export function mediaQueryList(query: string): MediaQueryList | undefined {
  try {
    return typeof globalThis.matchMedia === 'function'
      ? globalThis.matchMedia(query)
      : undefined
  } catch {
    return undefined
  }
}

export function prefersReducedMotion(): boolean {
  return mediaQueryList('(prefers-reduced-motion: reduce)')?.matches ?? false
}

export function watchMediaQuery(
  queryList: MediaQueryList | undefined,
  listener: () => void,
): () => void {
  if (!queryList) return () => undefined

  if (typeof queryList.addEventListener === 'function') {
    queryList.addEventListener('change', listener)
    return () => queryList.removeEventListener('change', listener)
  }

  queryList.addListener?.(listener)
  return () => queryList.removeListener?.(listener)
}

function supportsCapability(read: () => unknown): boolean {
  try {
    return typeof read() === 'function'
  } catch {
    return false
  }
}

export function missingBrowserCapabilities(): string[] {
  const capabilities: Array<[string, () => unknown]> = [
    ['TextEncoder', () => globalThis.TextEncoder],
    ['URL', () => globalThis.URL],
    ['URLSearchParams', () => globalThis.URLSearchParams],
    ['fetch', () => globalThis.fetch],
    ['Headers', () => globalThis.Headers],
    ['Request', () => globalThis.Request],
    ['FormData', () => globalThis.FormData],
    ['AbortController', () => globalThis.AbortController],
    ['atob', () => globalThis.atob],
    ['btoa', () => globalThis.btoa],
    ['crypto.getRandomValues', () => browserCrypto()?.getRandomValues],
  ]

  return capabilities
    .filter(([, read]) => !supportsCapability(read))
    .map(([name]) => name)
}
