import type { NormalizedDistribution, SquadProduct } from '@/api/types'

export function featuredCatalogSquadIds(
  squads: readonly SquadProduct[],
  includedIds: readonly string[],
  distributions: readonly NormalizedDistribution[],
  comboId: string | null,
): string[] {
  if (!comboId) return []
  const composition = distributions.find(row => row.id === comboId)
  if (!composition) return []

  const included = new Set(includedIds)
  const shares = new Map(composition.segments.map(segment => [segment.id, segment.value]))
  const candidates = squads
    .filter(squad => !included.has(squad.id) && squad.stockRemaining !== 0)
    .map(squad => ({ id: squad.id, share: shares.get(squad.remnaSquadUuid) ?? 0 }))
    .filter(candidate => Number.isFinite(candidate.share) && candidate.share > 0)

  const maximum = Math.max(0, ...candidates.map(candidate => candidate.share))
  return maximum > 0
    ? candidates.filter(candidate => candidate.share === maximum).map(candidate => candidate.id)
    : []
}
