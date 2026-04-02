let BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'
if (BASE_URL !== '/api/v1' && !BASE_URL.endsWith('/api/v1')) {
    BASE_URL = BASE_URL.replace(/\/$/, '') + '/api/v1'
}

import type { User, Subscription, Combo, Order, OrderDetail, OrderEvent, CreditLog, JellyfinAccount, AppConfig, BandwidthEntry, DeviceEntry, HistoryEntry, Squad, PaginatedResponse, SubInfo, SubKeys } from '@/types'

function getInitData(): string {
    return window.Telegram?.WebApp?.initData || ''
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const initData = getInitData()
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(initData ? { 'X-Telegram-Init-Data': initData } : {}),
        ...(options.headers as Record<string, string> || {}),
    }

    const resp = await fetch(`${BASE_URL}${path}`, {
        ...options,
        headers,
    })

    const json = await resp.json()
    if (!resp.ok || (typeof json.code === 'number' && json.code >= 400)) {
        throw new Error(json.message || 'Request failed')
    }
    return json.data as T
}

export const api = {
    // Auth
    getMe: () => request<{ user: User; subscription?: Subscription; jellyfin?: JellyfinAccount; config?: AppConfig }>('/user/me'),

    // Credit
    getCreditBalance: () => request<{ balance: number; name: string }>('/credit/balance'),
    creditSignup: () => request<{ value: number; new_balance: number; auto_delete: boolean }>('/credit/signup', { method: 'POST' }),
    creditBet: (amount: number) => request<{ result: number; new_balance: number }>('/credit/bet', { method: 'POST', body: JSON.stringify({ amount }) }),
    getCreditHistory: (limit = 20, offset = 0) => request<CreditLog[]>(`/credit/history?limit=${limit}&offset=${offset}`),

    // Combos
    listCombos: () => request<Combo[]>('/combos'),

    // Subscription
    purchaseCombo: (data: { combo_uuid: string; payment_method: string; payment_type?: string; auto_renew?: boolean; use_txb?: boolean; discount_rmb?: number }) => request<{ order_uuid: string; payment_url?: string; amount?: number; qrcode_url?: string }>('/subscribe', { method: 'POST', body: JSON.stringify(data) }),
    getSubInfo: () => request<{ user: SubInfo; subscription: Subscription }>('/sub-info'),
    getSubKeys: () => request<SubKeys>('/sub-keys'),

    // Payment
    createPayment: (data: { amount: number; payment_method: string }) => request<{ order_uuid: string; payment_url?: string; qrcode_url?: string }>('/payment/create', { method: 'POST', body: JSON.stringify(data) }),
    getOrders: (limit = 20, offset = 0) => request<Order[]>(`/orders?limit=${limit}&offset=${offset}`),
    getOrder: (uuid: string) => request<Order>(`/orders/${uuid}`),

    // VPN Info
    getBandwidth: () => request<BandwidthEntry[]>('/vpn/bandwidth'),
    getDevices: () => request<DeviceEntry[]>('/vpn/devices'),
    getIPs: () => request<any>('/vpn/ips'),
    getSubHistory: () => request<HistoryEntry[]>('/vpn/history'),

    // Squads
    getExternalSquads: () => request<Squad[]>('/squads/external'),
    updateExternalSquad: (squadUUID: string) => request<void>('/squads/external', { method: 'PUT', body: JSON.stringify({ squad_uuid: squadUUID }) }),

    // IP Change
    changeIP: (data: any = {}) => request<{ order_uuid?: string; payment_url?: string; amount?: number; message?: string }>('/ip/change', { method: 'POST', body: JSON.stringify(data) }),
    getIPStatus: () => request<{ recent_changes: any[]; txb_price: number; rmb_price: number }>('/ip/status'),

    // Jellyfin
    purchaseJellyfin: (data: { months: number; payment_method: string; payment_type?: string; use_txb?: boolean; discount_rmb?: number }) => request<{ order_uuid: string; payment_url?: string; qrcode_url?: string }>('/jellyfin/purchase', { method: 'POST', body: JSON.stringify(data) }),
    jellyfinQuickConnect: (code: string) => request<void>('/jellyfin/quick-connect', { method: 'POST', body: JSON.stringify({ code }) }),
    jellyfinUpdatePassword: (currentPassword: string, newPassword: string) => request<boolean>('/jellyfin/password', { method: 'PUT', body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }) }),
    jellyfinGetDevices: () => request<any>('/jellyfin/devices'),
    jellyfinUpdateParentalRating: (rating: number) => request<void>('/jellyfin/parental-rating', { method: 'PUT', body: JSON.stringify({ rating }) }),

    // Admin
    getConfig: () => request<AppConfig>('/admin/config'),
    updateConfig: (data: AppConfig) => request<void>('/admin/config', { method: 'PUT', body: JSON.stringify(data) }),
    adminListCombos: () => request<Combo[]>('/admin/combos'),
    createCombo: (data: Partial<Combo>) => request<Combo>('/admin/combos', { method: 'POST', body: JSON.stringify(data) }),
    updateCombo: (uuid: string, data: Partial<Combo>) => request<Combo>(`/admin/combos/${uuid}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteCombo: (uuid: string) => request<void>(`/admin/combos/${uuid}`, { method: 'DELETE' }),
    getInternalSquads: () => request<Squad[]>('/admin/squads/internal'),
    adminListUsers: (params: { search?: string; limit?: number; offset?: number } = {}) => {
        const query = new URLSearchParams()
        if (params.search) query.set('search', params.search)
        if (params.limit) query.set('limit', String(params.limit))
        if (params.offset) query.set('offset', String(params.offset))
        return request<{ users: User[]; total: number }>(`/admin/users?${query}`)
    },
    adminGetUser: (id: number) => request<{ user: User; subscription?: Subscription; jellyfin?: JellyfinAccount }>(`/admin/users/${id}`),
    adminUpdateUser: (id: number, data: Partial<User> & { subscription?: any; jellyfin?: any }) => request<void>(`/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    adminListOrders: (params: { search?: string; status?: string; service_status?: string; order_type?: string; date_from?: string; date_to?: string; limit?: number; offset?: number } = {}) => {
        const query = new URLSearchParams()
        if (params.search) query.set('search', params.search)
        if (params.status) query.set('status', params.status)
        if (params.service_status) query.set('service_status', params.service_status)
        if (params.order_type) query.set('order_type', params.order_type)
        if (params.date_from) query.set('date_from', params.date_from)
        if (params.date_to) query.set('date_to', params.date_to)
        if (params.limit) query.set('limit', String(params.limit))
        if (params.offset) query.set('offset', String(params.offset))
        return request<{ orders: OrderDetail[]; total: number }>(`/admin/orders?${query}`)
    },
    adminUpdateOrder: (uuid: string, data: Partial<Order>) => request<{ order: OrderDetail; order_events: OrderEvent[] }>(`/admin/orders/${uuid}`, { method: 'PUT', body: JSON.stringify(data) }),
    adminOrderAction: (uuid: string, action: string) => request<void>(`/admin/orders/${uuid}/actions/${action}`, { method: 'POST' }),

    // Subscription Binding
    bindSubscription: (subUrl: string) => request<{ rw_user: string }>('/bind-sub', { method: 'POST', body: JSON.stringify({ sub_url: subUrl }) }),

    // Custom Payment
    customPayment: (data: { amount: number; payment_method: string; payment_type?: string; message?: string; use_txb?: boolean; discount_rmb?: number }) => request<{ order_uuid: string; payment_url?: string; qrcode_url?: string }>('/payment/custom', { method: 'POST', body: JSON.stringify(data) }),
}
