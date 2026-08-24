import type { NormalizedDistribution, SquadProduct } from '@/api/types'

export interface CatalogSquadPresentation {
  featuredIds: readonly string[]
  orderedIds: readonly string[]
}

export function catalogSquadPresentation(
  squads: readonly SquadProduct[],
  includedIds: readonly string[],
  distributions: readonly NormalizedDistribution[],
  comboId: string | null,
): CatalogSquadPresentation {
  const composition = comboId ? distributions.find(row => row.id === comboId) : undefined
  const shares = new Map(composition?.segments.map(segment => [segment.id, segment.value]) ?? [])
  const included = new Set(includedIds)
  const rows = squads.map((squad, index) => {
    const value = shares.get(squad.remnaSquadUuid) ?? 0
    return {
      id: squad.id,
      index,
      bottom: included.has(squad.id) || (squad.stockRemaining === 0 && !squad.stockHeldByCurrentUser),
      share: Number.isFinite(value) && value > 0 ? value : 0,
    }
  })
  const selectable = rows
    .filter(row => !row.bottom)
    .sort((left, right) => right.share - left.share || left.index - right.index)
  const bottom = rows.filter(row => row.bottom)
  const maximum = selectable[0]?.share ?? 0

  return {
    featuredIds: maximum > 0
      ? selectable.filter(row => row.share === maximum).map(row => row.id)
      : [],
    orderedIds: [...selectable, ...bottom].map(row => row.id),
  }
}
