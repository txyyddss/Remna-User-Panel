import type { Money, PaymentProvider, RFC3339 } from '../types'

export interface EmbyLibrary {
  id: string
  name: string
}

export interface EmbyRating {
  name: string
  value: number
}

export interface EmbyAccount {
  id: string
  username: string
  status: 'queued' | 'provisioning' | 'active' | 'failed'
  maxParentalRating: number | null
  disabledLibraryIds: string[]
  retryable: boolean
  errorMessage?: string
  updatedAt: RFC3339
}

export interface EmbyOverview {
  configured: boolean
  setupPrice: Money
  ratings: EmbyRating[]
  libraries: EmbyLibrary[]
  account: EmbyAccount | null
}

export interface FeaturePaymentMethod {
  id: string
  provider: PaymentProvider
  rail: string
  name: string
  currency: 'TXB' | 'CNY' | 'USD' | 'USDT' | 'XTR'
  available: boolean
  note: string
  mode: 'order' | 'coupon_redemption'
}

export interface FeaturePaymentOrder {
  id: string
  methodId: string
  provider: PaymentProvider
  providerRail: string
  status: 'creating' | 'pending' | 'paid' | 'cancelled' | 'expired' | 'failed' | 'refunded'
  txb: Money
  payableAmount: string
  payableCurrency: 'CNY' | 'USD' | 'USDT' | 'XTR'
  rateSnapshot: string
  rateDirection: 'txb_per_currency' | 'currency_per_txb'
  paymentUrl: string | null
  qrPayload: string | null
  receivingAddress: string | null
  actualCryptoAmount: string | null
  actualCryptoCurrency: 'USDT' | null
  expiresAt: RFC3339
  paidAt: RFC3339 | null
  refundedAt: RFC3339 | null
  cancelledAt: RFC3339 | null
  cancelReason: string
  providerCancelStatus: '' | 'unsupported' | 'requested' | 'cancelled' | 'failed'
  createdAt: RFC3339
  updatedAt: RFC3339
}
