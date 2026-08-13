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
}
