import { afterEach, describe, expect, it } from 'vitest'

import { convertTXBMoney, displayInputFromTXBMinor, displayInputMinimumFromTXBMinor, displayInputToTXBMinor, formatCatalogMoney, formatMemberMoney, resetDisplayCurrency, setDisplayCurrency } from './displayCurrency'

const money = { currency: 'TXB', minor: '100', display: '1.00 TXB' } as const

describe('member display currency', () => {
  afterEach(() => resetDisplayCurrency())

  it('converts using fixed-point truncation instead of floating point rounding', () => {
    setDisplayCurrency({ currency: 'CNY', rates: { cnyTxbPerUnit: '3', usdTxbPerUnit: null } })

    expect(formatMemberMoney(money)).toBe('0.33 CNY')
    expect(formatCatalogMoney(money)).toBe('1.00 TXB (0.33 CNY)')
    expect(displayInputFromTXBMinor('300')).toBe('1.00')
    expect(displayInputToTXBMinor('1.00')).toBe('300')
    expect(displayInputMinimumFromTXBMinor('100')).toBe('0.34')
  })

  it('preserves signed values and uses USD when configured', () => {
    setDisplayCurrency({ currency: 'USD', rates: { cnyTxbPerUnit: null, usdTxbPerUnit: '2.5' } })

    expect(convertTXBMoney('-1250', 'USD', '2.5')).toBe('-5.00 USD')
    expect(formatMemberMoney({ currency: 'TXB', minor: '1250', display: '12.50 TXB' })).toBe('5.00 USD')
  })

  it('retains the authoritative TXB value when a stored rate is unavailable', () => {
    setDisplayCurrency({ currency: 'CNY', rates: { cnyTxbPerUnit: null, usdTxbPerUnit: null } })

    expect(formatMemberMoney(money)).toBe('1.00 TXB')
  })
})
