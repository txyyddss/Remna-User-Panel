import { inject, onMounted, onScopeDispose, provide, readonly, shallowRef, type InjectionKey, type ShallowRef } from 'vue'

import { api } from '@/api/client'

interface CommunityAccessContext {
  activeCombo: Readonly<ShallowRef<boolean>>
  loading: Readonly<ShallowRef<boolean>>
  refresh: () => Promise<void>
}

const communityAccessKey: InjectionKey<CommunityAccessContext> = Symbol('community-access')

function createCommunityAccess(): CommunityAccessContext {
  const activeCombo = shallowRef(false)
  const loading = shallowRef(true)
  let disposed = false
  let pending: Promise<void> | null = null

  function refresh(): Promise<void> {
    if (pending) return pending
    const request = api.getCommunityAccess()
      .then((response) => {
        if (!disposed) activeCombo.value = response.activeCombo
      })
      .catch(() => {
        if (!disposed) activeCombo.value = false
      })
      .finally(() => {
        if (!disposed) loading.value = false
        if (pending === request) pending = null
      })
    pending = request
    return request
  }

  onMounted(() => void refresh())
  onScopeDispose(() => { disposed = true })

  return { activeCombo: readonly(activeCombo), loading: readonly(loading), refresh }
}

export function useCommunityAccess(): CommunityAccessContext {
  const provided = inject(communityAccessKey, null)
  if (provided) return provided
  const context = createCommunityAccess()
  provide(communityAccessKey, context)
  return context
}
