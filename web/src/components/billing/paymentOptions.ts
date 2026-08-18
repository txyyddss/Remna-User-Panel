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
}

export function paymentChannelLogo(provider: string, rail: string): string | undefined {
  const value = `${provider}:${rail}`.toLowerCase()
  if (value.includes('alipay')) return '/assets/payments/alipay.svg'
  if (value.includes('wxpay') || value.includes('wechat')) return '/assets/payments/wechat-pay.svg'
  if (value.includes('usdt') || value.startsWith('bepusdt:')) return '/assets/payments/usdt.svg'
  return undefined
}
