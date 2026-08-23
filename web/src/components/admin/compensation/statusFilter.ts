import type { NodeCompensationStatus } from '@/api/contracts/compensation'

const eventStatuses = ['observing', 'pending_review', 'queued', 'dismissed', 'ineligible'] as const satisfies readonly NodeCompensationStatus[]

export const compensationStatusChoices = ['all', ...eventStatuses] as const
export type CompensationStatusChoice = typeof compensationStatusChoices[number]

export function fromCompensationStatus(status: NodeCompensationStatus | ''): CompensationStatusChoice {
  return status || 'all'
}

export function toCompensationStatus(choice: CompensationStatusChoice): NodeCompensationStatus | '' {
  return choice === 'all' ? '' : choice
}
