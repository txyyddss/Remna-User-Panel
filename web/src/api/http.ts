import type { ApiErrorBody } from './types'
import { applyRequestSignature, requestBodyBytes } from './request-signing'

export type QueryValue = string | number | boolean | undefined

export interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown
  query?: Record<string, QueryValue>
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly details?: Record<string, string>
  readonly requestId?: string

  constructor(status: number, body: ApiErrorBody) {
    super(body.message)
    this.name = 'ApiError'
    this.status = status
    this.code = body.code
    this.details = body.details
    this.requestId = body.requestId
  }
}

export function createUrl(path: string, query?: Record<string, QueryValue>): string {
  if (!query) return path
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined) params.set(key, String(value))
  }
  const encoded = params.toString()
  return encoded ? `${path}?${encoded}` : path
}

async function send(path: string, options: RequestOptions): Promise<Response> {
  const { body, query, ...init } = options
  const url = createUrl(path, query)
  const method = (init.method ?? 'GET').toUpperCase()
  const headers = new Headers(init.headers)
  headers.set('Accept', headers.get('Accept') ?? 'application/json')
  const isForm = typeof FormData !== 'undefined' && body instanceof FormData

  if (isForm) {
    const outgoing = new Request(url, { ...init, method, body, credentials: 'include', headers })
    const bytes = new Uint8Array(await outgoing.clone().arrayBuffer())
    await applyRequestSignature(outgoing.headers, method, url, bytes)
    return fetch(outgoing)
  }

  const serialized = body === undefined ? undefined : JSON.stringify(body)
  if (serialized !== undefined) headers.set('Content-Type', 'application/json')
  await applyRequestSignature(headers, method, url, requestBodyBytes(serialized))
  return fetch(url, { ...init, method, body: serialized, credentials: 'include', headers })
}

async function responseError(response: Response): Promise<ApiError> {
  const contentType = response.headers.get('content-type') ?? ''
  const payload = contentType.includes('application/json')
    ? await response.json() as unknown
    : await response.text()
  const body = typeof payload === 'object' && payload !== null
    ? payload as ApiErrorBody
    : { code: 'HTTP_ERROR', message: String(payload || response.statusText) }
  return new ApiError(response.status, body)
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await send(path, options)
  if (!response.ok) throw await responseError(response)
  if (response.status === 204) return undefined as T
  const contentType = response.headers.get('content-type') ?? ''
  const payload = contentType.includes('application/json')
    ? await response.json() as unknown
    : await response.text()
  if (typeof payload === 'object' && payload !== null && 'data' in payload) {
    return (payload as { data: T }).data
  }
  return payload as T
}

export async function requestBlob(path: string, options: RequestOptions = {}): Promise<Blob> {
  const response = await send(path, {
    ...options,
    headers: { Accept: 'application/octet-stream', ...Object.fromEntries(new Headers(options.headers)) },
  })
  if (!response.ok) throw await responseError(response)
  return response.blob()
}
