import { describe, expect, it } from 'vitest'

import type { ActivityResult, LuckyDrawPrizePreview } from '@/api/features'
import { hasPositiveDrawReward, previewPrizes, randomDrawStyles, resolveDrawStyle, selectedPreviewIndex, settlingStepCount } from './selection'

function receipt(prizeId: string, prizeName: string, reward: ActivityResult['reward'] = { kind: 'none' }): ActivityResult {
  return {
    id: 'result-1', kind: 'draw', outcome: 'complete', message: '', reward,
    drawId: 'draw-1', prizeId, prizeName,
    balanceAfter: { currency: 'TXB', minor: '100', display: '1.00 TXB' },
    createdAt: '2026-09-25T00:00:00Z',
  }
}

describe('draw presentation selection', () => {
  it('can choose every animated style in Random mode without choosing Simple', () => {
    for (let index = 0; index < randomDrawStyles.length; index += 1) {
      expect(resolveDrawStyle('random', () => (index + 0.5) / randomDrawStyles.length)).toBe(randomDrawStyles[index])
    }
    expect(randomDrawStyles).not.toContain('simple')
    expect(resolveDrawStyle('scratch', () => 0)).toBe('scratch')
  })

  it('keeps eight visible prizes and inserts a selected prize missing from the current preview', () => {
    const prizes: LuckyDrawPrizePreview[] = Array.from({ length: 200 }, (_, index) => ({
      id: 'prize-' + index, name: 'Prize ' + index,
    }))
    const selected = receipt('prize-199', 'Selected after stock changed')
    const preview = previewPrizes(prizes, 30, selected)
    expect(preview).toHaveLength(8)
    expect(preview[7]).toEqual({ id: 'prize-199', name: 'Selected after stock changed' })
    expect(selectedPreviewIndex(preview, selected)).toBe(7)
  })

  it('settles the grid and slot on the selected prize for every preview size', () => {
    for (let count = 1; count <= 8; count += 1) {
      for (let start = 0; start <= 8; start += 1) {
        for (let target = 0; target < count; target += 1) {
          const steps = settlingStepCount(start, target, count)
          expect(steps).toBeGreaterThanOrEqual(16)
          expect((start + steps) % count).toBe(target)
        }
      }
    }
  })

  it('does not celebrate no prize, zero, or negative TXB rewards', () => {
    expect(hasPositiveDrawReward(receipt('p', 'No prize'))).toBe(false)
    expect(hasPositiveDrawReward(receipt('p', 'No gain', { kind: 'txb_delta', txbDeltaMinor: '0' }))).toBe(false)
    expect(hasPositiveDrawReward(receipt('p', 'Deduction', { kind: 'txb_delta', txbDeltaMinor: '-50' }))).toBe(false)
    expect(hasPositiveDrawReward(receipt('p', 'Reward', { kind: 'txb_delta', txbDeltaMinor: '50' }))).toBe(true)
    expect(hasPositiveDrawReward(receipt('p', 'Coupon', { kind: 'coupon_grant', couponId: 'coupon-1' }))).toBe(true)
  })
})
