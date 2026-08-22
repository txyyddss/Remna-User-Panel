import { computed, onScopeDispose, shallowRef, type ComputedRef, type Ref } from 'vue'

import { api } from '@/api/client'
import type { NormalizedDistribution, SquadProduct } from '@/api/types'
import { catalogSquadPresentation } from './catalogSquadPresentation'

export function useCatalogSquadPresentation(
  squads: ComputedRef<readonly SquadProduct[]>,
  includedIds: ComputedRef<readonly string[]>,
  comboId: Ref<string | null>,
) {
  const composition = shallowRef<readonly NormalizedDistribution[]>([])
  let disposed = false

  void api.getStatistics()
    .then((snapshot) => {
      if (!disposed) composition.value = snapshot.database.squadByCombo
    })
    .catch(() => { /* Composition ranking must not block catalog browsing. */ })

  onScopeDispose(() => { disposed = true })

  const presentation = computed(() => catalogSquadPresentation(
    squads.value,
    includedIds.value,
    composition.value,
    comboId.value,
  ))

  return {
    featuredIds: computed(() => presentation.value.featuredIds),
    orderedIds: computed(() => presentation.value.orderedIds),
  }
}
