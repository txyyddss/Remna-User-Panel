import type { ShallowRef } from 'vue'
import type { AdminCatalogOptions } from '../adminOperations'
import { restoreCached } from './restore'

export function restoreCatalogOptions(target: ShallowRef<AdminCatalogOptions>): boolean {
  return [
    restoreCached<{ items: AdminCatalogOptions['combos'] }>('/api/v1/admin/combos', page => {
      target.value = { ...target.value, combos: page.items }
    }),
    restoreCached<{ items: AdminCatalogOptions['squads'] }>('/api/v1/admin/squad-products', page => {
      target.value = { ...target.value, squads: page.items }
    }),
  ].every(Boolean)
}
