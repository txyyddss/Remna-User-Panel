let BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'
if (BASE_URL !== '/api/v1' && !BASE_URL.endsWith('/api/v1')) {
    BASE_URL = BASE_URL.replace(/\/$/, '') + '/api/v1'
}

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
    getMe: () => request<any>('/user/me'),

    // Credit
    getCreditBalance: () => request<{ balance: number; name: string }>('/credit/balance'),
    creditSignup: () => request<{ value: number; new_balance: number; auto_delete: boolean }>('/credit/signup', { method: 'POST' }),
    creditBet: (amount: number) => request<{ result: number; new_balance: number }>('/credit/bet', { method: 'POST', body: JSON.stringify({ amount }) }),
    getCreditHistory: (limit = 20, offset = 0) => request<any[]>(`/credit/history?limit=${limit}&offset=${offset}`),

    // Combos
    listCombos: () => request<any[]>('/combos'),

    // Subscription
    purchaseCombo: (data: any) => request<any>('/subscribe', { method: 'POST', body: JSON.stringify(data) }),
    getSubInfo: () => request<any>('/sub-info'),
    getSubKeys: () => request<any>('/sub-keys'),

    // Payment
    createPayment: (data: any) => request<any>('/payment/create', { method: 'POST', body: JSON.stringify(data) }),
    getOrders: (limit = 20, offset = 0) => request<any[]>(`/orders?limit=${limit}&offset=${offset}`),
    getOrder: (uuid: string) => request<any>(`/orders/${uuid}`),

    // VPN Info
    getBandwidth: () => request<any[]>('/vpn/bandwidth'),
    getDevices: () => request<any[]>('/vpn/devices'),
    getIPs: () => request<any>('/vpn/ips'),
    getSubHistory: () => request<any[]>('/vpn/history'),

    // Squads
    getExternalSquads: () => request<any[]>('/squads/external'),
    updateExternalSquad: (squadUUID: string) => request<any>('/squads/external', { method: 'PUT', body: JSON.stringify({ squad_uuid: squadUUID }) }),

    // IP Change
    changeIP: (data: any = {}) => request<any>('/ip/change', { method: 'POST', body: JSON.stringify(data) }),
    getIPStatus: () => request<any>('/ip/status'),

    // Jellyfin
    purchaseJellyfin: (data: any) => request<any>('/jellyfin/purchase', { method: 'POST', body: JSON.stringify(data) }),
    jellyfinQuickConnect: (code: string) => request<any>('/jellyfin/quick-connect', { method: 'POST', body: JSON.stringify({ code }) }),
    jellyfinUpdatePassword: (currentPassword: string, newPassword: string) => request<any>('/jellyfin/password', { method: 'PUT', body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }) }),
    jellyfinGetDevices: () => request<any>('/jellyfin/devices'),
    jellyfinUpdateParentalRating: (rating: number) => request<any>('/jellyfin/parental-rating', { method: 'PUT', body: JSON.stringify({ rating }) }),

    // Admin
    getConfig: () => request<any>('/admin/config'),
    updateConfig: (data: any) => request<any>('/admin/config', { method: 'PUT', body: JSON.stringify(data) }),
    adminListCombos: () => request<any[]>('/admin/combos'),
    createCombo: (data: any) => request<any>('/admin/combos', { method: 'POST', body: JSON.stringify(data) }),
    updateCombo: (uuid: string, data: any) => request<any>(`/admin/combos/${uuid}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteCombo: (uuid: string) => request<any>(`/admin/combos/${uuid}`, { method: 'DELETE' }),
    getInternalSquads: () => request<any[]>('/admin/squads/internal'),
    adminListUsers: (params: { search?: string; limit?: number; offset?: number } = {}) => {
        const query = new URLSearchParams()
        if (params.search) query.set('search', params.search)
        if (params.limit) query.set('limit', String(params.limit))
        if (params.offset) query.set('offset', String(params.offset))
        return request<any>(`/admin/users?${query}`)
    },
    adminGetUser: (id: number) => request<any>(`/admin/users/${id}`),
    adminUpdateUser: (id: number, data: any) => request<any>(`/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
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
        return request<any>(`/admin/orders?${query}`)
    },
    adminUpdateOrder: (uuid: string, data: any) => request<any>(`/admin/orders/${uuid}`, { method: 'PUT', body: JSON.stringify(data) }),
    adminOrderAction: (uuid: string, action: string) => request<any>(`/admin/orders/${uuid}/actions/${action}`, { method: 'POST' }),

    // Subscription Binding
    bindSubscription: (subUrl: string) => request<any>('/bind-sub', { method: 'POST', body: JSON.stringify({ sub_url: subUrl }) }),

    // Custom Payment
    customPayment: (data: any) => request<any>('/payment/custom', { method: 'POST', body: JSON.stringify(data) }),
}
