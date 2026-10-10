import { onScopeDispose, reactive, shallowRef, watch } from 'vue'
import { api } from '@/api/client'
import { ApiError } from '@/api/http'
import { ipLookupApi, type IPLookupAdminSettings } from '@/api/ipLookup'
import type { Combo } from '@/api/types'
import { useSessionStore } from '@/stores/session'
import { useIPLookupAvailability } from '@/composables/useIPLookupAvailability'

export function useAdminIPLookup() {
  const session = useSessionStore()
  const availability = useIPLookupAvailability()
  const settings = reactive<IPLookupAdminSettings>({ enabled: false, lookupFeeTxb: '', refreshFeeTxb: '', providers: [], credentials: {}, configured: {}, comboQuotas: {} })
  const combos = shallowRef<Combo[]>([])
  const loading = shallowRef(true)
  const busy = shallowRef(false)
  const errorCode = shallowRef('')
  const saved = shallowRef(false)
  let disposed = false
  let version = 0
  const setError = (error: unknown) => { errorCode.value = error instanceof ApiError ? error.code : 'IP_LOOKUP_FAILED' }

  async function load(): Promise<void> {
    const current = ++version
    const owner = session.user?.id
    try {
      const [value, catalog] = await Promise.all([ipLookupApi.settings(), api.getAdminResource<{ items: Combo[] }>('combos')])
      if (disposed || current !== version || owner !== session.user?.id) return
      Object.assign(settings, value)
      settings.credentials = Object.fromEntries(value.providers.map(provider => [provider.id, { value: '', clear: false }]))
      combos.value = catalog.items
      errorCode.value = ''
    } catch (error) { if (!disposed && current === version) setError(error) }
    finally { if (!disposed && current === version) loading.value = false }
  }

  async function save(): Promise<void> {
    if (busy.value) return
    busy.value = true
    saved.value = false
    errorCode.value = ''
    const current = version
    try {
      const value = await ipLookupApi.saveSettings(settings)
      if (disposed || current !== version) return
      Object.assign(settings, value)
      settings.credentials = Object.fromEntries(value.providers.map(provider => [provider.id, { value: '', clear: false }]))
      saved.value = true
      await availability.refresh()
    } catch (error) { if (!disposed && current === version) setError(error) }
    finally { if (!disposed && current === version) busy.value = false }
  }

  watch(() => session.user?.id, () => { saved.value = false; loading.value = true; void load() }, { immediate: true })
  onScopeDispose(() => { disposed = true; version++ })
  return { settings, combos, loading, busy, errorCode, saved, save, load }
}
