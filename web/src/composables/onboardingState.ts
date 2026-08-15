import type { OnboardingStep } from '@/api/types'

export function onboardingStepAfterMembership(state: OnboardingStep, membershipsComplete: boolean): OnboardingStep {
  if (state === 'agreement' || state === 'complete') return state
  return membershipsComplete ? 'username' : 'membership'
}
