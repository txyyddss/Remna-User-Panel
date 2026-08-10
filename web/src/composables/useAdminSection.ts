import { onMounted, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { AdminResource, Paginated } from '@/api/types'
import { notifyHaptic } from '@/utils/telegram'
import { localizedError } from '@/i18n'

function extractItems<T>(payload: Paginated<T> | T[] | { items: T[] }): T[] {
  if (Array.isArray(payload)) return payload
  return payload.items
}

export function useAdminSection<T>(resource: AdminResource, options: { immediate?: boolean } = {}) {
  const items = shallowRef<T[]>([])
  const loading = shallowRef(options.immediate !== false)
  const busy = shallowRef(false)
  const error = shallowRef<string | null>(null)

  async function load(query?: Record<string, string | number | boolean | undefined>): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const payload = await api.getAdminResource<Paginated<T> | T[] | { items: T[] }>(resource, query)
      items.value = extractItems(payload)
    } catch (caught) {
      error.value = localizedError(caught, 'errors.adminLoad')
    } finally {
      loading.value = false
    }
  }

  async function perform(task: () => Promise<unknown>): Promise<boolean> {
    busy.value = true
    error.value = null
    try {
      await task()
      await load()
      notifyHaptic('success')
      return true
    } catch (caught) {
      error.value = localizedError(caught, 'errors.adminAction')
      notifyHaptic('error')
      return false
    } finally {
      busy.value = false
    }
  }

  function create(body: unknown): Promise<boolean> {
    return perform(() => api.createAdminResource(resource, body))
  }

  function update(id: string, body: unknown): Promise<boolean> {
    return perform(() => api.updateAdminResource(resource, id, body))
  }

  function remove(id: string): Promise<boolean> {
    return perform(() => api.deleteAdminResource(resource, id))
  }

  if (options.immediate !== false) onMounted(() => void load())

  return {
    items: readonly(items),
    loading: readonly(loading),
    busy: readonly(busy),
    error: readonly(error),
    load,
    perform,
    create,
    update,
    remove,
  }
}
