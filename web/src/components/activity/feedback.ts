import type { ActivityResult } from '@/api/features'

export type ActivityNotification = 'success' | 'warning'
export type ActivityResultSummary = Pick<ActivityResult, 'kind' | 'outcome'>

export function activityNotification(result: Pick<ActivityResult, 'outcome'>): ActivityNotification {
  return result.outcome === 'loss' ? 'warning' : 'success'
}

export function isSuccessfulBet(result: ActivityResultSummary | null | undefined): boolean {
  return result?.kind === 'bet' && result.outcome === 'win'
}
