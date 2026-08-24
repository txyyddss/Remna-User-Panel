import { describe, expect, it } from 'vitest'

import type { NormalizedDistribution, SquadProduct } from '@/api/types'
import { catalogSquadPresentation } from './catalogSquadPresentation'

function squad(id: string, remnaSquadUuid: string, stockRemaining: number | null = null): SquadProduct {
  return {
    id,
    remnaSquadUuid,
    name: id,
    description: '',
    profile: null,
    price: { currency: 'TXB', minor: '100', display: '1.00 TXB' },
    visible: true,
  upstreamPresent: true,
  stockHeldByCurrentUser: false,
  activationRequired: false,
    accessibleNodes: [],
    stockRemaining,
    createdAt: '2026-08-21T00:00:00Z',
    updatedAt: '2026-08-21T00:00:00Z',
  }
}

const squads = [
  squad('included', '00000000-0000-4000-8000-000000000001'),
  squad('full', '00000000-0000-4000-8000-000000000002', 0),
  squad('featured-a', '00000000-0000-4000-8000-000000000003'),
  squad('featured-b', '00000000-0000-4000-8000-000000000004'),
  squad('other', '00000000-0000-4000-8000-000000000005'),
]

const distributions: readonly NormalizedDistribution[] = [{
  id: 'combo-1',
  label: 'Core',
  segments: [
    { id: squads[0]!.remnaSquadUuid, label: 'Included', value: 90 },
    { id: squads[1]!.remnaSquadUuid, label: 'Full', value: 80 },
    { id: squads[2]!.remnaSquadUuid, label: 'Featured A', value: 40 },
    { id: squads[3]!.remnaSquadUuid, label: 'Featured B', value: 40 },
    { id: squads[4]!.remnaSquadUuid, label: 'Other', value: 10 },
  ],
}]

describe('catalogSquadPresentation', () => {
  it('sorts selectable squads by composition and moves included and full squads to the bottom', () => {
    expect(catalogSquadPresentation(squads, ['included'], distributions, 'combo-1'))
      .toEqual({
        featuredIds: ['featured-a', 'featured-b'],
        orderedIds: ['featured-a', 'featured-b', 'other', 'included', 'full'],
      })
  })

  it('keeps selectable and bottom groups stable without matching composition data', () => {
    expect(catalogSquadPresentation(squads, ['included'], distributions, 'missing'))
      .toEqual({
        featuredIds: [],
        orderedIds: ['featured-a', 'featured-b', 'other', 'included', 'full'],
      })
  })
})
