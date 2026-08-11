import { describe, expect, it, vi } from 'vitest'

import { focusWithoutScrolling } from './dom'

describe('DOM compatibility helpers', () => {
  it('falls back when a WebView rejects focus options', () => {
    const element = document.createElement('main')
    const focus = vi.spyOn(element, 'focus')
      .mockImplementationOnce(() => { throw new TypeError('focus options unsupported') })
      .mockImplementationOnce(() => undefined)

    focusWithoutScrolling(element)

    expect(focus).toHaveBeenNthCalledWith(1, { preventScroll: true })
    expect(focus).toHaveBeenNthCalledWith(2)
  })

  it('swallows focus failure during view teardown', () => {
    const element = document.createElement('main')
    vi.spyOn(element, 'focus').mockImplementation(() => { throw new TypeError('detached') })

    expect(() => focusWithoutScrolling(element)).not.toThrow()
  })
})
