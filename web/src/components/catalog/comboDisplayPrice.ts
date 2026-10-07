import type { Combo, Money, SquadProduct } from '@/api/types'

// Catalog browsing uses the same full-term squad prices as purchase quotes.
export function comboDisplayPrice(combo: Pick<Combo, 'price' | 'includedSquads'>, selected: readonly Pick<SquadProduct, 'remnaSquadUuid' | 'price'>[], includeNodes: boolean): Money {
  if (!includeNodes) return combo.price
  const included = new Set(combo.includedSquads.map(squad => squad.remnaSquadUuid))
  const seen = new Set<string>()
  let minor = BigInt(combo.price.minor)
  for (const squad of selected) {
    if (included.has(squad.remnaSquadUuid) || seen.has(squad.remnaSquadUuid)) continue
    seen.add(squad.remnaSquadUuid)
    minor += BigInt(squad.price.minor)
  }
  return { currency: 'TXB', minor: minor.toString(), display: `${minor / 100n}.${(minor % 100n).toString().padStart(2, '0')} TXB` }
}
