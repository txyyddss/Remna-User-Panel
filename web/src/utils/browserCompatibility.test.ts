import { afterEach, describe, expect, it, vi } from 'vitest'

import { BrowserCapabilityError, createUuid, installBrowserCompatibility, secureRandomBytes } from './browserCompatibility'

const properties = ['PointerEvent', 'FocusEvent']
const originals = new Map(properties.map((name) => [name, Object.getOwnPropertyDescriptor(window, name)]))

afterEach(() => {
  vi.unstubAllGlobals()
  for (const name of properties) {
    const original = originals.get(name)
    if (original) Object.defineProperty(window, name, original)
    else delete (window as unknown as Record<string, unknown>)[name]
  }
})

describe('browser compatibility', () => {
  it('installs event-constructor fallbacks required by Nuxt UI controls', () => {
    Object.defineProperty(window, 'PointerEvent', { configurable: true, value: undefined })
    Object.defineProperty(window, 'FocusEvent', { configurable: true, value: undefined })

    installBrowserCompatibility()

    expect(new window.PointerEvent('pointerdown')).toBeInstanceOf(window.MouseEvent)
    expect(new window.FocusEvent('focus')).toBeInstanceOf(window.Event)
  })

  it('creates RFC 4122 UUIDs without crypto.randomUUID', () => {
    vi.stubGlobal('crypto', {
      getRandomValues: (bytes: Uint8Array) => bytes.fill(0),
    })

    expect(createUuid()).toBe('00000000-0000-4000-8000-000000000000')
  })

  it('fails closed when no secure random generator exists', () => {
    vi.stubGlobal('crypto', {})

    expect(() => secureRandomBytes(16)).toThrow(BrowserCapabilityError)
  })
})
