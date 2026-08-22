import { computed, onScopeDispose, shallowRef, type ComputedRef, type Ref } from 'vue'

import { api } from '@/api/client'
import type { NormalizedDistribution, SquadProduct } from '@/api/types'
import { featuredCatalogSquadIds } from './catalogFeaturedSquads'

export function useCatalogFeaturedSquads(
  squads: ComputedRef<readonly SquadProduct[]>,
  includedIds: ComputedRef<readonly string[]>,
  comboId: Ref<string | null>,
): ComputedRef<string[]> {
  const composition = shallowRef<readonly NormalizedDistribution[]>([])
  let disposed = false

  void api.getStatistics()
    .then((snapshot) => {
      if (!disposed) composition.value = snapshot.database.squadByCombo
    })
    .catch(() => { /* Featured analytics must not block catalog browsing. */ })

  onScopeDispose(() => { disposed = true })

  return computed(() => featuredCatalogSquadIds(
    squads.value,
    includedIds.value,
    composition.value,
    comboId.value,
  ))
}
