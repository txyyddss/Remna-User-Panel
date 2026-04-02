/**
 * Core application type definitions.
 *
 * These interfaces mirror the backend Go models and are used throughout
 * the API client, Pinia stores, and Vue views to replace bare `any` types.
 */

// ─── User ────────────────────────────────────────────────────────────
export interface User {
    id: number
    telegram_id: number
    telegram_name: string
    remnawave_uuid: string
    jellyfin_user_id: string
    credit: number
    is_admin: boolean
    created_at: string
    updated_at: string
}

// ─── Subscription ────────────────────────────────────────────────────
export interface Subscription {
    id: number
    user_id: number
    combo_uuid: string
    remnawave_uuid: string
    status: string
    expires_at: string
    created_at: string
    updated_at: string
}

export interface SubInfo {
    uuid: string
    id: number
    shortUuid: string
    username: string
    status: string
    trafficLimitBytes: number
    trafficLimitStrategy: string
    usedTrafficBytes: number
    lifetimeUsedTrafficBytes: number
    expireAt: string
    createdAt: string
    lastTrafficResetAt: string | null
    telegramId: number | null
    email: string
    description: string
    tag: string
    hwidDeviceLimit: number
    subscriptionUrl: string
    trojanPassword: string
    vlessUuid: string
    ssPassword: string
    onlineAt: string | null
    subLastUserAgent: string
    subLastOpenedAt: string | null
    subRevokedAt: string | null
    activeInternalSquads: Squad[]
    externalSquadUuid: string
    userTraffic?: {
        usedTrafficBytes: number
        lifetimeUsedTrafficBytes: number
        onlineAt: string | null
    }
}

export interface SubKeys {
    subscription_url: string
    short_uuid?: string
    vless_uuid?: string
    trojan_password?: string
    ss_password?: string
    username?: string
    instructions?: string[]
}

// ─── Combo / Plan ────────────────────────────────────────────────────
export interface Combo {
    uuid: string
    name: string
    description: string
    squad_uuid: string
    traffic_gb: number
    strategy: string
    cycle: string
    price_rmb: number
    reset_price: number
    active: boolean
    created_at?: string
}

// ─── Orders ──────────────────────────────────────────────────────────
export interface Order {
    uuid: string
    user_id: number
    order_type: string
    amount: number
    txb_discount: number
    final_amount: number
    status: string
    service_status: string
    payment_method: string
    payment_type: string
    upstream_id: string
    metadata: string
    admin_note: string
    paid_at: string | null
    created_at: string
    updated_at: string
}

export interface OrderDetail extends Order {
    user_telegram_id: number
    user_telegram_name: string
}

export interface OrderEvent {
    id: number
    order_uuid: string
    actor_user_id: number | null
    event_type: string
    message: string
    payload: string
    created_at: string
}

// ─── Credits ─────────────────────────────────────────────────────────
export interface CreditLog {
    id: number
    user_id: number
    amount: number
    balance: number
    reason: string
    created_at: string
}

// ─── Jellyfin ────────────────────────────────────────────────────────
export interface JellyfinAccount {
    id: number
    user_id: number
    jellyfin_user_id: string
    username: string
    parental_rating: number
    expires_at: string
    created_at: string
}

// ─── App Configuration ──────────────────────────────────────────────
export interface UsdtNetwork {
    value: string
    label: string
}

export interface AppConfig {
    credit_name?: string
    rmb_to_txb_rate?: number
    txb_to_rmb_rate?: number
    credit?: {
        name?: string
        txb_to_rmb_rate?: number
        rmb_to_txb_rate?: number
        signup_min?: number
        signup_max?: number
        bet_loss_multiplier?: number
        bet_win_multiplier?: number
        log_retention_days?: number
    }
    jellyfin?: {
        monthly_price_rmb?: number
    }
    payments?: {
        usdt_networks?: UsdtNetwork[]
    }
    [key: string]: unknown
}

// ─── Bandwidth / Devices / History ──────────────────────────────────
export interface BandwidthEntry {
    nodeUuid?: string
    nodeUUID?: string
    nodeName?: string
    countryCode?: string
    total?: number
}

export interface DeviceEntry {
    hwid: string
    platform?: string
    deviceModel?: string
    osVersion?: string
    userAgent?: string
}

export interface HistoryEntry {
    id: number
    userAgent?: string
    ip?: string
    createdAt: string
}

// ─── Squad ──────────────────────────────────────────────────────────
export interface Squad {
    uuid: string
    name: string
}

// ─── IP Change ──────────────────────────────────────────────────────
export interface IPChangeStatus {
    count: number
    status: 'WAITING' | 'PENDING' | 'CHANGING'
}

export interface IPChangeResponse {
    success: boolean
}

export interface AdminIPChangeRequest {
    id: number
    request_key: string
    user_id?: number
    username: string
    short_uuid: string
    reason: string
    status: 'PENDING' | 'CHANGING' | 'COMPLETED' | 'REJECTED'
    agree_count: number
    decline_count: number
    message_id: number
    message_link?: string
    requested_at: string
    completed_at?: string
    updated_at: string
}

export interface MiniAppAccessStatus {
    user?: {
        id: number
        telegram_id: number
        telegram_name: string
        is_admin: boolean
    }
    channel_joined: boolean
    group_joined: boolean
    invite_link: string
    channel_url: string
}

// ─── Jellyfin Devices ───────────────────────────────────────────────
export interface JellyfinDevice {
    Id: string
    Name: string
    AppName?: string
    AppVersion?: string
    DateLastActivity?: string
    LastUserId?: string
}

export interface JellyfinDevicesResponse {
    Items: JellyfinDevice[]
    TotalRecordCount: number
}

// ─── IP List ────────────────────────────────────────────────────────
export interface IPListResponse {
    ips?: string[]
    [key: string]: unknown
}

// ─── Payment Response ───────────────────────────────────────────────
export interface PaymentResponse {
    order_uuid: string
    final_amount: number
    txb_discount: number
    txb_used: number
    is_zero_amount: boolean
    payment_method: string
    payment_type: string
    trade_id?: string
    payment_url?: string
    qr_content?: string
    display_amount?: string
    display_currency?: string
    payment_address?: string
    network?: string
    url_scheme?: string
    expires_in_seconds?: number
}

// ─── API Responses ──────────────────────────────────────────────────
export interface ApiResponse<T = unknown> {
    success: boolean
    data?: T
    error?: string
}

export interface PaginatedResponse<T> {
    items: T[]
    total: number
}

