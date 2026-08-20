import { computed, inject, onMounted, onScopeDispose, provide, readonly, shallowRef, type InjectionKey, type ShallowRef } from 'vue'

import { api } from '@/api/client'
import { localizedError, t } from '@/i18n'
import type { Catalog, CatalogNode, Dashboard, DashboardNodeUsage } from '@/api/types'
import { notifyHaptic } from '@/utils/telegram'
import { createLatestRequest } from '@/utils/latestRequest'
import { useDurableCommand } from './useDurableCommand'

const isoDate = /^\d{4}-\d{2}-\d{2}$/
const maximumNodeUsageRangeMilliseconds = 30 * 24 * 60 * 60 * 1000

interface NodeUsageController {
  nodeUsage: ShallowRef<DashboardNodeUsage | null>
  nodeUsageLoading: ShallowRef<boolean>
  nodeUsageError: ShallowRef<string | null>
  nodeUsageStart: ShallowRef<string>
  nodeUsageEnd: ShallowRef<string>
  loadNodeUsage: () => Promise<void>
  setNodeUsageStart: (value: string) => void
  setNodeUsageEnd: (value: string) => void
  dispose: () => void
}

const nodeUsageControllerKey: InjectionKey<NodeUsageController> = Symbol('dashboard-node-usage')

function utcDate(offsetDays: number): string {
  const date = new Date()
  date.setUTCDate(date.getUTCDate() + offsetDays)
  return date.toISOString().slice(0, 10)
}

function validRange(start: string, end: string): boolean {
  if (!validDate(start) || !validDate(end) || start > end) return false
  return Date.parse(`${end}T00:00:00.000Z`) - Date.parse(`${start}T00:00:00.000Z`) <= maximumNodeUsageRangeMilliseconds
}

function validDate(value: string): boolean {
  if (!isoDate.test(value)) return false
  const parsed = new Date(`${value}T00:00:00.000Z`)
  return Number.isFinite(parsed.valueOf()) && parsed.toISOString().slice(0, 10) === value
}

function createNodeUsageController(): NodeUsageController {
  const nodeUsage = shallowRef<DashboardNodeUsage | null>(null)
  const nodeUsageLoading = shallowRef(false)
  const nodeUsageError = shallowRef<string | null>(null)
  const nodeUsageStart = shallowRef(utcDate(-6))
  const nodeUsageEnd = shallowRef(utcDate(0))
  const latest = createLatestRequest()

  function resetNodeUsage(): void {
    latest.invalidate()
    nodeUsage.value = null
    nodeUsageError.value = null
    nodeUsageLoading.value = false
  }

  function setNodeUsageStart(value: string): void {
    nodeUsageStart.value = value
    resetNodeUsage()
  }

  function setNodeUsageEnd(value: string): void {
    nodeUsageEnd.value = value
    resetNodeUsage()
  }

  async function loadNodeUsage(): Promise<void> {
    const token = latest.begin()
    const start = nodeUsageStart.value
    const end = nodeUsageEnd.value
    if (!validRange(start, end)) {
      nodeUsageError.value = t('home.trafficRangeInvalid')
      return
    }
    if (nodeUsage.value?.startDate === start && nodeUsage.value.endDate === end) return
    nodeUsageLoading.value = true
    nodeUsageError.value = null
    try {
      const response = await api.getDashboardNodeUsage(start, end)
      if (latest.isCurrent(token)) nodeUsage.value = response
    } catch (caught) {
      if (!latest.isCurrent(token)) return
      nodeUsageError.value = localizedError(caught, 'errors.nodeUsageUnavailable')
      notifyHaptic('error')
    } finally {
      if (latest.isCurrent(token)) nodeUsageLoading.value = false
    }
  }

  return { nodeUsage, nodeUsageLoading, nodeUsageError, nodeUsageStart, nodeUsageEnd, loadNodeUsage, setNodeUsageStart, setNodeUsageEnd, dispose: latest.dispose }
}

export function useDashboard() {
  const dashboard = shallowRef<Dashboard | null>(null)
  const catalog = shallowRef<Catalog | null>(null)
  const loading = shallowRef(true)
  const refreshing = shallowRef(false)
  const loadError = shallowRef<string | null>(null)
  const latestLoad = createLatestRequest()
  const nodeUsageController = createNodeUsageController()
  const { nodeUsage, nodeUsageLoading, nodeUsageError, nodeUsageStart, nodeUsageEnd, loadNodeUsage, setNodeUsageStart, setNodeUsageEnd } = nodeUsageController
  const revokeCommand = useDurableCommand({
    errorKey: 'errors.subscriptionRevoke',
    onTerminal: async (receipt) => {
      if (receipt.status === 'succeeded') {
        await load({ quiet: true })
        notifyHaptic('success')
      } else notifyHaptic('error')
    },
  })
  const error = computed(() => loadError.value ?? revokeCommand.error.value)

  const hasEntitlement = computed(() => dashboard.value?.activePurchase != null)
  const catalogNodes = computed<readonly CatalogNode[]>(() => catalog.value?.nodes ?? [])
  const usageRatio = computed(() => {
    if (!dashboard.value?.statistics) return 0
    const used = Number(dashboard.value.statistics.usedTrafficBytes)
    const limit = Number(dashboard.value.statistics.trafficLimitBytes)
    if (!Number.isFinite(used) || !Number.isFinite(limit) || limit <= 0) return 0
    return Math.min(1, used / limit)
  })
  const activeSquadNames = computed(() => {
    const active = dashboard.value?.activePurchase
    if (!active || !catalog.value) return []
    const names = new Map(catalog.value.addons.map((squad) => [squad.remnaSquadUuid, squad.name]))
    for (const combo of catalog.value.combos) {
      for (const squad of combo.includedSquads) names.set(squad.remnaSquadUuid, squad.name)
    }
    return active.squadUuids.map((uuid) => names.get(uuid)).filter((name): name is string => Boolean(name))
  })

  async function load(options: { quiet?: boolean } = {}): Promise<void> {
    const token = latestLoad.begin()
    if (options.quiet) refreshing.value = true
    else loading.value = true
    loadError.value = null
    try {
      const [dashboardResponse, catalogResponse] = await Promise.all([
        api.getDashboard(),
        api.getCatalog().catch(() => null),
      ])
      if (!latestLoad.isCurrent(token)) return
      dashboard.value = dashboardResponse
      catalog.value = catalogResponse
    } catch (caught) {
      if (!latestLoad.isCurrent(token)) return
      loadError.value = localizedError(caught, 'errors.dashboardUnavailable')
    } finally {
      if (latestLoad.isCurrent(token)) {
        loading.value = false
        refreshing.value = false
      }
    }
  }

  async function revokeSubscription(): Promise<boolean> {
    const accepted = await revokeCommand.execute('subscription-revoke', 'subscription-revoke', api.revokeSubscription)
    if (!accepted) notifyHaptic('error')
    return accepted
  }

  provide(nodeUsageControllerKey, nodeUsageController)

  onMounted(() => void load())
  onScopeDispose(() => {
    latestLoad.dispose()
    nodeUsageController.dispose()
  })

  return {
    dashboard: readonly(dashboard), loading: readonly(loading), refreshing: readonly(refreshing),
    revoking: revokeCommand.busy, revokeBlocked: revokeCommand.blocksMutations,
    revokeReceipt: revokeCommand.receipt, revokeChecking: revokeCommand.checking,
    revokeError: revokeCommand.error,
    error: readonly(error),
    nodeUsage: readonly(nodeUsage), nodeUsageLoading: readonly(nodeUsageLoading),
    nodeUsageError: readonly(nodeUsageError), nodeUsageStart: readonly(nodeUsageStart),
    nodeUsageEnd: readonly(nodeUsageEnd), hasEntitlement,
    usageRatio,
    catalogNodes,
    activeSquadNames,
    load,
    revokeSubscription,
    refreshRevoke: revokeCommand.refresh,
    loadNodeUsage,
    setNodeUsageStart,
    setNodeUsageEnd,
  }
}

export function useDashboardNodeUsage(): NodeUsageController {
  const injected = inject(nodeUsageControllerKey, null)
  if (injected) return injected
  const controller = createNodeUsageController()
  onScopeDispose(controller.dispose)
  return controller
}
