export type AffiliateReward =
  | { kind: 'none'; couponId?: never; couponName?: never; txbMinor?: never; extensionDays?: never }
  | { kind: 'coupon'; couponId: string; couponName?: string; txbMinor?: never; extensionDays?: never }
  | { kind: 'txb'; txbMinor: number; couponId?: never; couponName?: never; extensionDays?: never }
  | { kind: 'subscription_extension'; extensionDays: number; couponId?: never; couponName?: never; txbMinor?: never }

export interface AffiliateTier {
  id: string
  name: string
  threshold: number
  enabled: boolean
  commissionEnabled: boolean
  commissionBps: number
  reward: AffiliateReward
}

export interface AffiliateOverview {
  inviteLink?: string
  totalCommission: { minor: string; currency: 'TXB'; display: string }
  registeredCount: number
  successfulCount: number
  conversionBps: number
  tierProgress: { current: AffiliateTier; next?: AffiliateTier; successful: number; remaining: number; topTier: boolean }
}

export interface AffiliateReferral {
  username: string
  registeredAt: string
  status: 'pending' | 'successful'
  paybackAt?: string
  commissionAmount?: { minor: string; currency: 'TXB'; display: string }
}

export interface AffiliateReferralPage {
  items: AffiliateReferral[]
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export interface AdminAffiliateView {
  version: number
  bot: { username?: string; status: 'unresolved' | 'unavailable' | 'ready' | 'stale'; checkedAt?: string }
  tiers: AffiliateTier[]
}
