import { computed, shallowRef } from 'vue'
import { defineStore } from 'pinia'

import { api, ApiError } from '@/api/client'
import type { Session } from '@/api/types'
import { getTelegramInitData } from '@/utils/telegram'

export type SessionStatus = 'idle' | 'loading' | 'ready' | 'error'

export const useSessionStore = defineStore('session', () => {
  const session = shallowRef<Session | null>(null)
  const status = shallowRef<SessionStatus>('idle')
  const error = shallowRef<string | null>(null)

  const user = computed(() => session.value?.user ?? null)
  const isAuthenticated = computed(() => session.value !== null)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const onboardingComplete = computed(() => user.value?.onboardingState === 'complete')

  async function bootstrap(force = false): Promise<void> {
    if (!force && (status.value === 'loading' || status.value === 'ready')) return
    status.value = 'loading'
    error.value = null

    try {
      try {
        session.value = await api.getMe()
      } catch (caught) {
        if (!(caught instanceof ApiError) || caught.status !== 401) throw caught
        const initData = await getTelegramInitData()
        if (!initData) {
          throw new Error('Open TX Carpool from the Telegram Mini App to continue.', { cause: caught })
        }
        session.value = await api.authTelegram(initData)
      }
      status.value = 'ready'
    } catch (caught) {
      session.value = null
      status.value = 'error'
      error.value = caught instanceof Error ? caught.message : 'Authentication could not be completed.'
    }
  }

  function updateSession(next: Session): void {
    session.value = next
    status.value = 'ready'
    error.value = null
  }

  function clear(): void {
    session.value = null
    status.value = 'idle'
    error.value = null
  }

  return {
    session,
    status,
    error,
    user,
    isAuthenticated,
    isAdmin,
    onboardingComplete,
    bootstrap,
    updateSession,
    clear,
  }
})
