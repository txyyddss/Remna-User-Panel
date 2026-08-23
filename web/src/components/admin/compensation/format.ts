import type { NodeCompensationEvent } from '@/api/contracts/compensation'

export function multiplierFactor(bps: number): string {
  return (bps / 10_000).toFixed(2)
}

export function durationParts(seconds: number | null): { hours: number; minutes: number } {
  const totalMinutes = Math.floor(Math.max(0, seconds ?? 0) / 60)
  return { hours: Math.floor(totalMinutes / 60), minutes: totalMinutes % 60 }
}

export function eventExtension(event: NodeCompensationEvent): number {
  return event.finalExtensionMinutes ?? event.proposedExtensionMinutes ?? 1
}
