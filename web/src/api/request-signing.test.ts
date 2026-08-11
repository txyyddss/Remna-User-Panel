import { afterEach, describe, expect, it, vi } from 'vitest'

import { applyRequestSignature, requestBodyBytes } from './request-signing'

const encoder = new TextEncoder()

function base64Url(bytes: Uint8Array): string {
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

function hex(value: ArrayBuffer): string {
  return [...new Uint8Array(value)].map((byte) => byte.toString(16).padStart(2, '0')).join('')
}

function buffer(value: Uint8Array): ArrayBuffer {
  return value.buffer.slice(value.byteOffset, value.byteOffset + value.byteLength) as ArrayBuffer
}

afterEach(() => {
  document.cookie = 'txc_request_key=; Max-Age=0; Path=/'
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('request signing', () => {
  it('signs the exact method, target, timestamp, nonce, and body digest', async () => {
    const keyBytes = new Uint8Array(32).fill(7)
    document.cookie = `txc_request_key=${base64Url(keyBytes)}; Path=/`
    vi.spyOn(Date, 'now').mockReturnValue(1_786_320_000_000)
    const headers = new Headers()
    const body = requestBodyBytes('{"value":"two"}')

    await applyRequestSignature(headers, 'post', '/api/v1/example?q=one+two', body)

    const timestamp = headers.get('X-TXC-Timestamp') ?? ''
    const nonce = headers.get('X-TXC-Nonce') ?? ''
    const bodyHash = hex(await crypto.subtle.digest('SHA-256', buffer(body)))
    const canonical = ['POST', '/api/v1/example?q=one+two', timestamp, nonce, bodyHash].join('\n')
    const key = await crypto.subtle.importKey('raw', buffer(keyBytes), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
    const expected = hex(await crypto.subtle.sign('HMAC', key, buffer(encoder.encode(canonical))))
    expect(timestamp).toBe('1786320000')
    expect(nonce).toMatch(/^[A-Za-z0-9_-]{22}$/)
    expect(headers.get('X-TXC-Signature')).toBe(expected)
  })

  it('leaves the Telegram authentication bootstrap unsigned', async () => {
    document.cookie = `txc_request_key=${base64Url(new Uint8Array(32))}; Path=/`
    const headers = new Headers()
    await applyRequestSignature(headers, 'POST', '/api/v1/auth/telegram', requestBodyBytes('{}'))
    expect(headers.has('X-TXC-Signature')).toBe(false)
  })

  it('keeps protected requests signed without Web Crypto subtle', async () => {
    const nativeCrypto = globalThis.crypto
    vi.stubGlobal('crypto', { getRandomValues: nativeCrypto.getRandomValues.bind(nativeCrypto) })
    document.cookie = `txc_request_key=${base64Url(new Uint8Array(32).fill(5))}; Path=/`
    const headers = new Headers()

    await applyRequestSignature(headers, 'POST', '/api/v1/example', requestBodyBytes('{}'))

    expect(headers.get('X-TXC-Nonce')).toMatch(/^[A-Za-z0-9_-]{22}$/)
    expect(headers.get('X-TXC-Signature')).toMatch(/^[a-f0-9]{64}$/)
  })

  it('fails closed when secure randomness is unavailable', async () => {
    vi.stubGlobal('crypto', {})
    document.cookie = `txc_request_key=${base64Url(new Uint8Array(32).fill(5))}; Path=/`

    await expect(applyRequestSignature(new Headers(), 'POST', '/api/v1/example', requestBodyBytes('{}')))
      .rejects.toMatchObject({ code: 'BROWSER_CAPABILITY_UNSUPPORTED' })
  })
})
