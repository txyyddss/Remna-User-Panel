import { onMounted, onScopeDispose, shallowRef } from 'vue'

import { abuseApi, type AbuseNode, type AbusePolicy, type AbusePunishment, type AbuseRecord, type AbuseRule } from '@/api/abuse'
import { localizedError } from '@/i18n'
import { haptic, notifyHaptic, type HapticIntent } from '@/utils/telegramHaptics'
import { createLatestRequest } from '@/utils/latestRequest'

interface ExecuteOptions {
  errorKey?: string
  reload?: boolean
  successHaptic?: HapticIntent
}

export function useAdminAbuse() {
  const policy = shallowRef<AbusePolicy | null>(null)
  const nodes = shallowRef<AbuseNode[]>([])
  const rules = shallowRef<AbuseRule[]>([])
  const punishments = shallowRef<AbusePunishment[]>([])
  const records = shallowRef<AbuseRecord[]>([])
  const statistics = shallowRef<Record<string, number>>({})
  const whitelist = shallowRef<string[]>([])
  const loading = shallowRef(true)
  const loadingMoreRecords = shallowRef(false)
  const busy = shallowRef(false)
  const loadError = shallowRef<string | null>(null)
  const operationError = shallowRef<string | null>(null)
  const nextRecordsCursor = shallowRef('')
  const latestLoad = createLatestRequest()
  const latestRecords = createLatestRequest()

  async function load(): Promise<boolean> {
    const loadToken = latestLoad.begin()
    const recordsToken = latestRecords.begin()
    loading.value = true
    loadError.value = null
    try {
      const [nextPolicy, nextNodes, nextRules, nextPunishments, page, nextStatistics, nextWhitelist] = await Promise.all([
        abuseApi.policy(), abuseApi.nodes(), abuseApi.rules(), abuseApi.punishments(),
        abuseApi.adminRecords(), abuseApi.statistics(), abuseApi.whitelist(),
      ])
      if (!latestLoad.isCurrent(loadToken) || !latestRecords.isCurrent(recordsToken)) return false
      policy.value = nextPolicy
      nodes.value = nextNodes
      rules.value = nextRules
      punishments.value = nextPunishments
      records.value = page.items
      nextRecordsCursor.value = page.nextCursor
      statistics.value = nextStatistics
      whitelist.value = nextWhitelist
      return true
    } catch (caught) {
      if (!latestLoad.isCurrent(loadToken)) return false
      loadError.value = localizedError(caught, 'adminAbuse.loadFailed')
      return false
    } finally {
      if (latestLoad.isCurrent(loadToken)) loading.value = false
    }
  }

  async function loadMoreRecords(): Promise<void> {
    if (!nextRecordsCursor.value || loadingMoreRecords.value) return
    const token = latestRecords.begin()
    loadingMoreRecords.value = true
    try {
      const page = await abuseApi.adminRecords(nextRecordsCursor.value)
      if (!latestRecords.isCurrent(token)) return
      records.value = [...records.value, ...page.items]
      nextRecordsCursor.value = page.nextCursor
    } catch (caught) {
      if (latestRecords.isCurrent(token)) operationError.value = localizedError(caught, 'adminAbuse.recordsFailed')
    } finally {
      if (latestRecords.isCurrent(token)) loadingMoreRecords.value = false
    }
  }

  async function execute(work: () => Promise<void>, options: ExecuteOptions = {}): Promise<boolean> {
    if (busy.value) return false
    const { errorKey = 'adminAbuse.saveFailed', reload = true, successHaptic } = options
    busy.value = true
    operationError.value = null
    try {
      await work()
      if (reload && !await load()) return false
      if (successHaptic) haptic(successHaptic)
      else notifyHaptic('success')
      return true
    } catch (caught) {
      operationError.value = localizedError(caught, errorKey)
      notifyHaptic('error')
      return false
    } finally {
      busy.value = false
    }
  }

  onMounted(() => { void load() })
  onScopeDispose(() => { latestLoad.dispose(); latestRecords.dispose() })

  return { policy, nodes, rules, punishments, records, statistics, whitelist, loading, loadingMoreRecords, busy, loadError, operationError, nextRecordsCursor, load, loadMoreRecords, execute }
}
