import type { components } from './generated'

type DeepReadonly<T> = T extends (...args: never[]) => unknown
  ? T
  : T extends readonly (infer Item)[]
    ? readonly DeepReadonly<Item>[]
    : T extends object
      ? { readonly [Key in keyof T]: DeepReadonly<T[Key]> }
      : T

// Public request and response contracts come directly from the generated
// OpenAPI components. UI-only view models are defined separately below.
export type ID = components['schemas']['ID']
export type DecimalString = components['schemas']['DecimalInteger']
export type RFC3339 = components['schemas']['Timestamp']
export type Money = DeepReadonly<components['schemas']['Money']>
export type ApiErrorBody = components['schemas']['ApiError']
export type OnboardingStep = components['schemas']['OnboardingStep']
export type User = DeepReadonly<components['schemas']['User']>
export type AuthState = DeepReadonly<components['schemas']['AuthState']>
export type Session = AuthState
export type JoinInvite = DeepReadonly<components['schemas']['JoinInvite']>
export type OnboardingInvites = DeepReadonly<components['schemas']['OnboardingInvites']>

export interface InviteLink extends JoinInvite {
  kind: 'group' | 'channel'
  label: string
  joined: boolean
}

// The client folds the server membership response into the current session so
// onboarding components can consume one stable view model.
export interface MembershipState {
  session: AuthState
  groupJoined: boolean
  channelJoined: boolean
  complete: boolean
}

export type ResetCadence = components['schemas']['ResetStrategy']
export type SquadProduct = DeepReadonly<components['schemas']['SquadProduct']>
export type SquadProductWrite = components['schemas']['SquadProductWrite']
export type Combo = DeepReadonly<components['schemas']['Combo']>
export type Catalog = DeepReadonly<components['schemas']['Catalog']>
export type EntitlementStatus = components['schemas']['EntitlementStatus']
export type Purchase = DeepReadonly<components['schemas']['Purchase']>
export type TopNode = DeepReadonly<components['schemas']['TopNode']>
export type UsageStatistics = DeepReadonly<components['schemas']['Statistics']>
export type Dashboard = DeepReadonly<components['schemas']['Dashboard']>
export type PaymentProvider = components['schemas']['PaymentProvider']
export type PaymentStatus = components['schemas']['PaymentStatus']
export type PaymentMethod = DeepReadonly<components['schemas']['PaymentMethod']>
export type PaymentOrder = DeepReadonly<components['schemas']['PaymentOrder']>
export type AdminPaymentOrder = DeepReadonly<components['schemas']['AdminPaymentOrder']>
export type AdminEntitlement = DeepReadonly<components['schemas']['AdminEntitlement']>
export type Refund = DeepReadonly<components['schemas']['Refund']>
export type LedgerEntry = DeepReadonly<components['schemas']['LedgerEntry']>
export type AdminSetting = DeepReadonly<components['schemas']['AdminSetting']>
export type AdminUserSummary = DeepReadonly<components['schemas']['AdminUserSummary']>
export type BackupRecord = DeepReadonly<components['schemas']['BackupRun']>
export type AuditEvent = DeepReadonly<components['schemas']['AuditEvent']>
export type JobRecord = DeepReadonly<components['schemas']['Job']>
export type HealthState = DeepReadonly<components['schemas']['RuntimeReadiness']>

export interface Paginated<T> {
  items: T[]
  page?: components['schemas']['PageInfo']
}

export type AdminResource =
  | 'settings'
  | 'combos'
  | 'squad-products'
  | 'users'
  | 'entitlements'
  | 'payments'
  | 'refunds'
  | 'backups'
  | 'jobs'
  | 'audit-events'
