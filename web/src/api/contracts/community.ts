import type { RFC3339 } from '../types'

export interface CouponDefinition {
  id: string
  code: string
  name: string
  kind: 'purchase_recurring' | 'purchase_once' | 'balance_add' | 'balance_multiply'
  discountMode?: 'fixed' | 'percent'
  valueMinorOrBps: string
  percentCapMinor?: string | null
  eligibleComboIds: string[]
  eligibleSquadIds: string[]
  expiresAt: RFC3339 | null
  globalUseLimit: number | null
  perUserUseLimit: number | null
  active: boolean
  usageCount: number
  createdAt: RFC3339
  updatedAt: RFC3339
}

export interface CouponGrant {
  id: string
  sourceType: string
  sourceId: string
  status: 'active' | 'consumed' | 'expired'
  useCount: number
  createdAt: RFC3339
  consumedAt: RFC3339 | null
  coupon: CouponDefinition
}

export interface CouponRedemption {
  id: string
  coupon: CouponDefinition
  grant: CouponGrant | null
  balanceDeltaMinor: string
  balanceAfterMinor: string
  idempotencyKey: string
  replayed: boolean
  createdAt: RFC3339
}

export interface QuestionnaireParticipation {
  id: string
  questionnaireId: string
  validationCode: string
  awardedAt: RFC3339 | null
  createdAt: RFC3339
}

export interface ActiveQuestionnaire {
  id: string
  title: string
  description: string
  formUrl: string
  rewardTxbMinor: string
  closesAt: RFC3339 | null
  participation: QuestionnaireParticipation | null
}

export interface QuestionnaireAdminRecord extends Omit<ActiveQuestionnaire, 'participation'> {
  status: 'draft' | 'active' | 'closed' | 'settling' | 'settled'
  participantCount: number
  rewardedCount: number
  createdAt: RFC3339
  updatedAt: RFC3339
}

export interface QuestionnaireImportPreview {
  id: string
  questionnaireId: string
  status: 'preview' | 'queued' | 'processing' | 'settled' | 'failed'
  headers: string[]
  sampleRows: string[][]
  delimiter: 'comma' | 'semicolon' | 'tab'
  dataRowCount: number
  malformedRowCount: number
  codeColumn?: string
  analysis?: QuestionnaireImportSummary
  idempotencyKey?: string
  createdAt: RFC3339
  updatedAt: RFC3339
}

export interface QuestionnaireImportSummary {
  importId: string
  questionnaireId: string
  codeColumn: string
  matchedCount: number
  duplicateCount: number
  unknownCount: number
  malformedCount: number
  alreadyAwardedCount: number
}

export interface QuestionnaireSettlementReport {
  importId: string
  questionnaireId: string
  matchedCount: number
  duplicateCount: number
  unknownCount: number
  malformedCount: number
  alreadyAwardedCount: number
  rewardedCount: number
  rewardTxbMinor: string
  settledAt: RFC3339
  replayed: boolean
}

export interface QuestionnaireImportState {
  preview: QuestionnaireImportPreview
  report?: QuestionnaireSettlementReport
}
