import type {
  AdminResource,
  ApiErrorBody,
  Catalog,
  Dashboard,
  InviteLink,
  LedgerEntry,
  MembershipState,
  Paginated,
  PaymentMethod,
  PaymentOrder,
  PaymentProvider,
  Purchase,
  Session,
  SquadProduct,
  SquadProductWrite,
} from './types'
import type { components } from './generated'

type TelegramAuthRequest = components['schemas']['TelegramAuthRequest']
type PurchaseRequest = components['schemas']['PurchaseRequest']
type PaymentOrderRequest = components['schemas']['PaymentOrderRequest']
type BalanceAdjustmentRequest = components['schemas']['BalanceAdjustmentRequest']
type ReasonRequest = components['schemas']['ReasonRequest']
type RefundRequest = components['schemas']['RefundRequest']
type GeneratedSquadProductWrite = components['schemas']['SquadProductWrite']

type QueryValue = string | number | boolean | undefined

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown
  query?: Record<string, QueryValue>
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly details?: Record<string, string>
  readonly requestId?: string

  constructor(status: number, body: ApiErrorBody) {
    super(body.message)
    this.name = 'ApiError'
    this.status = status
    this.code = body.code
    this.details = body.details
    this.requestId = body.requestId
  }
}

function createUrl(path: string, query?: Record<string, QueryValue>): string {
  if (!query) return path
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined) params.set(key, String(value))
  }
  const encoded = params.toString()
  return encoded ? `${path}?${encoded}` : path
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  if (options.body !== undefined) headers.set('Content-Type', 'application/json')

  const response = await fetch(createUrl(path, options.query), {
    ...options,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
    credentials: 'include',
    headers,
  })

  if (response.status === 204) return undefined as T

  const contentType = response.headers.get('content-type') ?? ''
  const payload = contentType.includes('application/json')
    ? await response.json() as unknown
    : await response.text()

  if (!response.ok) {
    const body = typeof payload === 'object' && payload !== null
      ? payload as ApiErrorBody
      : { code: 'HTTP_ERROR', message: String(payload || response.statusText) }
    throw new ApiError(response.status, body)
  }

  if (typeof payload === 'object' && payload !== null && 'data' in payload) {
    return (payload as { data: T }).data
  }
  return payload as T
}

export const api = {
  authTelegram: (initData: string) => request<Session>('/api/v1/auth/telegram', {
    method: 'POST',
    body: { initData } satisfies TelegramAuthRequest,
  }),
  getMe: () => request<Session>('/api/v1/me'),
  createInvites: async () => {
    const response = await request<{ group: Pick<InviteLink, 'url' | 'expiresAt'>; channel: Pick<InviteLink, 'url' | 'expiresAt'> }>('/api/v1/onboarding/invites', { method: 'POST' })
    return { invites: [
      { ...response.group, kind: 'group' as const, label: 'TX private group', joined: false },
      { ...response.channel, kind: 'channel' as const, label: 'TX updates channel', joined: false },
    ] }
  },
  checkMembership: async () => {
    const response = await request<{ user: Session['user']; groupJoined: boolean; channelJoined: boolean; complete: boolean }>('/api/v1/onboarding/membership/check', { method: 'POST' })
    const session: Session = { authenticated: true, user: response.user }
    return {
      session,
      groupJoined: response.groupJoined,
      channelJoined: response.channelJoined,
      complete: response.complete,
    } satisfies MembershipState
  },
  setUsername: (username: string) => request<Session>('/api/v1/onboarding/username', {
    method: 'PUT',
    body: { username },
  }),
  acceptAgreement: () => request<Session>('/api/v1/onboarding/agreement', {
    method: 'POST',
    body: { accepted: true },
  }),
  getDashboard: () => request<Dashboard>('/api/v1/dashboard'),
  getCatalog: () => request<Catalog>('/api/v1/catalog'),
  createPurchase: (comboId: string, squadProductIds: string[]) => request<Purchase>('/api/v1/purchases', {
    method: 'POST',
    body: { comboId, addonSquadProductIds: squadProductIds } satisfies PurchaseRequest,
  }),
  getPurchases: () => request<Paginated<Purchase>>('/api/v1/purchases'),
  revokeSubscription: () => request<{ subscriptionUrl: string }>('/api/v1/subscription/revoke', {
    method: 'POST',
  }),
  getBalance: () => request<{ balance: Dashboard['balance']; paymentMethods: PaymentMethod[] }>('/api/v1/balance'),
  getLedger: (cursor?: string) => request<Paginated<LedgerEntry>>('/api/v1/ledger', { query: { cursor } }),
  createPaymentOrder: (provider: PaymentProvider, txbMinorUnits: string) => request<PaymentOrder>('/api/v1/payments/orders', {
    method: 'POST',
    body: { provider, txbMinor: txbMinorUnits } satisfies PaymentOrderRequest,
  }),
  getPaymentOrder: (id: string) => request<PaymentOrder>(`/api/v1/payments/orders/${encodeURIComponent(id)}`),
  getAdminResource: <T>(resource: AdminResource, query?: Record<string, QueryValue>) =>
    request<T>(`/api/v1/admin/${resource}`, { query }),
  updateAdminSetting: <T>(key: string, value: string) =>
    request<T>(`/api/v1/admin/settings/${encodeURIComponent(key)}`, { method: 'PUT', body: { value } }),
  importAdminSquadProducts: () =>
    request<{ items: SquadProduct[] }>('/api/v1/admin/squad-products/import', { method: 'POST' }),
  updateAdminSquadProduct: (id: string, body: SquadProductWrite) =>
    request<SquadProduct>(`/api/v1/admin/squad-products/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: body satisfies GeneratedSquadProductWrite,
    }),
  createAdminResource: <T>(resource: AdminResource, body: unknown) =>
    request<T>(`/api/v1/admin/${resource}`, { method: 'POST', body }),
  updateAdminResource: <T>(resource: AdminResource, id: string, body: unknown) =>
    request<T>(`/api/v1/admin/${resource}/${encodeURIComponent(id)}`, { method: 'PUT', body }),
  deleteAdminResource: (resource: AdminResource, id: string) =>
    request<void>(`/api/v1/admin/${resource}/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  refundPayment: (paymentId: string, reason: string) => request<PaymentOrder>(`/api/v1/admin/payments/${encodeURIComponent(paymentId)}/refund`, {
    method: 'POST',
    body: { reason } satisfies RefundRequest,
  }),
  cancelAdminEntitlement: (entitlementId: string, reason: string) => request<Purchase>(`/api/v1/admin/entitlements/${encodeURIComponent(entitlementId)}/cancel`, {
    method: 'POST',
    body: { reason } satisfies ReasonRequest,
  }),
  retryAdminJob: (jobId: string) => request<void>(`/api/v1/admin/jobs/${encodeURIComponent(jobId)}/retry`, {
    method: 'POST',
  }),
  adjustBalance: (userId: string, amountMinor: string, reason: string) => request<LedgerEntry>(`/api/v1/admin/users/${encodeURIComponent(userId)}/balance-adjustments`, {
    method: 'POST',
    body: { deltaTxbMinor: amountMinor, reason } satisfies BalanceAdjustmentRequest,
  }),
}
