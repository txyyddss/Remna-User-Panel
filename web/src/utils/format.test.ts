import { describe, expect, it } from 'vitest'

import { formatBytes, formatMoney, moneyFromTxbInput, signedMoneyFromTxbInput, txbInputFromMinor } from './format'

describe('format utilities', () => {
  it('uses the server display amount without recalculating prices', () => {
    expect(formatMoney({ currency: 'TXB', minor: '1250', display: '12.50 TXB' })).toBe('12.50 TXB')
  })

  it('parses user TXB input into integer hundredths', () => {
    expect(moneyFromTxbInput('20')).toBe('2000')
    expect(moneyFromTxbInput('20.5')).toBe('2050')
    expect(moneyFromTxbInput('20.005')).toBe('')
    expect(moneyFromTxbInput('-2')).toBe('')
    expect(moneyFromTxbInput('150')).toBe('15000')
    expect(txbInputFromMinor('15000')).toBe('150.00')
  })

  it('parses signed reward amounts into integer hundredths', () => {
    expect(signedMoneyFromTxbInput('-1.25')).toBe('-125')
    expect(signedMoneyFromTxbInput('+2')).toBe('200')
    expect(signedMoneyFromTxbInput('1.234')).toBe('')
  })

  it('formats byte counts without mutating the source value', () => {
    expect(formatBytes('1073741824')).toBe('1.0 GB')
    expect(formatBytes(0)).toBe('0 GB')
  })
})
