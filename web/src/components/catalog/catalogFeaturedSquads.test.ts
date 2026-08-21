import { describe, expect, it } from 'vitest'

import type { NormalizedDistribution, SquadProduct } from '@/api/types'
import { featuredCatalogSquadIds } from './catalogFeaturedSquads'

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

describe('featuredCatalogSquadIds', () => {
  it('marks every tied leader after excluding included and full squads', () => {
    expect(featuredCatalogSquadIds(squads, ['included'], distributions, 'combo-1'))
      .toEqual(['featured-a', 'featured-b'])
  })

  it('returns no badge without matching positive composition data', () => {
    expect(featuredCatalogSquadIds(squads, [], distributions, 'missing')).toEqual([])
    expect(featuredCatalogSquadIds(squads, [], [], 'combo-1')).toEqual([])
    expect(featuredCatalogSquadIds(squads, [], [{ id: 'combo-1', label: 'Core', segments: [] }], 'combo-1')).toEqual([])
  })
})
