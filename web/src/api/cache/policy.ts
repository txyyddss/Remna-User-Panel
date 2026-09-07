import type { RequestOptions } from '../http'

export function isReadRequest(path: string, method: string): boolean {
  return method === 'GET' || method === 'HEAD' || (method === 'POST' && (
    path === '/api/v1/community/membership/check' ||
    /^\/api\/v1\/purchases(?:\/[^/]+\/addons)?\/quote$/.test(path) ||
    path === '/api/v1/subscription/connections' ||
    /^\/api\/v1\/admin\/database\/tables\/[^/]+\/query$/.test(path)
  ))
}

export function responseCacheKey(url: string, options: RequestOptions = {}): string | null {
  const parsed = new URL(url, 'https://session.invalid')
  const path = parsed.pathname
  const method = (options.method ?? 'GET').toUpperCase()
  if (!path.startsWith('/api/v1/') || !isReadRequest(path, method) || method === 'HEAD') return null
  // Authentication, capabilities, quotes and operation polling always require live responses.
  if (path === '/api/v1/me' || /\/(auth|operations|restores|connections)(\/|$)/.test(path) ||
      /\/(key|refund|traffic-reset|quote)$/.test(path) ||
      path.startsWith('/api/v1/payments/')) return null
  parsed.searchParams.sort()
  const body = options.body === undefined ? '' : JSON.stringify(options.body)
  return `${method} ${path}${parsed.search} ${body}`
}
