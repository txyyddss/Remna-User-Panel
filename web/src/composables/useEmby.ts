import { onMounted, readonly, shallowRef } from 'vue'

import type { EmbyAccount, EmbyOverview } from '@/api/features'
import { featuresApi } from '@/api/features'
import { notifyHaptic } from '@/utils/telegram'

export function useEmby() {
  const overview = shallowRef<EmbyOverview | null>(null)
  const loading = shallowRef(true)
  const busy = shallowRef<'setup' | 'preferences' | 'password' | null>(null)
  const error = shallowRef<string | null>(null)
  const message = shallowRef<string | null>(null)

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      overview.value = await featuresApi.getEmby()
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'Emby account details are unavailable.'
    } finally {
      loading.value = false
    }
  }

  async function perform(kind: NonNullable<typeof busy.value>, action: () => Promise<EmbyAccount | void>, success: string): Promise<boolean> {
    if (busy.value) return false
    busy.value = kind
    error.value = null
    message.value = null
    try {
      const account = await action()
      if (account && overview.value) overview.value = { ...overview.value, account }
      message.value = success
      notifyHaptic('success')
      return true
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'The Emby update could not be completed.'
      notifyHaptic('error')
      return false
    } finally {
      busy.value = null
    }
  }

  function setup(payload: { password: string; maxParentalRating: number | null; libraryIds: string[] }): Promise<boolean> {
    return perform('setup', () => featuresApi.setupEmby(payload), 'Provisioning started. You can leave this page safely.')
  }

  function updatePreferences(payload: { maxParentalRating: number | null; libraryIds: string[] }): Promise<boolean> {
    return perform('preferences', () => featuresApi.updateEmbyPreferences(payload), 'Emby preferences updated.')
  }

  function changePassword(password: string): Promise<boolean> {
    return perform('password', () => featuresApi.changeEmbyPassword(password), 'Emby password changed.')
  }

  onMounted(() => void load())

  return {
    overview: readonly(overview),
    loading: readonly(loading),
    busy: readonly(busy),
    error: readonly(error),
    message: readonly(message),
    load,
    setup,
    updatePreferences,
    changePassword,
  }
}
