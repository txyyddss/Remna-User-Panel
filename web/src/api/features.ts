import type { Money, PaymentProvider, RFC3339 } from './types'
import { ApiError } from './client'

type QueryValue = string | number | boolean | undefined

interface FeatureRequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown
  query?: Record<string, QueryValue>
}

export type Reward =
  | { kind: 'txb_delta'; txbDeltaMinor: string }
  | { kind: 'coupon_grant'; couponId: string }
  | { kind: 'subscription_extension'; extensionDays: number }
  | { kind: 'none' }

export interface BetGame {
  id: string
  name: string
  icon: string
  description: string
  winChanceBps: number
  minimumStakeMinor: string
  maximumStakeMinor: string
  returnMultiplierBps: number
  enabled: boolean
}

export interface LuckyDraw {
  id: string
  name: string
  description: string
  feeTxbMinor: string
  enabled: boolean
}

export interface ActivitySettings {
  timezone: string
  dailyRewardTxb: string
  dailyRewardTxbMinor: string
}

export interface LuckyDrawPrize {
  id: string
  name: string
  weight: string
  stockRemaining?: number | null
  reward: Reward
}

export interface LuckyDrawAdmin extends LuckyDraw {
  prizes: LuckyDrawPrize[]
  createdAt: RFC3339
  updatedAt: RFC3339
}

export interface LuckyDrawWrite {
  name: string
  description: string
  enabled: boolean
  feeTxbMinor: string
  prizes: LuckyDrawPrize[]
}

export interface ActivityResult {
  id: string
  kind: 'check_in' | 'bet' | 'draw'
  outcome: 'win' | 'loss' | 'complete'
  message: string
  reward: Reward
  stakeTxbMinor?: string
  balanceAfter: Money
  createdAt: RFC3339
}

export interface ActivityOverview {
  balance: Money
  timeZone: string
  checkedInToday: boolean
  dailyRewardTxbMinor: string
  games: BetGame[]
  draws: LuckyDraw[]
  recentResults: ActivityResult[]
}

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
  createdAt: RFC3339
  updatedAt: RFC3339
}

export interface CouponGrant {
  id: string
  sourceType: string
  sourceId: string
  status: string
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
  status: 'draft' | 'active' | 'closed' | 'settling' | 'settled' | 'failed'
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
  delimiter: string
  dataRowCount: number
  malformedRowCount: number
  codeColumn?: string | null
  analysis?: QuestionnaireImportSummary | null
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
  report?: QuestionnaireSettlementReport | null
}

export interface EmbyLibrary {
  id: string
  name: string
}

export interface EmbyRating {
  name: string
  value: number
}

export interface EmbyAccount {
  id: string
  username: string
  status: 'queued' | 'provisioning' | 'active' | 'failed'
  maxParentalRating: number | null
  libraryIds: string[]
  retryable?: boolean
  errorMessage?: string | null
  updatedAt: RFC3339
}

export interface EmbyOverview {
  configured: boolean
  setupPrice: Money
  ratings: EmbyRating[]
  libraries: EmbyLibrary[]
  account: EmbyAccount | null
}

export interface FeaturePaymentMethod {
  id: string
  provider: PaymentProvider
  rail: string
  name: string
  currency: 'CNY' | 'USD' | 'USDT' | 'XTR'
  available: boolean
  note: string
}

export interface FeaturePaymentOrder {
  id: string
  methodId: string
  provider: PaymentProvider
  providerRail: string
  status: 'creating' | 'pending' | 'paid' | 'cancelled' | 'expired' | 'failed' | 'refunded'
  txb: Money
  payableAmount: string
  payableCurrency: 'CNY' | 'USD' | 'USDT' | 'XTR'
  rateSnapshot: string
  rateDirection: 'txb_per_currency' | 'currency_per_txb'
  paymentUrl: string | null
  qrPayload: string | null
  receivingAddress: string | null
  actualCryptoAmount: string | null
  actualCryptoCurrency: 'USDT' | null
  expiresAt: RFC3339
  paidAt: RFC3339 | null
  refundedAt: RFC3339 | null
  cancelledAt: RFC3339 | null
  cancelReason: string
  providerCancelStatus: '' | 'unsupported' | 'requested' | 'cancelled' | 'failed'
  createdAt: RFC3339
  updatedAt: RFC3339
}

export interface DatabaseColumn {
  name: string
  declaredType: string
  nullable: boolean
  primaryKeyPosition: number
  editable: boolean
  sensitive: boolean
}

export interface DatabaseTable {
  name: string
  columns: DatabaseColumn[]
  highRisk: boolean
  supportsRowId: boolean
  warning: string
}

export type DatabaseValue = string | boolean | null | { blobBase64: string }

export interface DatabaseRow {
  key: Record<string, string>
  values: Record<string, DatabaseValue>
  recordHash: string
}

export interface DatabaseRowsPage {
  items: DatabaseRow[]
  nextCursor: string | null
}

export interface DatabaseMutationInput {
  action: 'insert' | 'update' | 'delete'
  table: string
  key?: Record<string, string>
  values?: Record<string, DatabaseValue>
  recordHash?: string
  reason: string
}

export interface DatabaseMutationReview {
  action: DatabaseMutationInput['action']
  table: string
  key?: Record<string, string>
  before?: Record<string, DatabaseValue> | null
  after?: Record<string, DatabaseValue> | null
  changedColumns: string[]
  reviewHash: string
  requiredConfirmation: string
  rescueBackupRequired: true
  warning: string
}

export interface DatabaseMutationResult {
  row?: DatabaseRow
  deleted: boolean
  rescueBackupId: string
}

export interface RestoreOperation {
  id: string
  backupId: string
  status: 'staging' | 'restarting' | 'complete' | 'failed'
  error?: string | null
  createdAt: RFC3339
  updatedAt: RFC3339
}

function featureUrl(path: string, query?: Record<string, QueryValue>): string {
  if (!query) return path
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined) params.set(key, String(value))
  }
  const encoded = params.toString()
  return encoded ? `${path}?${encoded}` : path
}

async function featureRequest<T>(path: string, options: FeatureRequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  const isForm = options.body instanceof FormData
  if (options.body !== undefined && !isForm) headers.set('Content-Type', 'application/json')

  const requestBody: BodyInit | undefined = options.body === undefined
    ? undefined
    : isForm ? options.body as FormData : JSON.stringify(options.body)
  const response = await fetch(featureUrl(path, options.query), {
    ...options,
    body: requestBody,
    credentials: 'include',
    headers,
  })
  if (response.status === 204) return undefined as T
  const contentType = response.headers.get('content-type') ?? ''
  const payload = contentType.includes('application/json')
    ? await response.json() as unknown
    : await response.text()
  if (!response.ok) {
    const errorBody = typeof payload === 'object' && payload !== null
      ? payload as { code: string; message: string; details?: Record<string, string>; requestId?: string }
      : { code: 'HTTP_ERROR', message: String(payload || response.statusText) }
    throw new ApiError(response.status, errorBody)
  }
  if (typeof payload === 'object' && payload !== null && 'data' in payload) {
    return (payload as { data: T }).data
  }
  return payload as T
}

export const featuresApi = {
  getActivity: () => featureRequest<ActivityOverview>('/api/v1/activity'),
  checkIn: (idempotencyKey: string) => featureRequest<ActivityResult>('/api/v1/activity/check-ins', { method: 'POST', headers: { 'Idempotency-Key': idempotencyKey } }),
  placeBet: (gameId: string, stakeTxbMinor: string, idempotencyKey: string) => featureRequest<ActivityResult>('/api/v1/activity/bets', {
    method: 'POST', body: { gameId, stakeTxbMinor }, headers: { 'Idempotency-Key': idempotencyKey },
  }),
  drawLuckyPrize: (drawId: string, idempotencyKey: string) => featureRequest<ActivityResult>('/api/v1/activity/draws', {
    method: 'POST', body: { drawId }, headers: { 'Idempotency-Key': idempotencyKey },
  }),
  getCouponWallet: () => featureRequest<{ items: CouponGrant[] }>('/api/v1/coupons/wallet'),
  redeemCoupon: (code: string, idempotencyKey: string) => featureRequest<CouponRedemption>('/api/v1/coupons/redeem', {
    method: 'POST', body: { code }, headers: { 'Idempotency-Key': idempotencyKey },
  }),
  getActiveQuestionnaire: () => featureRequest<ActiveQuestionnaire | null>('/api/v1/questionnaires/active'),
  joinQuestionnaire: (questionnaireId: string, idempotencyKey: string) => featureRequest<QuestionnaireParticipation>(
    `/api/v1/questionnaires/${encodeURIComponent(questionnaireId)}/participation`, { method: 'POST', headers: { 'Idempotency-Key': idempotencyKey } },
  ),
  getEmby: () => featureRequest<EmbyOverview>('/api/v1/emby/account'),
  setupEmby: (body: { password: string; maxParentalRating: number | null; libraryIds: string[] }) =>
    featureRequest<EmbyAccount>('/api/v1/emby/setup', { method: 'POST', body }),
  updateEmbyPreferences: (body: { maxParentalRating: number | null; libraryIds: string[] }) =>
    featureRequest<EmbyAccount>('/api/v1/emby/preferences', { method: 'PUT', body }),
  changeEmbyPassword: (password: string) => featureRequest<void>('/api/v1/emby/password', {
    method: 'PUT', body: { password },
  }),
  getAdminActivityGames: () => featureRequest<{ items: BetGame[] }>('/api/v1/admin/activity-games'),
  getAdminActivitySettings: () => featureRequest<ActivitySettings>('/api/v1/admin/activity-settings'),
  saveAdminActivitySettings: (body: { timezone: string; dailyRewardTxb: string }) =>
    featureRequest<ActivitySettings>('/api/v1/admin/activity-settings', { method: 'PUT', body }),
  saveAdminActivityGame: (id: string | null, body: Omit<BetGame, 'id'>) => featureRequest<BetGame>(
    id ? `/api/v1/admin/activity-games/${encodeURIComponent(id)}` : '/api/v1/admin/activity-games',
    { method: id ? 'PUT' : 'POST', body },
  ),
  getAdminLuckyDraws: () => featureRequest<{ items: LuckyDrawAdmin[] }>('/api/v1/admin/lucky-draw'),
  saveAdminLuckyDraw: (id: string | null, body: LuckyDrawWrite) => featureRequest<LuckyDrawAdmin>(
    id ? `/api/v1/admin/lucky-draw/${encodeURIComponent(id)}` : '/api/v1/admin/lucky-draw',
    { method: id ? 'PUT' : 'POST', body },
  ),
  getAdminCoupons: () => featureRequest<{ items: CouponDefinition[] }>('/api/v1/admin/coupons'),
  saveAdminCoupon: (id: string | null, body: Partial<CouponDefinition>) => featureRequest<CouponDefinition>(
    id ? `/api/v1/admin/coupons/${encodeURIComponent(id)}` : '/api/v1/admin/coupons',
    { method: id ? 'PUT' : 'POST', body },
  ),
  deactivateAdminCoupon: (id: string) => featureRequest<void>(`/api/v1/admin/coupons/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  getAdminQuestionnaires: () => featureRequest<{ items: QuestionnaireAdminRecord[] }>('/api/v1/admin/questionnaires'),
  saveAdminQuestionnaire: (id: string | null, body: Partial<QuestionnaireAdminRecord>) => featureRequest<QuestionnaireAdminRecord>(
    id ? `/api/v1/admin/questionnaires/${encodeURIComponent(id)}` : '/api/v1/admin/questionnaires',
    { method: id ? 'PUT' : 'POST', body },
  ),
  activateAdminQuestionnaire: (id: string) => featureRequest<QuestionnaireAdminRecord>(
    `/api/v1/admin/questionnaires/${encodeURIComponent(id)}/activate`, { method: 'POST' },
  ),
  closeAdminQuestionnaire: (id: string) => featureRequest<void>(
    `/api/v1/admin/questionnaires/${encodeURIComponent(id)}`, { method: 'DELETE' },
  ),
  previewQuestionnaireCsv: (id: string, file: File) => {
    const body = new FormData()
    body.set('file', file)
    return featureRequest<QuestionnaireImportPreview>(`/api/v1/admin/questionnaires/${encodeURIComponent(id)}/imports`, {
      method: 'POST', body,
    })
  },
  analyzeQuestionnaireCsv: (id: string, uploadId: string, codeColumn: string) =>
    featureRequest<QuestionnaireImportSummary>(`/api/v1/admin/questionnaires/${encodeURIComponent(id)}/imports/${encodeURIComponent(uploadId)}/analyze`, {
      method: 'POST', body: { codeColumn },
    }),
  settleQuestionnaireCsv: (id: string, uploadId: string) =>
    featureRequest<QuestionnaireImportPreview>(`/api/v1/admin/questionnaires/${encodeURIComponent(id)}/imports/${encodeURIComponent(uploadId)}/settle`, {
      method: 'POST',
    }),
  getQuestionnaireImportState: (id: string, uploadId: string) =>
    featureRequest<QuestionnaireImportState>(`/api/v1/admin/questionnaires/${encodeURIComponent(id)}/imports/${encodeURIComponent(uploadId)}`),
  getAdminEmbyAccounts: () => featureRequest<{ items: EmbyAccount[] }>('/api/v1/admin/emby-accounts'),
  retryAdminEmbyAccount: (id: string) => featureRequest<EmbyAccount>(`/api/v1/admin/emby-accounts/${encodeURIComponent(id)}/retry`, {
    method: 'POST',
  }),
  downloadBackup: async (id: string) => {
    const response = await fetch(`/api/v1/admin/backups/${encodeURIComponent(id)}/download`, { credentials: 'include' })
    if (!response.ok) throw new Error(`Backup download failed with status ${response.status}.`)
    return response.blob()
  },
  restoreBackup: (id: string, reason: string, confirmation: string) => featureRequest<RestoreOperation>(`/api/v1/admin/backups/${encodeURIComponent(id)}/restore`, {
    method: 'POST', body: { reason, confirmation },
  }),
  getRestoreStatus: (id: string) => featureRequest<RestoreOperation>(`/api/v1/admin/restores/${encodeURIComponent(id)}`),
  getDatabaseTables: () => featureRequest<{ items: DatabaseTable[] }>('/api/v1/admin/database/tables'),
  getDatabaseRows: (table: string, cursor?: string) => featureRequest<DatabaseRowsPage>(
    `/api/v1/admin/database/tables/${encodeURIComponent(table)}/rows`, { query: { cursor, limit: 50 } },
  ),
  reviewDatabaseMutation: (body: DatabaseMutationInput) => featureRequest<DatabaseMutationReview>('/api/v1/admin/database/mutations/review', {
    method: 'POST', body,
  }),
  applyDatabaseMutation: (body: DatabaseMutationInput & { reviewHash: string; confirmation: string }) =>
    featureRequest<DatabaseMutationResult>('/api/v1/admin/database/mutations', { method: 'POST', body }),
}
