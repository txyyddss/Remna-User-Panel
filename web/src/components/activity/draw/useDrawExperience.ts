import { onMounted, onScopeDispose, readonly, shallowRef, watch } from 'vue'

import type { ActivityResult, LuckyDraw } from '@/api/features'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { localizedError } from '@/i18n'
import { resolveDrawStyle, type DrawPhase, type DrawStyle, type DrawStyleChoice } from './selection'

export function useDrawExperience(play: (drawId: string) => Promise<ActivityResult>) {
  const draw = shallowRef<LuckyDraw | null>(null)
  const style = shallowRef<DrawStyle>('simple')
  const phase = shallowRef<DrawPhase>('running')
  const result = shallowRef<ActivityResult | null>(null)
  const error = shallowRef<string | null>(null)
  const resetToken = shallowRef(0)
  const { reducedMotion } = useMotionPreferences()
  let watchdog: ReturnType<typeof setTimeout> | undefined

  function reveal(): void {
    if (!result.value) return
    phase.value = 'receipt'
  }

  async function submit(): Promise<void> {
    const current = draw.value
    if (!current) return
    phase.value = 'running'
    error.value = null
    try {
      const received = await play(current.id)
      if (draw.value?.id !== current.id) return
      result.value = received
      phase.value = reducedMotion.value || globalThis.document?.hidden || !received.prizeId
        ? 'receipt' : 'settling'
    } catch (caught) {
      if (draw.value?.id !== current.id) return
      error.value = localizedError(caught, 'errors.activityFailed')
      phase.value = 'error'
    }
  }

  function begin(selectedDraw: LuckyDraw, choice: DrawStyleChoice): void {
    if (draw.value) return
    draw.value = selectedDraw
    style.value = resolveDrawStyle(choice)
    result.value = null
    void submit()
  }

  function retry(): void {
    if (phase.value === 'error' && draw.value) void submit()
  }

  function close(): void {
    if (phase.value !== 'receipt' && phase.value !== 'error') return
    draw.value = null
    result.value = null
    error.value = null
    resetToken.value += 1
  }

  function handleVisibility(): void {
    if (globalThis.document?.hidden && result.value) reveal()
  }

  watch(phase, (value) => {
    if (watchdog) clearTimeout(watchdog)
    if (value === 'settling') watchdog = setTimeout(reveal, 6000)
  })
  watch(reducedMotion, (value) => {
    if (value && result.value) reveal()
  })
  onMounted(() => globalThis.document?.addEventListener('visibilitychange', handleVisibility))
  onScopeDispose(() => {
    if (watchdog) clearTimeout(watchdog)
    globalThis.document?.removeEventListener('visibilitychange', handleVisibility)
  })

  return {
    draw: readonly(draw), style: readonly(style), phase: readonly(phase),
    result: readonly(result), error: readonly(error), resetToken: readonly(resetToken),
    begin, retry, reveal, close,
  }
}
