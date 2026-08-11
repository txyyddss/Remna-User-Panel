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

  it('keeps browser history shareable outside Telegram', () => {
    const history = createAppHistory()
    history.push('/catalog')

    expect(history.location).toBe('/catalog')
    expect(window.location.pathname).toBe('/catalog')
  })
})
