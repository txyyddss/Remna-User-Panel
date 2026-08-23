import type { components } from '../generated'

export type NodeCompensationConfig = components['schemas']['NodeCompensationConfig']
export type NodeCompensationConfigWrite = components['schemas']['NodeCompensationConfigWrite']
export type NodeCompensationEvent = components['schemas']['NodeCompensationEvent']
export type NodeCompensationEventPage = components['schemas']['NodeCompensationEventPage']
export type NodeCompensationStatus = NodeCompensationEvent['status']

export interface NodeCompensationReview {
  revision: number
  extensionMinutes?: number
  reason: string
}
