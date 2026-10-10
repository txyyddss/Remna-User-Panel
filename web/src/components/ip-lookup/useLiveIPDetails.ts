import { onScopeDispose, shallowRef, watch } from 'vue'
import { ipLookupApi, type IPLookupLiveDetails } from '@/api/ipLookup'
import { ApiError } from '@/api/http'
import { useSessionStore } from '@/stores/session'

export function useLiveIPDetails(address: () => string, reportKey: () => string) {
  const session = useSessionStore()
  const details = shallowRef<IPLookupLiveDetails | null>(null)
  const loading = shallowRef(false)
  const errorCode = shallowRef('')
  let controller: AbortController | undefined
  let revision = 0

  async function refresh(): Promise<void> {
    const ip = address()
    if (!reportKey() || !session.user?.id || loading.value) return
    controller?.abort()
    controller = new AbortController()
    const current = ++revision
    loading.value = true
    details.value = null
    errorCode.value = ''
    try {
      const result = await ipLookupApi.details(ip, controller.signal)
      if (current === revision && !controller.signal.aborted && result.ip === ip) details.value = result
    } catch (error) {
      if (current === revision && !controller.signal.aborted) errorCode.value = error instanceof ApiError ? error.code : 'IP_DETAILS_UNAVAILABLE'
    } finally { if (current === revision) loading.value = false }
  }

  watch(() => `${session.user?.id ?? ''}|${address()}|${reportKey()}`, () => {
    revision++
    controller?.abort()
    details.value = null
    loading.value = false
    errorCode.value = ''
    void refresh()
  }, { immediate: true })
  onScopeDispose(() => { revision++; controller?.abort() })
  return { details, loading, errorCode, refresh }
}
