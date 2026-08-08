import { afterEach, describe, expect, it, vi } from 'vitest'

import { getTelegramInitData, isTelegramWebAppDetected } from './telegram'

describe('Telegram bootstrap', () => {
  afterEach(() => {
    vi.useRealTimers()
    window.history.replaceState({}, '', '/')
    window.Telegram = undefined
  })

  it('reads standard WebApp data from the launch hash', async () => {
    window.Telegram = { WebApp: { initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() } }
    window.history.replaceState({}, '', '/#tgWebAppData=query_id%3Ddelayed')
    await expect(getTelegramInitData(100)).resolves.toBe('query_id=delayed')
    expect(isTelegramWebAppDetected()).toBe(true)
  })

  it('preserves an unencoded nested launch query', async () => {
    window.Telegram = { WebApp: { initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() } }
    window.history.replaceState({}, '', '/#tgWebAppData=query_id=abc&auth_date=1&hash=deadbeef&tgWebAppVersion=7.0')
    await expect(getTelegramInitData(100)).resolves.toBe('query_id=abc&auth_date=1&hash=deadbeef')
  })

  it('waits for delayed initData when Telegram has already created WebApp', async () => {
    vi.useFakeTimers()
    const app = { initData: '', initDataUnsafe: {}, colorScheme: 'dark' as const, ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() }
    window.Telegram = { WebApp: app }
    const pending = getTelegramInitData(500)
    await vi.advanceTimersByTimeAsync(100)
    app.initData = 'query_id=ready-later'
    await vi.advanceTimersByTimeAsync(50)
    await expect(pending).resolves.toBe('query_id=ready-later')
  })

  it('distinguishes an absent Telegram context', () => {
    expect(isTelegramWebAppDetected()).toBe(false)
  })
})
