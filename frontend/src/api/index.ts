const BASE_URL = '/api/v1'

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
    if (json.code !== 200 && resp.ok === false) {
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

    // VPN Info
    getBandwidth: () => request<any[]>('/vpn/bandwidth'),
    getDevices: () => request<any[]>('/vpn/devices'),
    getIPs: () => request<any>('/vpn/ips'),
    getSubHistory: () => request<any[]>('/vpn/history'),

    // Squads
    getExternalSquads: () => request<any[]>('/squads/external'),
    updateExternalSquad: (squadUUID: string) => request<any>('/squads/external', { method: 'PUT', body: JSON.stringify({ squad_uuid: squadUUID }) }),

    // IP Change
    changeIP: () => request<any>('/ip/change', { method: 'POST' }),
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
    createCombo: (data: any) => request<any>('/admin/combos', { method: 'POST', body: JSON.stringify(data) }),
    getInternalSquads: () => request<any[]>('/admin/squads/internal'),
}
