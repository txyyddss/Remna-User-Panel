import { computed, onMounted, onScopeDispose, shallowRef, watch } from 'vue'

import { previewPrizes, type DrawPresenterProps } from '../selection'
import { useMotionPreferences } from '@/composables/useMotionPreferences'

export function usePreview(props: DrawPresenterProps) {
  const offset = shallowRef(0)
  const { reducedMotion } = useMotionPreferences()
  let interval: ReturnType<typeof setInterval> | undefined
  function stop(): void {
    if (interval) clearInterval(interval)
    interval = undefined
  }
  onMounted(() => {
    if (props.prizes.length > 8 && !props.result && !reducedMotion.value) {
      interval = setInterval(() => { offset.value = (offset.value + 1) % props.prizes.length }, 700)
    }
  })
  watch(() => props.result, stop)
  watch(reducedMotion, (value) => { if (value) stop() })
  onScopeDispose(stop)
  return computed(() => previewPrizes(props.prizes, offset.value, props.result))
}
