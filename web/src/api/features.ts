import { request, requestBlob } from './http'
import type { ActivityOverview, ActivityResult, ActivitySettings, BetGame, LuckyDrawAdmin, LuckyDrawWrite } from './contracts/activity'
import type { ActiveQuestionnaire, CouponDefinition, CouponGrant, CouponRedemption, QuestionnaireAdminRecord, QuestionnaireImportPreview, QuestionnaireImportState, QuestionnaireImportSummary, QuestionnaireParticipation } from './contracts/community'
import type { EmbyAccount, EmbyOverview } from './contracts/commerce'
import type { AdminStatistics, DatabaseMutationInput, DatabaseMutationResult, DatabaseMutationReview, DatabaseQueryInput, DatabaseRowsPage, DatabaseTable, OnboardingBundle, OnboardingLocalizedContent, PublishedOnboarding, RestoreOperation, StatisticsQuery } from './contracts/admin'

export type * from './contracts/activity'
export type * from './contracts/community'
export type * from './contracts/commerce'
export type * from './contracts/admin'

const featureRequest = request

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
  setupEmby: (body: { password: string; maxParentalRating: number | null; disabledLibraryIds: string[] }) =>
    featureRequest<EmbyAccount>('/api/v1/emby/setup', { method: 'POST', body }),
  updateEmbyPreferences: (body: { maxParentalRating: number | null; disabledLibraryIds: string[] }) =>
    featureRequest<EmbyAccount>('/api/v1/emby/preferences', { method: 'PUT', body }),
  changeEmbyPassword: (password: string) => featureRequest<void>('/api/v1/emby/password', {
    method: 'PUT', body: { password },
  }),
  getAdminActivityGames: () => featureRequest<{ items: BetGame[] }>('/api/v1/admin/activity-games'),
  getAdminActivitySettings: () => featureRequest<ActivitySettings>('/api/v1/admin/activity-settings'),
  saveAdminActivitySettings: (body: { timezone: string; groupMessageThreshold: number }) =>
    featureRequest<ActivitySettings>('/api/v1/admin/activity-settings', { method: 'PUT', body }),
  saveAdminActivityGame: (id: string | null, body: Omit<BetGame, 'id'>) => featureRequest<BetGame>(
    id ? `/api/v1/admin/activity-games/${encodeURIComponent(id)}` : '/api/v1/admin/activity-games',
    { method: id ? 'PUT' : 'POST', body },
  ),
  deleteAdminActivityGame: (id: string) => featureRequest<void>(`/api/v1/admin/activity-games/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  getAdminActivityGameStatistics: (id: string, query: StatisticsQuery = {}) =>
    featureRequest<AdminStatistics>(`/api/v1/admin/activity-games/${encodeURIComponent(id)}/statistics`, { query: { from: query.from, to: query.to, bucket: query.bucket, timeZone: query.timeZone } }),
  getAdminLuckyDraws: () => featureRequest<{ items: LuckyDrawAdmin[] }>('/api/v1/admin/lucky-draw'),
  saveAdminLuckyDraw: (id: string | null, body: LuckyDrawWrite) => featureRequest<LuckyDrawAdmin>(
    id ? `/api/v1/admin/lucky-draw/${encodeURIComponent(id)}` : '/api/v1/admin/lucky-draw',
    { method: id ? 'PUT' : 'POST', body },
  ),
  deleteAdminLuckyDraw: (id: string) => featureRequest<void>(`/api/v1/admin/lucky-draw/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  getAdminLuckyDrawStatistics: (id: string, query: StatisticsQuery = {}) =>
    featureRequest<AdminStatistics>(`/api/v1/admin/lucky-draw/${encodeURIComponent(id)}/statistics`, { query: { from: query.from, to: query.to, bucket: query.bucket, timeZone: query.timeZone } }),
  getAdminCoupons: () => featureRequest<{ items: CouponDefinition[] }>('/api/v1/admin/coupons'),
  getAdminComboStatistics: (id: string, query: StatisticsQuery = {}) =>
    featureRequest<AdminStatistics>(`/api/v1/admin/combos/${encodeURIComponent(id)}/statistics`, { query: { from: query.from, to: query.to, bucket: query.bucket, timeZone: query.timeZone } }),
  getAdminSquadStatistics: (id: string, query: StatisticsQuery = {}) =>
    featureRequest<AdminStatistics>(`/api/v1/admin/squad-products/${encodeURIComponent(id)}/statistics`, { query: { from: query.from, to: query.to, bucket: query.bucket, timeZone: query.timeZone } }),
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
    `/api/v1/admin/questionnaires/${encodeURIComponent(id)}/close`, { method: 'POST' },
  ),
  deleteAdminQuestionnaire: (id: string) => featureRequest<void>(`/api/v1/admin/questionnaires/${encodeURIComponent(id)}`, { method: 'DELETE' }),
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
    return requestBlob(`/api/v1/admin/backups/${encodeURIComponent(id)}/download`)
  },
  restoreBackup: (id: string, reason: string, confirmation: string) => featureRequest<RestoreOperation>(`/api/v1/admin/backups/${encodeURIComponent(id)}/restore`, {
    method: 'POST', body: { reason, confirmation },
  }),
  deleteBackup: (id: string) => featureRequest<void>(`/api/v1/admin/backups/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  getRestoreStatus: (id: string) => featureRequest<RestoreOperation>(`/api/v1/admin/restores/${encodeURIComponent(id)}`),
  getDatabaseTables: () => featureRequest<{ items: DatabaseTable[] }>('/api/v1/admin/database/tables'),
  getDatabaseRows: (table: string, cursor?: string) => featureRequest<DatabaseRowsPage>(
    `/api/v1/admin/database/tables/${encodeURIComponent(table)}/rows`, { query: { cursor, limit: 50 } },
  ),
  queryDatabaseRows: (table: string, body: DatabaseQueryInput) => featureRequest<DatabaseRowsPage>(
    `/api/v1/admin/database/tables/${encodeURIComponent(table)}/query`, { method: 'POST', body },
  ),
  getPublishedOnboarding: (locale: string) => featureRequest<PublishedOnboarding>('/api/v1/onboarding/content', { query: { locale } }),
  getAdminOnboardingBundle: (kind: 'welcome' | 'agreements') => featureRequest<OnboardingBundle>(`/api/v1/admin/onboarding/content/${kind}`),
  saveAdminOnboardingDraft: (kind: 'welcome' | 'agreements', draftRevision: number, content: OnboardingLocalizedContent) =>
    featureRequest<OnboardingBundle>(`/api/v1/admin/onboarding/content/${kind}/draft`, { method: 'PUT', body: { draftRevision, content } }),
  publishAdminOnboarding: (kind: 'welcome' | 'agreements', draftRevision: number) =>
    featureRequest<OnboardingBundle>(`/api/v1/admin/onboarding/content/${kind}/publish`, { method: 'POST', body: { draftRevision } }),
  reviewDatabaseMutation: (body: DatabaseMutationInput) => featureRequest<DatabaseMutationReview>('/api/v1/admin/database/mutations/review', {
    method: 'POST', body,
  }),
  applyDatabaseMutation: (body: DatabaseMutationInput & { reviewHash: string; confirmation: string }) =>
    featureRequest<DatabaseMutationResult>('/api/v1/admin/database/mutations', { method: 'POST', body }),
}
