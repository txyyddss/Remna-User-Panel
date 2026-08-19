import { afterEach, describe, expect, it, vi } from 'vitest'

import { chartSegments, formatShortStatisticDate, safeStatisticBytes } from './statisticsFormat'

describe('statistics formatting', () => {
  afterEach(() => vi.restoreAllMocks())

  it('formats provider date buckets in UTC', () => {
    const originalDateTimeFormat = Intl.DateTimeFormat
    const formatter = vi.spyOn(Intl, 'DateTimeFormat').mockImplementation(
      class {
        private readonly delegate: Intl.DateTimeFormat

        constructor(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions) {
          this.delegate = new originalDateTimeFormat(locales, options)
        }

        format(value?: number | Date): string {
          return this.delegate.format(value)
        }
      } as unknown as typeof Intl.DateTimeFormat,
    )

    expect(formatShortStatisticDate('2026-08-17')).toBe('Mon')
    expect(formatter).toHaveBeenLastCalledWith('en', { weekday: 'short', timeZone: 'UTC' })
  })

  it('drops invalid and non-positive chart facts', () => {
    const segments = chartSegments([
      { id: 'paid', label: 'Paid', value: 3 },
      { id: 'invalid', label: 'Invalid', value: Number.NaN },
      { id: 'empty', label: 'Empty', value: 0 },
    ])

    expect(segments).toHaveLength(1)
    expect(segments[0]?.percentage).toBe(100)
    expect(safeStatisticBytes('-1')).toBe(0n)
  })
})
