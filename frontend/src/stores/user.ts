import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api'
import type { User, Subscription, JellyfinAccount, AppConfig, SubInfo, SubKeys, Order } from '@/types'

export const useUserStore = defineStore('user', () => {
    const user = ref<User | null>(null)
    const subscription = ref<Subscription | null>(null)
    const jellyfin = ref<JellyfinAccount | null>(null)
    const appConfig = ref<AppConfig | null>(null)
    const liveSubInfo = ref<{ has_subscription?: boolean; user?: SubInfo } | null>(null)
    const subKeys = ref<SubKeys | null>(null)
    const recentOrders = ref<Order[]>([])
    const loading = ref(false)
    const refreshing = ref(false)
    const error = ref<string | null>(null)
    let refreshTimer: number | null = null

    const isAdmin = computed(() => user.value?.is_admin || false)
    const hasSubscription = computed(() => !!subscription.value)
    const hasJellyfin = computed(() => !!jellyfin.value)
    const credit = computed(() => user.value?.credit || 0)
    const telegramName = computed(() => user.value?.telegram_name || 'User')
    const currentExternalSquadUUID = computed(() => liveSubInfo.value?.user?.externalSquadUuid || '')

    async function fetchMe(background = false) {
        if (!background) {
            loading.value = true
        }
        error.value = null
        try {
            const data = await api.getMe()
            user.value = data.user
            subscription.value = data.subscription || null
            jellyfin.value = data.jellyfin || null
            appConfig.value = data.config || null
        } catch (e: any) {
            error.value = e.message
        } finally {
            if (!background) {
                loading.value = false
            }
        }
    }

    async function refreshCredit() {
        try {
            const data = await api.getCreditBalance()
            if (user.value) {
                user.value.credit = data.balance
            }
        } catch (e) { }
    }

    async function refreshSubInfo() {
        if (!user.value?.remnawave_uuid && !subscription.value) {
            liveSubInfo.value = null
            subKeys.value = null
            return
        }
        try {
            liveSubInfo.value = await api.getSubInfo()
            if (liveSubInfo.value?.has_subscription) {
                subKeys.value = await api.getSubKeys()
            } else {
                subKeys.value = null
            }
        } catch (e) {
            liveSubInfo.value = null
            subKeys.value = null
        }
    }

    async function refreshOrders(limit = 10) {
        try {
            recentOrders.value = (await api.getOrders(limit, 0)) || []
        } catch (e) {
            recentOrders.value = []
        }
    }

    async function refreshState(options: { background?: boolean; ordersLimit?: number } = {}) {
        if (refreshing.value) {
            return
        }
        refreshing.value = true
        try {
            await fetchMe(!!options.background)
            await Promise.all([
                refreshSubInfo(),
                refreshOrders(options.ordersLimit ?? 10),
            ])
        } finally {
            refreshing.value = false
        }
    }

    function startAutoRefresh(intervalMs = 15000) {
        stopAutoRefresh()
        refreshTimer = window.setInterval(() => {
            refreshState({ background: true, ordersLimit: 10 })
        }, intervalMs)
    }

    function stopAutoRefresh() {
        if (refreshTimer !== null) {
            window.clearInterval(refreshTimer)
            refreshTimer = null
        }
    }

    return {
        user, subscription, jellyfin, appConfig, liveSubInfo, subKeys, recentOrders, loading, refreshing, error,
        isAdmin, hasSubscription, hasJellyfin, credit, telegramName, currentExternalSquadUUID,
        fetchMe, refreshCredit, refreshSubInfo, refreshOrders, refreshState, startAutoRefresh, stopAutoRefresh,
    }
})
