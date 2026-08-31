import { onMounted, onUnmounted, readonly, shallowRef } from 'vue'

import { api } from '@/api/client'
import type { CommunityMembership, CommunitySpace } from '@/api/types'
import { localizedError } from '@/i18n'
import { useSessionStore } from '@/stores/session'
import { getTelegramWebApp, notifyHaptic, openExternalLink } from '@/utils/telegram'

export function useCommunityMembership() {
  const sessionStore = useSessionStore()
  const membership = shallowRef<CommunityMembership | null>(null)
  const loading = shallowRef(true)
  const refreshing = shallowRef(false)
  const joining = shallowRef<readonly CommunitySpace[]>([])
  const error = shallowRef<string | null>(null)
  let disposed = false
  let telegram: TelegramWebApp | undefined

  async function load(initial = false): Promise<void> {
    if (initial) loading.value = true
    else refreshing.value = true
    error.value = null
    try {
      const next = await api.checkCommunityMembership()
      if (disposed) return
      membership.value = next
      if (sessionStore.session) sessionStore.updateSession({ ...sessionStore.session, user: next.user })
    } catch (caught) {
      if (!disposed) error.value = localizedError(caught, 'community.errorDescription')
    } finally {
      if (initial) loading.value = false
      else refreshing.value = false
    }
  }

  async function join(space: CommunitySpace): Promise<void> {
    if (joining.value.includes(space)) return
    joining.value = [...joining.value, space]
    error.value = null
    try {
      const invite = await api.createCommunityInvite(space)
      openExternalLink(invite.url)
    } catch (caught) {
      error.value = localizedError(caught, 'community.errorDescription')
      notifyHaptic('error')
    } finally {
      joining.value = joining.value.filter((value) => value !== space)
    }
  }

  const refreshAfterActivation = (): void => { void load() }

  onMounted(() => {
    disposed = false
    telegram = getTelegramWebApp()
    telegram?.onEvent?.('activated', refreshAfterActivation)
    void load(true)
  })
  onUnmounted(() => {
    disposed = true
    telegram?.offEvent?.('activated', refreshAfterActivation)
  })

  return {
    membership: readonly(membership),
    loading: readonly(loading),
    refreshing: readonly(refreshing),
    joining: readonly(joining),
    error: readonly(error),
    load,
    join,
  }
}
