export const maxExtensionMinutes = 5_256_000

export type DurationUnit = 'minutes' | 'hours' | 'days'

export interface DurationDraft {
  amount: number
  unit: DurationUnit
}

const minuteFactors: Record<DurationUnit, number> = {
  minutes: 1,
  hours: 60,
  days: 24 * 60,
}

export function durationMinutes(value: DurationDraft): number {
  return value.amount * minuteFactors[value.unit]
}

export function maxDurationAmount(unit: DurationUnit): number {
  return Math.floor(maxExtensionMinutes / minuteFactors[unit])
}

export function validDurationDraft(value: DurationDraft): boolean {
  return Number.isInteger(value.amount) && value.amount >= 1 && value.amount <= maxDurationAmount(value.unit)
}
