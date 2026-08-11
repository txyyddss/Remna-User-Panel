import { afterEach, describe, expect, it } from 'vitest'

import { createAppHistory } from './history'

describe('application history', () => {
  afterEach(() => {
    window.history.replaceState({}, '', '/')
    window.Telegram = undefined
    sessionStorage.removeItem('txc-route-recovery')
  })

  it('keeps Telegram route changes inside the WebView', () => {
    window.Telegram = { WebApp: { version: '9.0', initData: 'query_id=test', initDataUnsafe: {}, colorScheme: 'dark', ready: () => undefined, expand: () => undefined, close: () => undefined, openLink: () => undefined, openTelegramLink: () => undefined, openInvoice: () => undefined } }

    const history = createAppHistory()
    history.push('/catalog')

    expect(history.location).toBe('/catalog')
    expect(window.location.pathname).toBe('/')
  })

  it('keeps Telegram route changes inside the WebView before delayed initData', () => {
    window.Telegram = { WebApp: { version: '9.0', platform: 'ios', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: () => undefined, expand: () => undefined, close: () => undefined, openLink: () => undefined, openTelegramLink: () => undefined, openInvoice: () => undefined } }
    window.history.replaceState({}, '', '/?tgWebAppVersion=9.0&keep=1#tgWebAppStartParam=home')
    const launchURL = window.location.href

    const history = createAppHistory()
    expect(history.location).toBe('/?keep=1')
    history.push('/catalog')

    expect(window.location.href).toBe(launchURL)
    expect(window.location.pathname).toBe('/')
  })

  it('keeps route changes inside the WebView after SDK captures launch params', () => {
    window.Telegram = {
      WebView: { initParams: { tgWebAppData: 'query_id%3Dcaptured', tgWebAppPlatform: 'ios' } },
      WebApp: { version: '9.0', platform: 'unknown', initData: '', initDataUnsafe: {}, colorScheme: 'dark', ready: () => undefined, expand: () => undefined, close: () => undefined, openLink: () => undefined, openTelegramLink: () => undefined, openInvoice: () => undefined },
    }
    const launchURL = window.location.href

    const history = createAppHistory()
    history.push('/activity')

    expect(history.location).toBe('/activity')
    expect(window.location.href).toBe(launchURL)
  })

  it('keeps browser history shareable outside Telegram', () => {
    const history = createAppHistory()
    history.push('/catalog')

    expect(history.location).toBe('/catalog')
    expect(window.location.pathname).toBe('/catalog')
  })
})
