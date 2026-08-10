const encoder = new TextEncoder()
const requestKeyCookie = 'txc_request_key'
const keyPattern = /^[A-Za-z0-9_-]{43}$/

function readCookie(name: string): string | undefined {
  const prefix = `${name}=`
  for (const part of document.cookie.split(';')) {
    const value = part.trim()
    if (value.startsWith(prefix)) return decodeURIComponent(value.slice(prefix.length))
  }
  return undefined
}

function decodeBase64Url(value: string): Uint8Array {
  const standard = value.replaceAll('-', '+').replaceAll('_', '/')
  const padded = standard.padEnd(Math.ceil(standard.length / 4) * 4, '=')
  return Uint8Array.from(atob(padded), (character) => character.charCodeAt(0))
}

function encodeBase64Url(value: Uint8Array): string {
  let binary = ''
  for (const byte of value) binary += String.fromCharCode(byte)
  return btoa(binary).replaceAll('+', '-').replaceAll('/', '_').replace(/=+$/, '')
}

function hex(value: ArrayBuffer): string {
  return [...new Uint8Array(value)].map((byte) => byte.toString(16).padStart(2, '0')).join('')
}

function buffer(value: Uint8Array): ArrayBuffer {
  return value.buffer.slice(value.byteOffset, value.byteOffset + value.byteLength) as ArrayBuffer
}

function canonicalTarget(url: string): string {
  const parsed = new URL(url, window.location.origin)
  return `${parsed.pathname}${parsed.search}`
}

function signingRequired(target: string): boolean {
  return target !== '/api/v1/auth/telegram'
}

export async function applyRequestSignature(
  headers: Headers,
  method: string,
  url: string,
  body: Uint8Array,
): Promise<void> {
  const target = canonicalTarget(url)
  if (!signingRequired(target)) return
  const encodedKey = readCookie(requestKeyCookie)
  if (!encodedKey || !keyPattern.test(encodedKey)) return

  const timestamp = Math.floor(Date.now() / 1000).toString()
  const nonceBytes = crypto.getRandomValues(new Uint8Array(16))
  const nonce = encodeBase64Url(nonceBytes)
  const bodyHash = hex(await crypto.subtle.digest('SHA-256', buffer(body)))
  const canonical = [method.toUpperCase(), target, timestamp, nonce, bodyHash].join('\n')
  const key = await crypto.subtle.importKey(
    'raw', buffer(decodeBase64Url(encodedKey)), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'],
  )
  const signature = hex(await crypto.subtle.sign('HMAC', key, buffer(encoder.encode(canonical))))
  headers.set('X-TXC-Timestamp', timestamp)
  headers.set('X-TXC-Nonce', nonce)
  headers.set('X-TXC-Signature', signature)
}

export function requestBodyBytes(body?: string): Uint8Array {
  return encoder.encode(body ?? '')
}
