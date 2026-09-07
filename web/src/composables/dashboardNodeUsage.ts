import { shallowRef, type ShallowRef } from 'vue'
import { api } from '@/api/client'
import { restoreRef } from '@/api/cache/restore'
import type { DashboardNodeUsage } from '@/api/types'
import { localizedError, t } from '@/i18n'
import { notifyHaptic } from '@/utils/telegram'
import { createLatestRequest } from '@/utils/latestRequest'

const isoDate = /^\d{4}-\d{2}-\d{2}$/
const maximumNodeUsageRangeMilliseconds = 30 * 24 * 60 * 60 * 1000

export interface NodeUsageController {
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

export function createNodeUsageController(): NodeUsageController {
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
    nodeUsageLoading.value = !restoreRef('/api/v1/dashboard/node-usage', nodeUsage, { query: { start, end } })
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
