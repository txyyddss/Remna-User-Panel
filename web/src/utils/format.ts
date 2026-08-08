import type { Money, RFC3339 } from '@/api/types'
import { getLocale, t } from '@/i18n'

const unitSymbols: Record<Money['currency'], string> = {
  TXB: 'TXB',
  CNY: 'CNY ',
  USD: '$',
  XTR: 'XTR',
}

export function formatMoney(money: Money): string {
  if (money.display) return money.display
  const negative = money.minor.startsWith('-')
  const absolute = BigInt(negative ? money.minor.slice(1) : money.minor || '0')
  const scale = money.currency === 'XTR' ? 0 : 2
  const scaleFactor = 10n ** BigInt(scale)
  const whole = absolute / scaleFactor
  const fraction = (absolute % scaleFactor).toString().padStart(scale, '0')
  const numeric = scale > 0 ? `${whole}.${fraction}` : `${whole}`
  const value = negative ? `-${numeric}` : numeric
  return money.currency === 'TXB' || money.currency === 'XTR'
    ? `${value} ${unitSymbols[money.currency]}`
    : `${unitSymbols[money.currency]}${value}`
}

export function formatBytes(raw: string | number): string {
  const bytes = Number(raw)
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 GB'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / 1024 ** index
  return `${value >= 100 ? value.toFixed(0) : value.toFixed(1)} ${units[index]}`
}

export function formatDate(value?: RFC3339): string {
  if (!value) return t('common.notScheduled')
  return new Intl.DateTimeFormat(getLocale(), {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  }).format(new Date(value))
}

export function formatDateTime(value?: RFC3339): string {
  if (!value) return t('common.notAvailable')
  return new Intl.DateTimeFormat(getLocale(), {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

export function moneyFromTxbInput(value: string): string {
  const normalized = value.trim()
  if (!/^\d+(?:\.\d{0,2})?$/.test(normalized)) return ''
  const [whole, fraction = ''] = normalized.split('.')
  return (BigInt(whole) * 100n + BigInt(fraction.padEnd(2, '0'))).toString()
}

export function signedMoneyFromTxbInput(value: string): string {
  const normalized = value.trim()
  if (!/^[+-]?\d+(?:\.\d{0,2})?$/.test(normalized)) return ''
  const negative = normalized.startsWith('-')
  const unsigned = normalized.replace(/^[+-]/, '')
  const minor = moneyFromTxbInput(unsigned)
  if (minor === '') return ''
  return negative && minor !== '0' ? `-${minor}` : minor
}

export function txbInputFromMinor(value: string): string {
  if (!/^-?\d+$/.test(value)) return ''
  const negative = value.startsWith('-')
  const absolute = BigInt(negative ? value.slice(1) : value)
  const whole = absolute / 100n
  const fraction = (absolute % 100n).toString().padStart(2, '0')
  return `${negative ? '-' : ''}${whole}.${fraction}`
}

export function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('') || 'TX'
}
