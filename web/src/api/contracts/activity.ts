import type { Money, RFC3339 } from '../types'

export type Reward =
  | { kind: 'txb_delta'; txbDeltaMinor: string }
  | { kind: 'coupon_grant'; couponId: string }
  | { kind: 'subscription_extension'; extensionDays: number }
  | { kind: 'none' }

export interface BetGame {
  id: string
  name: string
  icon: string
  description: string
  winChanceBps: number
  minimumStakeMinor: string
  maximumStakeMinor: string
  returnMultiplierBps: number
  enabled: boolean
}

export interface LuckyDraw {
  id: string
  name: string
  description: string
  feeTxbMinor: string
  enabled: boolean
}

export interface ActivitySettings {
  timezone: string
  dailyRewardMinTxb: string
  dailyRewardMinTxbMinor: string
  dailyRewardMaxTxb: string
  dailyRewardMaxTxbMinor: string
  groupMessageThreshold: number
  groupMessageRewardTxb: string
  groupMessageRewardTxbMinor: string
}

export interface GroupMessageRewardStatus {
  enabled: boolean
  localDate: string
  messageCount: number
  threshold: number
  rewardMinor: string
  rewarded: boolean
  rewardedAt?: RFC3339
}

export interface LuckyDrawPrize {
  id: string
  name: string
  weight: string
  stockRemaining?: number | null
  reward: Reward
}

export interface LuckyDrawAdmin extends LuckyDraw {
  prizes: LuckyDrawPrize[]
  createdAt: RFC3339
  updatedAt: RFC3339
}

export interface LuckyDrawWrite {
  name: string
  description: string
  enabled: boolean
  feeTxbMinor: string
  prizes: LuckyDrawPrize[]
}

export interface ActivityResult {
  id: string
  kind: 'check_in' | 'bet' | 'draw'
  outcome: 'win' | 'loss' | 'complete'
  message: string
  reward: Reward
  stakeTxbMinor?: string
  balanceAfter: Money
  createdAt: RFC3339
}

export interface ActivityOverview {
  balance: Money
  timeZone: string
  checkedInToday: boolean
  dailyRewardMinTxbMinor: string
  dailyRewardMaxTxbMinor: string
  games: BetGame[]
  draws: LuckyDraw[]
  recentResults: ActivityResult[]
  groupMessageReward: GroupMessageRewardStatus
}
