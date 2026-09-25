import { onMounted, onScopeDispose, shallowRef, watch, type Ref } from 'vue'
import { gsap } from 'gsap'

import { selectedPreviewIndex, settlingStepCount, type DrawPresenterProps } from '../selection'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { usePreview } from './usePreview'

export function useSelectionTicker(
  props: DrawPresenterProps,
  scope: Readonly<Ref<HTMLElement | null>>,
  finished: () => void,
) {
  const visible = usePreview(props)
  const { reducedMotion } = useMotionPreferences()
  const active = shallowRef(0)
  const cursor = { value: 0 }
  let context: gsap.Context | undefined
  let loop: gsap.core.Tween | undefined

  function index(): void {
    active.value = Math.floor(cursor.value) % Math.max(visible.value.length, 1)
  }
  function settle(): void {
    loop?.kill()
    const target = selectedPreviewIndex(visible.value, props.result)
    if (target < 0) { finished(); return }
    const count = visible.value.length
    const start = Math.floor(cursor.value)
    const steps = settlingStepCount(start, target, count)
    context?.add(() => gsap.to(cursor, {
      value: start + steps, duration: 2.1, ease: 'power3.out', onUpdate: index, onComplete: finished,
    }))
  }
  onMounted(() => {
    context = gsap.context(() => undefined, scope.value ?? undefined)
    if (props.result) { settle(); return }
    if (reducedMotion.value) return
    context.add(() => {
      loop = gsap.to(cursor, { value: 8, duration: 1.35, repeat: -1, ease: 'none', onUpdate: index })
    })
  })
  watch(() => props.result, (value) => { if (value && context) settle() })
  watch(reducedMotion, (value) => { if (value) loop?.kill() })
  onScopeDispose(() => context?.revert())
  return { visible, active }
}
