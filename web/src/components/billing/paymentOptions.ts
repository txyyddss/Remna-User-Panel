import { t } from '@/i18n'

export interface PaymentProviderOption {
  value: string
  label: string
  description: string
  icon: string
  available: boolean
}

export interface PaymentChannelOption {
  label: string
  value: string
  description: string
  disabled: boolean
  logo?: string
  cryptoCurrency?: 'USDT' | 'USDC'
  network?: string
  networkName?: string
}

export function paymentChannelLogo(provider: string, rail: string): string | undefined {
  const value = `${provider}:${rail}`.toLowerCase()
  if (value.includes('alipay')) return '/assets/payments/alipay.svg'
  if (value.includes('wxpay') || value.includes('wechat')) return '/assets/payments/wechat-pay.svg'
  if (value.includes('usdc')) return paymentCurrencyLogo('USDC')
  if (value.includes('usdt') || value.startsWith('bepusdt:')) return paymentCurrencyLogo('USDT')
  return undefined
}

export function paymentCurrencyLogo(currency: 'USDT' | 'USDC'): string {
  return `/assets/payments/${currency.toLowerCase()}.svg`
}

export function paymentNetworkLogo(network: string, currency: 'USDT' | 'USDC'): string {
  const logos: Record<string, string> = {
    tron: 'network-tron.svg', ethereum: 'network-ethereum.svg', polygon: 'network-polygon.svg',
    bsc: 'network-bsc.svg', aptos: 'network-aptos.svg', solana: 'network-solana.svg',
    'x-layer': 'network-x-layer.svg', xlayer: 'network-x-layer.svg',
    'arbitrum-one': 'network-arbitrum.svg', arbitrum: 'network-arbitrum.svg',
    plasma: 'network-plasma.svg', ton: 'network-ton.svg',
  }
  const logo = logos[network.trim().toLowerCase()]
  return logo ? `/assets/payments/${logo}` : paymentCurrencyLogo(currency)
}

export function paymentChannelLabel(rail: string, fallback: string): string {
  const normalizedRail = rail.trim().toLowerCase()
  const key = normalizedRail === 'wechatpay' ? 'wechat' : normalizedRail
  const translated = t(`payment.channels.${key}`)
  return translated === `payment.channels.${key}` ? fallback : translated
}
