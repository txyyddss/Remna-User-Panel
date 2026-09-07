import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Session } from '../types'

const { request } = vi.hoisted(() => ({ request: vi.fn() }))
vi.mock('../http', () => ({ request }))
vi.mock('@/i18n', () => ({ getLocale: () => 'en' }))
import { preloadSession, stopSessionPreload } from './preload'

const session = (role: 'user' | 'admin') => ({ user: { id: 'one', role, onboardingState: 'complete' } }) as Session
beforeEach(() => { vi.useFakeTimers(); request.mockReset(); request.mockResolvedValue({ items: [] }) })
afterEach(() => { stopSessionPreload(); vi.useRealTimers() })

describe('authenticated preload scheduling', () => {
  it('never requests admin data for a member', async () => {
    preloadSession(session('user'))
    await vi.runAllTimersAsync()
    expect(request.mock.calls.some(([path]) => path.startsWith('/api/v1/admin/'))).toBe(false)
    expect(request.mock.calls.some(([path]) => path === '/api/v1/emby/account')).toBe(true)
  })

  it('preloads admin pages and the first database query, containing partial failure', async () => {
    request.mockImplementation(async (path: string) => {
      if (path === '/api/v1/statistics') throw { status: 503 }
      return { items: path.endsWith('/database/tables') ? [{ name: 'users' }] : [] }
    })
    preloadSession(session('admin'))
    await vi.runAllTimersAsync()
    expect(request.mock.calls.some(([path]) => path === '/api/v1/admin/audit-events')).toBe(true)
    expect(request).toHaveBeenCalledWith('/api/v1/admin/database/tables/users/query', expect.objectContaining({ method: 'POST', body: { filters: [], limit: 50 } }))
  })

  it('limits pending work to three reads and cancels it on logout', async () => {
    const signals: AbortSignal[] = []
    request.mockImplementation((_path, options) => new Promise((_resolve, reject) => {
      signals.push(options.signal)
      options.signal.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')))
    }))
    preloadSession(session('admin'))
    await vi.advanceTimersByTimeAsync(0)
    expect(request).toHaveBeenCalledTimes(3)
    stopSessionPreload()
    await vi.runAllTimersAsync()
    expect(signals.every(signal => signal.aborted)).toBe(true)
    expect(request).toHaveBeenCalledTimes(3)
  })
})
