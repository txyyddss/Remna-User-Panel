import type { ActivityResult, LuckyDrawPrizePreview } from '@/api/features'

export const drawStyles = ['simple', 'wheel', 'grid', 'cards', 'gift', 'slot', 'scratch'] as const
export const randomDrawStyles = drawStyles.filter((style) => style !== 'simple')
export type DrawStyle = (typeof drawStyles)[number]
export type DrawStyleChoice = DrawStyle | 'random'
export type DrawPhase = 'running' | 'settling' | 'receipt' | 'error'

export interface DrawPresenterProps {
  prizes: readonly LuckyDrawPrizePreview[]
  result: ActivityResult | null
  revealRequested?: number
}

export function resolveDrawStyle(choice: DrawStyleChoice, random: () => number = Math.random): DrawStyle {
  if (choice !== 'random') return choice
  const index = Math.min(randomDrawStyles.length - 1, Math.max(0, Math.floor(random() * randomDrawStyles.length)))
  return randomDrawStyles[index]!
}

export function previewPrizes(
  prizes: readonly LuckyDrawPrizePreview[],
  offset: number,
  result: ActivityResult | null,
): LuckyDrawPrizePreview[] {
  const count = Math.min(prizes.length, 8)
  const visible = Array.from({ length: count }, (_, index) => prizes[(offset + index) % prizes.length]!)
  if (!result?.prizeId || !result.prizeName) return visible
  const selected = { id: result.prizeId, name: result.prizeName }
  const existing = visible.findIndex((prize) => prize.id === selected.id)
  if (existing >= 0) {
    visible[existing] = selected
  } else if (visible.length < 8) {
    visible.push(selected)
  } else {
    visible[visible.length - 1] = selected
  }
  return visible
}

export function selectedPreviewIndex(prizes: readonly LuckyDrawPrizePreview[], result: ActivityResult | null): number {
  if (!result?.prizeId) return -1
  return prizes.findIndex((prize) => prize.id === result.prizeId)
}

export function settlingStepCount(start: number, target: number, count: number): number {
  if (count <= 0) return 0
  const fullLaps = Math.ceil(16 / count) * count
  return fullLaps + ((target - start % count + count) % count)
}

export function hasPositiveDrawReward(result: ActivityResult): boolean {
  if (result.kind !== 'draw') return false
  const reward = result.reward
  if (reward.kind === 'coupon_grant' || reward.kind === 'subscription_extension') return true
  if (reward.kind !== 'txb_delta') return false
  try { return BigInt(reward.txbDeltaMinor) > 0n } catch { return false }
}
