import { useReducedMotion } from 'motion-v'

import { motionDurations, reducedMotionTransition } from './motionPresets'

export function useMotionPreferences() {
  const reducedMotion = useReducedMotion()

  function transition(duration: number = motionDurations.normal) {
    return reducedMotion.value ? reducedMotionTransition : { duration, ease: 'easeOut' as const }
  }

  function offset(distance: number): number {
    return reducedMotion.value ? 0 : distance
  }

  return { reducedMotion, transition, offset }
}
