import { computed, onScopeDispose, shallowRef, watch } from 'vue'
import { z } from 'zod'
import { useRoute, useRouter } from 'vue-router'
import { ipLookupApi, type IPLookupCheck, type IPLookupQuote, type IPLookupState } from '@/api/ipLookup'
import { ApiError } from '@/api/http'
import { useSessionStore } from '@/stores/session'
import { createUuid } from '@/utils/browserCompatibility'

const terminal = new Set(['succeeded', 'partial', 'failed', 'compensated', 'pending_review'])
const ipSchema = z.union([z.ipv4(), z.ipv6()])

export function useIPLookup() {
  const session = useSessionStore()
  const route = useRoute()
  const router = useRouter()
  const ip = shallowRef('')
  const state = shallowRef<IPLookupState | null>(null)
  const quote = shallowRef<IPLookupQuote | null>(null)
  const check = shallowRef<IPLookupCheck | null>(null)
  const loading = shallowRef(true)
  const busy = shallowRef(false)
  const errorCode = shallowRef('')
  const validIP = computed(() => ipSchema.safeParse(ip.value.trim()).success)
  let disposed = false
  let revision = 0
  let sessionRevision = 0
  let quoteTimer: ReturnType<typeof setTimeout> | undefined
  let pollTimer: ReturnType<typeof setTimeout> | undefined
  let pendingSubmission: { quote: IPLookupQuote; key: string } | null = null
  const ownerSnapshot = () => ({ id: session.user?.id, version: sessionRevision })
  const isCurrent = (owner: ReturnType<typeof ownerSnapshot>) => !disposed && owner.id === session.user?.id && owner.version === sessionRevision
  const report = computed(() => check.value?.report ?? quote.value?.cachedReport ?? null)
  const requestedIP = computed(() => check.value?.requestedIP ?? quote.value?.ip ?? ip.value)
  const cacheMatch = computed(() => check.value?.cacheMatch ?? quote.value?.cacheMatch ?? 'none')
  const setError = (error: unknown) => { errorCode.value = error instanceof ApiError ? error.code : 'IP_LOOKUP_FAILED' }

  async function load(): Promise<void> {
    const owner = ownerSnapshot()
    try { const value = await ipLookupApi.state(); if (isCurrent(owner)) state.value = value }
    catch (error) { if (isCurrent(owner)) setError(error) }
    finally { if (isCurrent(owner)) loading.value = false }
  }

  async function updateQuote(): Promise<void> {
    const version = ++revision
    const owner = ownerSnapshot()
    quote.value = null
    if (!validIP.value || !state.value?.enabled || busy.value) return
    try {
      const value = await ipLookupApi.quote(ip.value.trim(), false)
      if (version !== revision || !isCurrent(owner)) return
      quote.value = value
    } catch (error) { if (version === revision && isCurrent(owner)) setError(error) }
  }

  async function poll(id: string, owner: ReturnType<typeof ownerSnapshot>): Promise<void> {
    try {
      const value = await ipLookupApi.check(id)
      if (!isCurrent(owner)) return
      check.value = value
      if (!ip.value && value.report) ip.value = value.requestedIP || value.report.ip
      if (terminal.has(value.operation.status)) {
        busy.value = false
        errorCode.value = value.operation.errorCode ?? ''
        await load()
        if (isCurrent(owner)) await updateQuote()
      } else pollTimer = setTimeout(() => void poll(id, owner), 1000)
    } catch (error) {
      if (!isCurrent(owner)) return
      setError(error)
      if (error instanceof ApiError && error.status < 500 && error.status !== 429) { busy.value = false; return }
      pollTimer = setTimeout(() => void poll(id, owner), 3000)
    }
  }

  async function submit(): Promise<void> {
    if (busy.value) return
    const selected = quote.value
    if (!selected || selected.cachedReport || !validIP.value) return
    busy.value = true
    errorCode.value = ''
    const owner = ownerSnapshot()
    pendingSubmission ??= { quote: selected, key: createUuid() }
    try {
      const operation = await ipLookupApi.submit(pendingSubmission.quote, pendingSubmission.key)
      if (!isCurrent(owner)) return
      pendingSubmission = null
      void router.replace({ query: { ...route.query, check: operation.id } }).catch(() => undefined)
      await poll(operation.id, owner)
    } catch (error) {
      if (!isCurrent(owner)) return
      busy.value = false
      setError(error)
      if (error instanceof ApiError && error.status < 500) pendingSubmission = null
      if (!pendingSubmission) { await load(); if (isCurrent(owner)) await updateQuote() }
    }
  }

  watch(ip, () => {
    if (check.value?.requestedIP === ip.value.trim()) return
    revision++
    quote.value = null
    if (!busy.value) { errorCode.value = ''; pendingSubmission = null; check.value = null }
    clearTimeout(quoteTimer)
    quoteTimer = setTimeout(() => void updateQuote(), 350)
  })
  watch(() => session.user?.id, () => {
    revision++
    sessionRevision++
    clearTimeout(quoteTimer)
    clearTimeout(pollTimer)
    state.value = quote.value = check.value = null
    ip.value = ''
    errorCode.value = ''
    pendingSubmission = null
    busy.value = false
    loading.value = true
    void load()
    if (session.user?.id && typeof route.query.check === 'string') {
      busy.value = true
      void poll(route.query.check, ownerSnapshot())
    }
  }, { immediate: true })
  onScopeDispose(() => { disposed = true; revision++; clearTimeout(quoteTimer); clearTimeout(pollTimer) })
  return { ip, state, quote, check, report, requestedIP, cacheMatch, loading, busy, errorCode, validIP, submit, load }
}
