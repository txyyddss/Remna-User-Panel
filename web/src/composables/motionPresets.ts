export const motionDurations = {
  instant: 0.11,
  fast: 0.16,
  normal: 0.2,
  step: 0.24,
  panel: 0.26,
  data: 0.45,
  flip: 0.45,
  celebration: 0.72,
} as const

export const motionSpring = {
  type: 'spring' as const,
  stiffness: 420,
  damping: 34,
  mass: 0.7,
}

export const reducedMotionTransition = {
  duration: 0.08,
  ease: 'easeOut' as const,
}
