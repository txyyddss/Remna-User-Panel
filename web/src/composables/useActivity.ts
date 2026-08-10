import { onMounted, readonly, shallowRef } from 'vue'

import { featuresApi } from '@/api/features'
import type { ActivityOverview, ActivityResult } from '@/api/features'
import { localizedError } from '@/i18n'
import { notifyHaptic } from '@/utils/telegram'

export function useActivity() {
  const overview = shallowRef<ActivityOverview | null>(null)
  const result = shallowRef<ActivityResult | null>(null)
  const loading = shallowRef(true)
  const busy = shallowRef<'check-in' | 'bet' | 'draw' | null>(null)
  const error = shallowRef<string | null>(null)
  const actionKeys = new Map<string, string>()

  function idempotencyKey(action: string): string {
    const existing = actionKeys.get(action)
    if (existing) return existing
    const created = globalThis.crypto.randomUUID()
    actionKeys.set(action, created)
    return created
  }

  async function load(options: { quiet?: boolean } = {}): Promise<void> {
    if (!options.quiet) loading.value = true
    error.value = null
    try {
      overview.value = await featuresApi.getActivity()
    } catch (caught) {
      error.value = localizedError(caught, 'errors.activityUnavailable')
    } finally {
      loading.value = false
    }
  }

  async function run(kind: NonNullable<typeof busy.value>, actionId: string, action: (key: string) => Promise<ActivityResult>): Promise<void> {
    if (busy.value) return
    busy.value = kind
    error.value = null
    try {
      result.value = await action(idempotencyKey(actionId))
      actionKeys.delete(actionId)
      notifyHaptic(result.value.outcome === 'loss' ? 'warning' : 'success')
      await load({ quiet: true })
    } catch (caught) {
      error.value = localizedError(caught, 'errors.activityFailed')
      notifyHaptic('error')
    } finally {
      busy.value = null
    }
  }

  function checkIn(): Promise<void> {
    return run('check-in', 'check-in', featuresApi.checkIn)
  }

  function placeBet(payload: { gameId: string; stakeTxbMinor: string }): Promise<void> {
    const actionId = `bet:${payload.gameId}:${payload.stakeTxbMinor}`
    return run('bet', actionId, (key) => featuresApi.placeBet(payload.gameId, payload.stakeTxbMinor, key))
  }

  function draw(drawId: string): Promise<void> {
    return run('draw', `draw:${drawId}`, (key) => featuresApi.drawLuckyPrize(drawId, key))
  }

  function clearResult(): void {
    result.value = null
  }

  onMounted(() => void load())

  return {
    overview: readonly(overview),
    result: readonly(result),
    loading: readonly(loading),
    busy: readonly(busy),
    error: readonly(error),
    load,
    checkIn,
    placeBet,
    draw,
    clearResult,
  }
}
