import { afterEach, describe, expect, it, vi } from 'vitest'

import { getTelegramInitData, installHapticClickFeedback, isTelegramUserAgent, isTelegramWebAppDetected, openExternalLink, supportsTelegramVersion, waitForTelegramContext } from './telegram'

const defaultUserAgent = navigator.userAgent

describe('Telegram bootstrap', () => {
  afterEach(() => {
    vi.useRealTimers()
    window.history.replaceState({}, '', '/')
    window.Telegram = undefined
    Object.defineProperty(navigator, 'userAgent', { configurable: true, value: defaultUserAgent })
  })

  it('reads standard WebApp data from the launch hash', async () => {
    window.Telegram = { WebApp: { version: '9.0', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() } }
    window.history.replaceState({}, '', '/#tgWebAppData=query_id%3Ddelayed')
    await expect(getTelegramInitData(100)).resolves.toBe('query_id=delayed')
    expect(isTelegramWebAppDetected()).toBe(true)
  })

  it('preserves an unencoded nested launch query', async () => {
    window.Telegram = { WebApp: { version: '9.0', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() } }
    window.history.replaceState({}, '', '/#tgWebAppData=query_id=abc&auth_date=1&hash=deadbeef&tgWebAppVersion=7.0')
    await expect(getTelegramInitData(100)).resolves.toBe('query_id=abc&auth_date=1&hash=deadbeef')
  })

  it('waits for delayed initData when Telegram has already created WebApp', async () => {
    vi.useFakeTimers()
    const app = { version: '9.0', initData: '', initDataUnsafe: {}, colorScheme: 'dark' as const, ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() }
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

  it('does not treat the SDK placeholder in a regular browser as Telegram', () => {
    window.Telegram = { WebApp: { version: '9.0', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() } }

    expect(isTelegramWebAppDetected()).toBe(false)
  })

  it('recognizes Telegram before initData is available', () => {
    window.Telegram = { WebApp: { version: '9.0', platform: 'ios', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() } }

    expect(isTelegramWebAppDetected()).toBe(true)
  })

  it('recognizes a Telegram user before signed initData is populated', () => {
    window.Telegram = { WebApp: { version: '9.0', initData: '', initDataUnsafe: { user: { id: 42, first_name: 'Mira' } }, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() } }

    expect(isTelegramWebAppDetected()).toBe(true)
  })

  it('waits for delayed Telegram initData before router construction', async () => {
    vi.useFakeTimers()
    const app = { version: '9.0', platform: 'unknown', initData: '', initDataUnsafe: {}, colorScheme: 'dark' as const, ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() }
    window.Telegram = { WebApp: app }

    const pending = waitForTelegramContext(500)
    await vi.advanceTimersByTimeAsync(100)
    app.initData = 'query_id=ready-later'
    await vi.advanceTimersByTimeAsync(50)

    await expect(pending).resolves.toBe(true)
  })

  it('recognizes Telegram launch markers without initData', () => {
    window.Telegram = { WebApp: { version: '9.0', platform: 'unknown', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(), openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn() } }
    window.history.replaceState({}, '', '/#tgWebAppVersion=9.0')

    expect(isTelegramWebAppDetected()).toBe(true)
  })

  it('recognizes Telegram mobile and desktop user agents', () => {
    expect(isTelegramUserAgent('Mozilla/5.0 Telegram/11.0 Mobile')).toBe(true)
    expect(isTelegramUserAgent('Mozilla/5.0 Telegram-Android/11.0 Mobile')).toBe(true)
    expect(isTelegramUserAgent('Mozilla/5.0 Telegram-iOS/11.0 Mobile')).toBe(true)
    expect(isTelegramUserAgent('Mozilla/5.0 TelegramDesktop/5.0')).toBe(true)
    expect(isTelegramUserAgent('Mozilla/5.0 Chrome/140.0 Mobile')).toBe(false)
  })

  it('uses browser navigation rather than unsupported native bridge methods', () => {
    const openLink = vi.fn()
    const browserOpen = vi.spyOn(window, 'open').mockReturnValue(null)
    window.Telegram = { WebApp: {
      version: '6.0', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(),
      openLink, openTelegramLink: vi.fn(), openInvoice: vi.fn(),
    } }

    openExternalLink('https://example.com')

    expect(supportsTelegramVersion('6.1')).toBe(false)
    expect(openLink).not.toHaveBeenCalled()
    expect(browserOpen).toHaveBeenCalledWith('https://example.com', '_blank', 'noopener,noreferrer')
  })

  it('adds marked click feedback and disposes the delegated listener', () => {
    const impactOccurred = vi.fn()
    window.Telegram = { WebApp: {
      version: '9.0', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: vi.fn(), expand: vi.fn(), close: vi.fn(),
      openLink: vi.fn(), openTelegramLink: vi.fn(), openInvoice: vi.fn(),
      HapticFeedback: { impactOccurred, notificationOccurred: vi.fn() },
    } }
    const button = document.createElement('button')
    button.dataset.haptic = 'medium'
    document.body.append(button)
    const dispose = installHapticClickFeedback()

    button.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    expect(impactOccurred).toHaveBeenCalledWith('medium')

    button.setAttribute('disabled', 'true')
    button.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    expect(impactOccurred).toHaveBeenCalledTimes(1)

    dispose()
    button.removeAttribute('disabled')
    button.click()
    expect(impactOccurred).toHaveBeenCalledTimes(1)
  })
})
